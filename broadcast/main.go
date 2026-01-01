package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	n := maelstrom.NewNode()
	state := NewNodeState()

	// Create broadcast function
	broadcast := createBroadcastFunc(n, state)

	// Start periodic gossip loop for retry logic
	startGossipLoop(ctx, state, broadcast)

	// Register message handlers
	n.Handle("topology", handleTopology(n, state))
	n.Handle("broadcast", handleBroadcast(n, state, broadcast))
	n.Handle("read", handleRead(n, state))

	// Run the node
	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
