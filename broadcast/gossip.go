package main

import (
	"context"
	"time"
)

const (
	// gossipInterval is how often to retry unacked messages
	gossipInterval = 5 * time.Second
)

// startGossipLoop periodically retries unacked messages to handle network faults.
func startGossipLoop(ctx context.Context, state *NodeState, broadcast func(string, int)) {
	go func() {
		ticker := time.NewTicker(gossipInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Handle context cancellation for graceful shutdown
				return
			case <-ticker.C:
				// Get unacked messages without holding lock
				unacked := state.GetUnackedCopy()

				// Broadcast unacked messages
				for message, neighbors := range unacked {
					for _, neighbor := range neighbors {
						go broadcast(neighbor, message)
					}
				}
			}
		}
	}()
}
