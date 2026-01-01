package main

import "sync"

// NodeState holds the node's state including messages and neighbor acknowledgments
type NodeState struct {
	mu        sync.Mutex
	neighbors []string
	messages  map[int]struct{} // Set of seen messages
	// Map of message to the set of neighbors that have not acknowledged the message
	messageToUnackedNeighbors map[int]map[string]struct{}
}

// NewNodeState creates a new NodeState instance
func NewNodeState() *NodeState {
	return &NodeState{
		messages:                  make(map[int]struct{}),
		messageToUnackedNeighbors: make(map[int]map[string]struct{}),
	}
}

// SetNeighbors updates the list of neighbors (thread-safe)
func (s *NodeState) SetNeighbors(neighbors []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.neighbors = neighbors
}

// GetNeighborsCopy returns a copy of the neighbors list (thread-safe)
func (s *NodeState) GetNeighborsCopy() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	neighbors := make([]string, len(s.neighbors))
	copy(neighbors, s.neighbors)
	return neighbors
}

// AddMessage adds a message to the seen set and returns whether it was new
func (s *NodeState) AddMessage(message int) (isNew bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, seen := s.messages[message]
	s.messages[message] = struct{}{}
	return !seen
}

// GetMessages returns a copy of all messages (thread-safe)
func (s *NodeState) GetMessages() []int {
	s.mu.Lock()
	defer s.mu.Unlock()
	messages := make([]int, 0, len(s.messages))
	for msg := range s.messages {
		messages = append(messages, msg)
	}
	return messages
}

// TrackUnacked marks all neighbors as unacknowledged for a message
func (s *NodeState) TrackUnacked(message int, neighbors []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messageToUnackedNeighbors[message] = make(map[string]struct{})
	for _, neighbor := range neighbors {
		s.messageToUnackedNeighbors[message][neighbor] = struct{}{}
	}
}

// MarkAcked removes a neighbor from the unacked set for a message
func (s *NodeState) MarkAcked(message int, neighbor string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.messageToUnackedNeighbors[message], neighbor)
	if len(s.messageToUnackedNeighbors[message]) == 0 {
		delete(s.messageToUnackedNeighbors, message)
	}
}

// GetUnackedCopy returns a copy of unacked messages (thread-safe)
func (s *NodeState) GetUnackedCopy() map[int][]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	unacked := make(map[int][]string)
	for message, neighbors := range s.messageToUnackedNeighbors {
		if len(neighbors) > 0 {
			neighborList := make([]string, 0, len(neighbors))
			for neighbor := range neighbors {
				neighborList = append(neighborList, neighbor)
			}
			unacked[message] = neighborList
		}
	}
	return unacked
}
