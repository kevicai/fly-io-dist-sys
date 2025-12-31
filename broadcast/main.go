package main

import (
	"encoding/json"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type NodeState struct {
	nodeID      string
	neighbors   []string
	messages    []int
}

func main() {
	n := maelstrom.NewNode()
	state := &NodeState{}

	// Handle init message to extract node ID
	n.Handle("init", func(msg maelstrom.Message) error {
		var body struct {
			NodeID string `json:"node_id"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		state.nodeID = body.NodeID

		return n.Reply(msg, map[string]any{
			"type": "init_ok",
		})
	})

	// Handle topology message to extract neighbors
	n.Handle("topology", func(msg maelstrom.Message) error {
		var body struct {
			Topology map[string][]string `json:"topology"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		if neighbors, ok := body.Topology[state.nodeID]; !ok {
			return fmt.Errorf("Node ID not found in topology")
		} else {
			state.neighbors = neighbors
		}

		return n.Reply(msg, map[string]any{
			"type": "topology_ok",
		})
	})

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body struct {
			Message int `json:"message"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		state.messages = append(state.messages, body.Message)

		for _, neighbor := range state.neighbors {
			n.Send(neighbor, map[string]any{
				"type":    "broadcast",
				"message": body.Message,
			})
		}

		return n.Reply(msg, map[string]any{
			"type": "broadcast_ok",
		})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type": "read_ok",
			"messages": state.messages,
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

