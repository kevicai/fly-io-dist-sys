package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type IDState struct {
	mu          sync.Mutex
	lastEpoch   int64
	counter     int64
	nodeID      int64
}

func main() {
	n := maelstrom.NewNode()
	state := &IDState{}

	// Handle init message to extract node ID
	n.Handle("init", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Extract numeric part from nodeID (e.g., "n1" -> 1)
		nodeIDStr, ok := body["node_id"].(string)
		if !ok {
			log.Fatal("Invalid node_id in init message")
		}

		var nodeIDNum int64
		if _, err := fmt.Sscanf(nodeIDStr, "n%d", &nodeIDNum); err != nil {
			log.Fatal("Invalid node ID doesn't match the format 'n<number>'")
		}

		state.nodeID = nodeIDNum

		body["type"] = "init_ok"
		return n.Reply(msg, body)
	})

	n.Handle("generate", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		id := generateID(state)

		body["type"] = "generate_ok"
		body["id"] = id.Encode()
		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func generateID(state *IDState) *ID {
	state.mu.Lock()
	defer state.mu.Unlock()

	currentEpoch := time.Now().Unix()

	if currentEpoch > state.lastEpoch {
		state.lastEpoch = currentEpoch
		state.counter = 0
	} else {
		state.counter++
	}

	return NewID(currentEpoch, state.nodeID, state.counter)
}
