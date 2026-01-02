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


Features:
- **Deduplication**: Messages are stored in a set to prevent duplicates
- **Gossip protocol**: Periodic retry mechanism for network fault tolerance
- **Acknowledgment tracking**: Tracks which neighbors have received each message
- **Thread-safe**: All state operations are protected with mutexes
- **Graceful shutdown**: Context-based cancellation for clean termination

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

