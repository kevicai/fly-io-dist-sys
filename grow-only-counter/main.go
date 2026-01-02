package main

import (
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	n.Handle("init", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{"type": "init_ok"})
	})
}
