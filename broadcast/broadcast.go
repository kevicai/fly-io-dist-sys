package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

// createBroadcastFunc creates a function that broadcasts a message to a neighbor
// and handles acknowledgments.
func createBroadcastFunc(n *maelstrom.Node, state *NodeState) func(string, int) {
	return func(neighbor string, message int) {
		err := n.RPC(neighbor, map[string]any{
			"type":    "broadcast",
			"message": message,
		}, func(msg maelstrom.Message) error {
			// Handle broadcast_ok response
			var body struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(msg.Body, &body); err != nil {
				log.Printf("failed to unmarshal RPC response: %v", err)
				return err
			}
			if body.Type == "broadcast_ok" {
				state.MarkAcked(message, neighbor)
			}
			return nil
		})

		// Log RPC errors but don't fail - gossip will retry
		if err != nil {
			log.Printf("RPC to %s failed: %v", neighbor, err)
		}
	}
}
