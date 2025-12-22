package main

// Bit widths for encoding: 
const BITS_COUNTER = 12  // Allows 4096 IDs per second
const BITS_NODE_ID = 10  // Allows 1024 nodes
const BITS_EPOCH = 42    // Allows ~139 years of IDs

const SHIFT_NODE_ID = BITS_COUNTER        		 // 12 bits
const SHIFT_EPOCH = BITS_COUNTER + BITS_NODE_ID  // 22 bits

// ID represents a unique identifier with epoch, nodeID, and counter components
type ID struct {
	Epoch   int64
	NodeID  int64
	Counter int64
}

// NewID creates a new ID with the given components
func NewID(epoch int64, nodeID int64, counter int64) *ID {
	return &ID{
		Epoch:   epoch,
		NodeID:  nodeID,
		Counter: counter,
	}
}

// Encodes the ID into an int64
func (id *ID) Encode() int64 {
	encodedEpoch := id.Epoch << SHIFT_EPOCH
	encodedNodeID := id.NodeID << SHIFT_NODE_ID
	encodedCounter := id.Counter

	return encodedEpoch | encodedNodeID | encodedCounter
}

// Decodes an int64 back into an ID object
func Decode(encoded int64) *ID {
	epoch := encoded >> SHIFT_EPOCH
	nodeID := (encoded >> SHIFT_NODE_ID) & ((1 << BITS_NODE_ID) - 1)
	counter := encoded & ((1 << BITS_COUNTER) - 1)

	return &ID{
		Epoch:   epoch,
		NodeID:  nodeID,
		Counter: counter,
	}
}

