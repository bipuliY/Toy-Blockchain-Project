# Toy Blockchain Project — Phase 2
## Networked Multi-Node Blockchain

A distributed multi-node blockchain implementation written in Go.

This project extends the Phase 1 single-process blockchain into a
networked blockchain system where multiple independent nodes communicate,
exchange transactions and blocks, synchronize their chains, and resolve
temporary forks.

The main focus of Phase 2 is introducing:

- Peer-to-peer node communication
- Transaction propagation
- Block propagation
- Blockchain synchronization
- Fork detection and resolution
- Concurrent node operation
- Multi-node testing

---

# 1. Project Evolution

## Phase 1 — Single Process Blockchain

The first phase implemented the fundamental blockchain concepts:

- Block structure
- SHA-256 hashing
- Proof-of-work mining
- Transaction handling
- Merkle tree calculation
- Blockchain validation
- Local chain management


## Phase 2 — Networked Multi-Node Blockchain

Phase 2 extends the blockchain into a distributed system.

Multiple blockchain nodes can now run independently and communicate
with each other.

New capabilities:

- Multiple blockchain nodes
- Peer management
- Transaction gossip
- Block gossip
- Chain synchronization
- Fork detection
- Chain reorganization
- Orphan transaction recovery
- Concurrent state protection
- Three-node cluster execution

---

# 2. Project Objective

The objective of Phase 2 is to demonstrate how a blockchain changes from
a single local application into a distributed network.

Each node maintains its own blockchain copy.

Nodes communicate with peers to achieve eventual agreement on the
network state.

The project demonstrates:
