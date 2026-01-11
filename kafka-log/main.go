package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type LogServer struct {
	n  *maelstrom.Node
	kv *maelstrom.KV

	mu sync.Mutex
}

func main() {
	n := maelstrom.NewNode()
	s := &LogServer{
		n:  n,
		kv: maelstrom.NewLinKV(n),
	}

	n.Handle("send", s.handleSend)
	n.Handle("poll", s.handlePoll)
	n.Handle("commit_offsets", s.handleCommitOffsets)
	n.Handle("list_committed_offsets", s.handleListCommittedOffsets)

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func (s *LogServer) handleSend(msg maelstrom.Message) error {
	var body struct {
		Key string `json:"key"`
		Msg int    `json:"msg"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	ctx := context.Background()
	offsetKey := body.Key + "_offset_counter"

	// Read current offset
	currOffset, err := s.kv.ReadInt(ctx, offsetKey)
	if err != nil {
		if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
			currOffset = 0
		} else {
			return err
		}
	}

	// Retry loop to get an offset (ticket system)
	for {
		// CAS to exactly curr + 1
		err = s.kv.CompareAndSwap(ctx, offsetKey, currOffset, currOffset+1, true)
		if err == nil {
			// Success: We got the offset
			break
		} else if maelstrom.ErrorCode(err) == maelstrom.PreconditionFailed {
			// Another node got curr+1. We loop to try for curr+2.
			currOffset++
			continue
		} else {
			return err
		}
	}

	// Write message to the offset
	msgKey := fmt.Sprintf("%s_msg_%d", body.Key, currOffset)
	if err := s.kv.Write(ctx, msgKey, body.Msg); err != nil {
		return err
	}
	return s.n.Reply(msg, map[string]any{"type": "send_ok", "offset": currOffset})
}

func (s *LogServer) handlePoll(msg maelstrom.Message) error {
	var body struct {
		Offsets map[string]int `json:"offsets"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res := make(map[string][][2]int)

	// Poll each topic concurrently.
	var wg sync.WaitGroup
	var mu sync.Mutex
	const maxMsgs = 20

	for topic, start := range body.Offsets {
		wg.Add(1)
		go func(t string, sOff int) {
			// Get messages from the topic starting at the start offset
			defer wg.Done()
			var msgs [][2]int

			// Stop at the first error since there's sequence guarantee
			for i := sOff; ; i++ {
				val, err := s.kv.ReadInt(ctx, fmt.Sprintf("%s_msg_%d", t, i))
				if err != nil {
					break
				}
				msgs = append(msgs, [2]int{i, val})
				if len(msgs) >= maxMsgs {
					break
				}
			}

			if len(msgs) > 0 {
				mu.Lock()
				res[t] = msgs
				mu.Unlock()
			}
		}(topic, start)
	}
	wg.Wait()

	return s.n.Reply(msg, map[string]any{"type": "poll_ok", "msgs": res})
}

func (s *LogServer) commitKey(topic string) string {
	return fmt.Sprintf("%s_committed_offset", topic)
}

func (s *LogServer) handleCommitOffsets(msg maelstrom.Message) error {
	var body struct {
		Offsets map[string]int `json:"offsets"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	ctx := context.Background()
	for topic, offset := range body.Offsets {
		// Simply overwrite the latest committed offset for this topic
		s.kv.Write(ctx, s.commitKey(topic), offset)
	}
	return s.nodeReply(msg, map[string]any{"type": "commit_offsets_ok"})
}

func (s *LogServer) handleListCommittedOffsets(msg maelstrom.Message) error {
	var body struct {
		Keys []string `json:"keys"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	ctx := context.Background()
	res := make(map[string]int)
	for _, topic := range body.Keys {
		val, err := s.kv.ReadInt(ctx, s.commitKey(topic))
		if err == nil {
			res[topic] = val
		}
	}
	return s.nodeReply(msg, map[string]any{"type": "list_committed_offsets_ok", "offsets": res})
}

func (s *LogServer) nodeReply(msg maelstrom.Message, body any) error {
	return s.n.Reply(msg, body)
}
