package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

// Record stores the value and the time it was written to handle conflicts (LWW)
type Record struct {
	Val     any   `json:"v"`
	Version int64 `json:"t"`
}

type TxnServer struct {
	n     *maelstrom.Node
	mu    sync.RWMutex
	store map[any]Record
}

func main() {
	n := maelstrom.NewNode()
	s := &TxnServer{
		n:     n,
		store: make(map[any]Record),
	}

	n.Handle("txn", s.handleTxn)
	n.Handle("gossip", s.handleGossip)

	// Start background gossip to ensure eventual consistency
	go s.gossipLoop()

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func (s *TxnServer) handleTxn(msg maelstrom.Message) error {
	var body struct {
		Type string          `json:"type"`
		Txn  [][]interface{} `json:"txn"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	// We lock the whole transaction to ensure "Read Committed" isolation locally
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixNano()

	for _, op := range body.Txn {
		opType := op[0].(string)
		key := op[1]

		switch opType {
		case "r":
			if record, ok := s.store[key]; ok {
				op[2] = record.Val
			} else {
				op[2] = nil
			}
		case "w":
			s.store[key] = Record{
				Val:     op[2],
				Version: now,
			}
		}
	}

	return s.n.Reply(msg, map[string]any{
		"type": "txn_ok",
		"txn":  body.Txn,
	})
}

func (s *TxnServer) gossipLoop() {
	ticker := time.NewTicker(300 * time.Millisecond)
	for range ticker.C {
		s.mu.RLock()
		// Create a snapshot of the current store to send
		snapshot := s.store
		s.mu.RUnlock()

		if len(snapshot) == 0 {
			continue
		}

		// Send gossip to all other nodes
		for _, dest := range s.n.NodeIDs() {
			if dest == s.n.ID() {
				continue
			}
			s.n.Send(dest, map[string]any{
				"type":  "gossip",
				"store": snapshot,
			})
		}
	}
}

func (s *TxnServer) handleGossip(msg maelstrom.Message) error {
	var body struct {
		Store map[string]Record `json:"store"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Merge logic: Last-Write-Wins (LWW)
	for key, incoming := range body.Store {
		local, exists := s.store[key]
		if !exists || incoming.Version > local.Version {
			s.store[key] = incoming
		}
	}
	return nil
}
