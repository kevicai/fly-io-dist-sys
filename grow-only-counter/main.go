package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

// CounterServer encapsulates dependencies and state
type CounterServer struct {
	node *maelstrom.Node
	kv   *maelstrom.KV
}

func main() {
	n := maelstrom.NewNode()
	s := &CounterServer{
		node: n,
		kv:   maelstrom.NewSeqKV(n),
	}

	n.Handle("add", s.handleAdd)
	n.Handle("read", s.handleRead)

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

// keyFor returns the consistent key name for a given node
func (s *CounterServer) keyFor(nodeID string) string {
	return "counter_" + nodeID
}

func (s *CounterServer) handleAdd(msg maelstrom.Message) error {
	var body struct {
		Delta int `json:"delta"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	// Use a request-scoped context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	key := s.keyFor(s.node.ID())

	// Retry Loop
	for {
		// Check if context is already done (e.g., timeout)
		if err := ctx.Err(); err != nil {
			return err
		}

		current, err := s.kv.ReadInt(ctx, key)
		if err != nil {
			if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
				current = 0
			}
		}

		err = s.kv.CompareAndSwap(ctx, key, current, current+body.Delta, true)
		if err == nil {
			break // Success
		}

		if maelstrom.ErrorCode(err) != maelstrom.PreconditionFailed {
			return err // Unexpected error
		}

		// Another go routine handler already updated the value,
		// or transient network error, retry
		time.Sleep(100 * time.Millisecond)
		continue
	}

	return s.node.Reply(msg, map[string]any{"type": "add_ok"})
}

// Eventually consistent read
func (s *CounterServer) handleRead(msg maelstrom.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nodeIDs := s.node.NodeIDs()
	results := make(chan int, len(nodeIDs))

	// Fire off all reads concurrently
	for _, id := range nodeIDs {
		go func(nodeID string) {
			for {
				val, err := s.kv.ReadInt(ctx, s.keyFor(nodeID))
				if err != nil {
					// retry read
					time.Sleep(100 * time.Millisecond)
					continue
				}
				results <- val
				break
			}
		}(id)
	}

	total := 0
	for i := 0; i < len(nodeIDs); i++ {
		total += <-results // Wait for all reads to complete
	}

	return s.node.Reply(msg, map[string]any{
		"type":  "read_ok",
		"value": total,
	})
}
