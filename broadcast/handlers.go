package main

import (
	"encoding/json"
	"fmt"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

// handleTopology processes topology messages and updates neighbors.
func handleTopology(n *maelstrom.Node, state *NodeState) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		var body struct {
			Topology map[string][]string `json:"topology"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return fmt.Errorf("failed to unmarshal topology: %w", err)
		}

		neighbors, ok := body.Topology[n.ID()]
		if !ok {
			return fmt.Errorf("node ID %q not found in topology", n.ID())
		}

		state.SetNeighbors(neighbors)

		return n.Reply(msg, map[string]any{
			"type": "topology_ok",
		})
	}
}

// handleBroadcast processes broadcast messages and propagates them to neighbors.
func handleBroadcast(n *maelstrom.Node, state *NodeState, broadcast func(string, int)) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		var body struct {
			Message int `json:"message"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return fmt.Errorf("failed to unmarshal broadcast: %w", err)
		}

		isNew := state.AddMessage(body.Message)
		neighbors := state.GetNeighborsCopy()

		// Track all neighbors as unacked for this message
		state.TrackUnacked(body.Message, neighbors)

		// Only broadcast if this is a new message
		if isNew {
			for _, neighbor := range neighbors {
				go broadcast(neighbor, body.Message)
			}
		}

		return n.Reply(msg, map[string]any{
			"type": "broadcast_ok",
		})
	}
}

// handleRead returns all messages seen by this node.
func handleRead(n *maelstrom.Node, state *NodeState) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		messages := state.GetMessages()

		return n.Reply(msg, map[string]any{
			"type":     "read_ok",
			"messages": messages,
		})
	}
}
