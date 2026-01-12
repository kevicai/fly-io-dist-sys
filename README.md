# Fly.io Distributed Systems Challenge

Solutions for the [Fly.io Distributed Systems Challenge](https://fly.io/dist-sys/), implementing distributed systems algorithms using Maelstrom.

## Prerequisites

- Go 1.22+
- OpenJDK (for Maelstrom)
- Graphviz & gnuplot (for Maelstrom visualization)

## Solutions

### Echo

The "Hello World" of distributed systems.

Key Concept: RPC Lifecycle. Demonstrates the basic Request-Response pattern and JSON serialization/deserialization over stdin/stdout. It ensures the node can correctly initialize and join the Maelstrom network.

Build and test:
```bash
cd echo
go install .
./maelstrom/maelstrom test -w echo --bin ~/go/bin/maelstrom-echo --node-count 1 --time-limit 10
```

### Unique ID Generation

A high-performance ID generator.

Key Concept: Coordination-Free Consensus. To maintain "Total Availability," this avoids a central sequencer. By using Snowflake ID (Node ID, timestamp, and timestamp tie breaker counter), we guarantee global uniqueness without any network cross-talk or locks.

Build and test:
```bash
cd unique-ids
go install .
./maelstrom/maelstrom test -w unique-ids --bin ~/go/bin/maelstrom-unique-ids --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition
```

### Broadcast

A fault-tolerant broadcast system that propagates messages to all nodes in a cluster.
Key Concepts: 
- Push-based Gossip: Immediately sends new messages to neighbors.
- Periodic Reconciliation: Nodes exchange summaries of seen messages to ensure that even if a packet is dropped, the data eventually reaches every node.

Build and test:
```bash
cd broadcast
go install .
# Benchmark: 25 nodes, 100 msg/s, 100ms latency
./maelstrom/maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100
```
Results:
- **Median latency**: 461ms
- **95th percentile**: 703ms
- **99th percentile**: 785ms
- **Max latency**: 853ms

### Grow-Only Counter

A CRDT-inspired (Conflict-free Replicated Data Type) counter.
Key Concepts: 
- Sharded State & Commutativity. Since counter increments are commutative (order doesn't matter), we shard the state by node.
- Write Path: A node only ever modifies its own "slot" in the KV store using Compare-And-Swap (CAS) to avoid lost updates and shared hot key.
- Read Path: We perform a Parallel Scatter-Gather to sum all shards, providing an eventually consistent global total.

Build and test:
```bash
cd grow-only-counter
go install .
./maelstrom/maelstrom test -w g-counter --bin ~/go/bin/maelstrom-grow-only-counter --node-count 3 --rate 100 --time-limit 20 --nemesis partition
```

### Kafka-Style Log

A partitioned, append-only linearizable log similiar to Apache Kafka.

Key Concepts: 
- Atomic Sequencers. To ensure strict ordering per partition, nodes must agree on a "Ticket".
- Linearizable Offsets: Uses a lin-kv store to perform a Compare-And-Swap on a counter key. This acts as a global lock ticketing system, ensuring no two messages ever share the same offset.
- Offset-based Addressing: Once a ticket is claimed, the message is stored at a deterministic key (topic_offset), allowing O(1) random access reads.

Build and test:
```bash
cd kafka-log
go install .
./maelstrom/maelstrom test -w kafka --bin ~/go/bin/maelstrom-kafka --node-count 2 --concurrency 2n --time-limit 20 --rate 1000
```

### Total Available Transaction KV Store

A totally available, multi-master, eventually consistent key-value store.
Key Concepts: 
- AP over CP (CAP Theorem). This Partition Tolerance and Availability and over consistency.
- Read Committed Isolation: Prevents "Dirty Reads" by using local Mutual Exclusion (Mutexes) during the transaction loop so that intermediate states are never visible to other readers.
- Last-Write-Wins (LWW): Uses high-resolution timestamps to resolve conflicts during gossip, ensuring the cluster eventually converges to the same state after a network partition (eventually consistent).

```bash
cd totally-available-transactions
go install .
./maelstrom test -w txn-rw-register --bin ~/go/bin/maelstrom-txn --node-count 2 --concurrency 2n --time-limit 20 --rate 1000 --consistency-models read-committed --availability total –-nemesis partition
```