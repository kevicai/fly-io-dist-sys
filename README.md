# Fly.io Distributed Systems Challenge

Solutions for the [Fly.io Distributed Systems Challenge](https://fly.io/dist-sys/1/), implementing distributed systems algorithms using Maelstrom.

## Prerequisites

- Go 1.25+
- OpenJDK (for Maelstrom)
- Graphviz & gnuplot (for Maelstrom visualization)

## Running Tests

### Echo Challenge

Build the echo binary:
```bash
cd echo
go get github.com/jepsen-io/maelstrom/demo/go
go install .
```

Run the test:
```bash
./maelstrom/maelstrom test -w echo --bin ~/go/bin/maelstrom-echo --node-count 1 --time-limit 10
```

## Debugging

To view detailed test results in a web interface:
```bash
./maelstrom/maelstrom serve
```

Then open http://localhost:8080 in your browser.

## Challenges

- [x] Challenge #1: Echo
- [ ] Challenge #2: Unique ID Generation
- [ ] Challenge #3: Broadcast
- [ ] Challenge #4: Grow-Only Counter
- [ ] Challenge #5: Kafka-Style Log
- [ ] Challenge #6: Totally-Available Transactions

