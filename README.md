# Fly.io Distributed Systems Challenge

Solutions for the [Fly.io Distributed Systems Challenge](https://fly.io/dist-sys/), implementing distributed systems algorithms using Maelstrom.

## Prerequisites

- Go 1.22+
- OpenJDK (for Maelstrom)
- Graphviz & gnuplot (for Maelstrom visualization)

## Solutions

### Echo

Build and test:
```bash
cd echo
go install .
./maelstrom/maelstrom test -w echo --bin ~/go/bin/maelstrom-echo --node-count 1 --time-limit 10
```

### Unique ID Generation

Build and test:
```bash
cd unique-ids
go install .
./maelstrom/maelstrom test -w unique-ids --bin ~/go/bin/maelstrom-unique-ids --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition
```

### Broadcast

A fault-tolerant broadcast system that propagates messages to all nodes in a cluster.
- Immediate send with periodic gossip mechanism retries unacknowledged messages for network fault tolerance

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

A distributed counter implementation using a partitioned approach where each node maintains its own counter value.
- Each node maintains its own counter in the KV store
- Add operations use CAS with retry loop for local consistency
- Read operations sum all node counters in parallel using goroutines

Build and test:
```bash
cd grow-only-counter
go install .
./maelstrom/maelstrom test -w g-counter --bin ~/go/bin/maelstrom-grow-only-counter --node-count 3 --rate 100 --time-limit 20 --nemesis partition
```

### Kafka-Style Log

A distributed append-only log similar to Apache Kafka, providing linearizable message ordering per partition.
- Ticket-based ordering system using Linearizable KV store for log offset allocation

Build and test:
```bash
cd kafka-log
go install .
./maelstrom/maelstrom test -w kafka --bin ~/go/bin/maelstrom-kafka --node-count 2 --concurrency 2n --time-limit 20 --rate 1000
```