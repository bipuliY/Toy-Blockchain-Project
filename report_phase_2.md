# Phase 2 Development Report
# Networked Multi-Node Blockchain

**Project:** Toy Blockchain Project  
**Phase:** Phase 2 – Networked Multi-Node Blockchain  
**Language:** Go  
**Communication:** HTTP  
**Development Focus:** Networking, Synchronization, Fork Resolution and Concurrency

---

# 1. Introduction

Phase 2 extends the blockchain implementation developed during Phase 1
from a single-process blockchain into a networked multi-node blockchain.

In Phase 1, the blockchain operated locally within one process. The main
blockchain concepts such as blocks, transactions, hashing, Merkle roots,
proof-of-work mining and chain validation were implemented.

Phase 2 introduces multiple independent blockchain nodes. Each node
maintains its own local copy of the blockchain and communicates with other
nodes through HTTP.

The main purpose of this phase is to demonstrate how a blockchain behaves
when multiple nodes operate concurrently and need to exchange information
and eventually reach the same blockchain state.

The Phase 2 implementation introduces:

- Multiple independent blockchain nodes
- Peer management
- Transaction propagation
- Transaction deduplication
- Block propagation
- Block deduplication
- Blockchain synchronization
- Runtime synchronization
- Fork detection
- Longer-chain selection
- Chain reorganization
- Orphaned transaction restoration
- Concurrent state protection
- Three-node cluster execution
- HTTP API endpoints for node inspection
- Unit and integration testing
- Race-condition testing

---

# 2. Objectives

The main objectives of Phase 2 were to extend the existing blockchain
into a small distributed network while preserving the functionality
implemented during Phase 1.

The objectives were:

1. Create an independent `Node` abstraction.
2. Allow multiple blockchain nodes to run simultaneously.
3. Introduce peer-to-peer communication between nodes.
4. Propagate transactions between connected peers.
5. Propagate newly mined blocks between peers.
6. Prevent duplicate transaction and block processing.
7. Allow nodes to synchronize when one node is ahead.
8. Detect temporary forks between nodes.
9. Resolve forks by selecting a longer valid chain.
10. Restore valid transactions from orphaned blocks.
11. Protect shared node state from concurrent access.
12. Verify concurrency using Go's race detector.
13. Provide a simple three-node local cluster for demonstration.
14. Provide API endpoints for inspecting node state.

---

# 3. Development Approach

Phase 2 was developed as an extension of the existing Phase 1
implementation.

The existing blockchain packages were retained and the new networking
functionality was placed around the blockchain rather than replacing the
existing implementation.

The general architecture became:

```text
                    Network
                       |
          ----------------------------
          |            |             |
          v            v             v
       Node 1        Node 2        Node 3
          |            |             |
          v            v             v
     Blockchain   Blockchain    Blockchain