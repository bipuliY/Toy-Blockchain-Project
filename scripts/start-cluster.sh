#!/bin/bash

set -e

mkdir -p logs

echo "Starting 3-node blockchain cluster..."

go run ./cmd/node \
  -address localhost:8001 \
  -peers http://localhost:8002,http://localhost:8003 \
  > logs/node1.log 2>&1 &

NODE1_PID=$!

go run ./cmd/node \
  -address localhost:8002 \
  -peers http://localhost:8001,http://localhost:8003 \
  > logs/node2.log 2>&1 &

NODE2_PID=$!

go run ./cmd/node \
  -address localhost:8003 \
  -peers http://localhost:8001,http://localhost:8002 \
  > logs/node3.log 2>&1 &

NODE3_PID=$!

cleanup() {
  echo
  echo "Stopping blockchain cluster..."

  kill "$NODE1_PID" 2>/dev/null || true
  kill "$NODE2_PID" 2>/dev/null || true
  kill "$NODE3_PID" 2>/dev/null || true

  wait 2>/dev/null || true
}

trap cleanup EXIT INT TERM

echo
echo "Node 1: http://localhost:8001"
echo "Node 2: http://localhost:8002"
echo "Node 3: http://localhost:8003"

echo
echo "Logs:"
echo "  logs/node1.log"
echo "  logs/node2.log"
echo "  logs/node3.log"

echo
echo "Press Ctrl+C to stop the cluster."

wait