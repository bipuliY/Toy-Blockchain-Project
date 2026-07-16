# Toy Blockchain and Ledger Simulator

## Research Report

### 1. Introduction

This project is a local, command-line blockchain and ledger simulator implemented in Go. Its purpose is to demonstrate core blockchain ideas without relying on a real network or external nodes. The implementation includes blocks, transactions, a pending transaction pool, proof-of-work mining, Merkle roots, chain validation, JSON persistence, fork-resolution logic, and optional Ed25519-based transaction signatures. The project is therefore best understood as an educational simulator rather than a production-ready blockchain system.

The implementation is organized around a small set of packages: [block/block.go](block/block.go) defines the block model and mining logic; [chain/chain.go](chain/chain.go) and [chain/fork.go](chain/fork.go) implement the blockchain state machine, validation, difficulty adjustment, and fork resolution; [ledger/ledger.go](ledger/ledger.go) applies and checks balances; [merkle/merkle.go](merkle/merkle.go) computes Merkle roots; [internal/transaction/transaction.go](internal/transaction/transaction.go) defines transactions and signature handling; [storage/storage.go](storage/storage.go) persists the blockchain as JSON; and [cmd/toychain/main.go](cmd/toychain/main.go) provides the CLI.

This version also includes the optional stretch-goal features requested for the project: Ed25519 digital signatures for non-faucet transactions, Merkle roots for summarizing block transactions, concurrent mining across goroutines, automatic difficulty retargeting, and fork resolution that validates competing chains and adopts the longest valid one. The implementation is present in the current codebase, with one important nuance: digital-signature enforcement is implemented in the validation path used by the CLI, but the lower-level transaction validator still allows older unsigned non-faucet transactions when both signature fields are empty.

### 2. System Design

The overall architecture is intentionally simple and single-process. The CLI loads or saves a blockchain from a JSON file, adds transactions to a pending pool, mines blocks, prints the chain, validates the chain, resolves forks between candidate chains, and shows balances. The system uses the standard library and a few cryptographic packages from Go, including SHA-256 and Ed25519.

#### Optional stretch-goal features implemented in the current project

The codebase currently implements the following optional features:

- Digital signatures: non-faucet transactions can be signed with Ed25519 keys using the CLI, and verification is performed during transaction validation. The current implementation also preserves backward compatibility for unsigned non-faucet transactions when no signature fields are supplied.
- Merkle roots: each block stores a Merkle root calculated from the block’s transactions and validation checks that the stored root matches the recomputed root.
- Concurrent mining: block mining can run across multiple goroutines and stop as soon as one worker finds a valid nonce.
- Difficulty retargeting: the blockchain adjusts difficulty based on the observed block-production cadence and validates the expected difficulty at retarget points.
- Fork resolution: the project accepts competing chains, validates them, and adopts the longest valid chain under the implemented rule.

These features are implemented in the current codebase and are reflected in the corresponding modules and tests.

The main data structures are straightforward:

- A blockchain is represented by the struct in [chain/chain.go](chain/chain.go). It contains a slice of blocks, a slice of pending transactions, and configuration fields for difficulty, block size, retargeting, and difficulty limits.
- A block is represented by the struct in [block/block.go](block/block.go). Each block stores a height, timestamp, transaction list, Merkle root, previous hash, difficulty, nonce, and hash.
- A transaction is represented by the struct in [internal/transaction/transaction.go](internal/transaction/transaction.go). It includes sender, recipient, amount, and optional public-key and signature fields.
- Validation results are returned as a simple struct in [chain/chain.go](chain/chain.go), with fields indicating validity, the offending block height, and the failure reason.

The blockchain begins with a deterministic genesis block created by the function `NewGenesisBlock()`. The genesis block has height 0, a fixed previous hash constant, an empty transaction list, and a Merkle root computed from empty input. New blocks are created by `NewBlock()`, which copies the transactions, computes a Merkle root, and records the previous block hash.

Mining is implemented in [block/block.go](block/block.go). The block hash is mined by searching for a nonce that satisfies the requested difficulty, and the code can mine either sequentially or concurrently across goroutines. The blockchain’s `MinePending()` method selects transactions from the pending pool up to the configured block size, constructs a new block, mines it, appends it to the chain, and removes mined transactions from the pending pool.

The project also includes a fork-resolution mechanism in [chain/fork.go](chain/fork.go). It validates candidate chains, rejects incompatible ones, and adopts the longest valid chain when appropriate. The local pending transactions are then re-evaluated against the adopted chain.

Persistence is implemented through [storage/storage.go](storage/storage.go), which saves the blockchain to JSON and loads it again later. The CLI commands use this storage layer to make the blockchain state persistent across invocations.

### 3. Hashing Design

The hashing design is explicit and deterministic. Transactions are hashed individually in [merkle/merkle.go](merkle/merkle.go) by serializing the transaction using Go’s JSON encoder and applying SHA-256. The resulting leaf hashes are then combined in a Merkle tree structure: adjacent hashes are concatenated and hashed again, and if an odd number of hashes remains at a level, the last hash is duplicated so the pairing process continues. This produces a single Merkle root for the block’s transaction set.

The block hash is computed from a dedicated input structure in [block/block.go](block/block.go). The fields included in the block-hash input are:

1. Height
2. Timestamp
3. Merkle root
4. Previous hash
5. Difficulty
6. Nonce

The raw transaction list is not included directly in the block hash. Instead, the Merkle root acts as a compact summary of the transactions. This choice is explicit in the implementation comment and is also reinforced by the code paths that recompute the Merkle root before mining. The hash field itself is excluded from the hash calculation to avoid circular dependency; the block hash is computed from other block fields rather than from the hash value it would produce.

Deterministic hashing works because the implementation uses stable serialization and fixed field ordering. For transactions, the struct fields are serialized in a stable order and hashed with SHA-256. For blocks, the hash input is a struct that is marshaled to JSON using the order of declaration in the Go code. Therefore, the same block contents yield the same hash as long as the same nonce and other fields are used.

### 4. Validation Design

Validation is implemented in [chain/chain.go](chain/chain.go) and is fairly comprehensive for a toy system. The validation loop checks the following conditions in order:

- The chain must contain at least one block.
- The configured blockchain parameters must be sane, including a positive block size and valid difficulty limits.
- Block heights must match their index in the slice.
- Each block must not exceed the configured block-size limit.
- The stored Merkle root must match the Merkle root recomputed from the block’s transactions.
- The stored hash must match the hash recomputed from the block’s contents.
- The genesis block must have the required fixed previous hash and zero difficulty.
- Each non-genesis block must have a previous hash that matches the prior block’s hash.
- Timestamps must be non-decreasing relative to the previous block.
- Each block’s difficulty must match the expected difficulty at that point in the chain.
- The block hash must satisfy the proof-of-work requirement for its stored difficulty.
- Each transaction must pass the ledger validation logic and must not cause unsound balances.

The ledger validation logic in [ledger/ledger.go](ledger/ledger.go) applies transactions sequentially to a running balance map. A faucet transaction adds funds, while a non-faucet transaction deducts and credits amounts. A transaction is rejected if it fails basic validation or if the sender does not have enough available balance. For signed transactions, validation depends on the transaction’s public key and signature fields. The CLI requires a private key for non-faucet transactions, but the lower-level validator still accepts unsigned non-faucet transactions when both signature fields are empty. This means the signature feature is implemented and used in the CLI workflow, but the current code still retains a compatibility path for older unsigned transactions.

### 5. Required Investigation 1 – Tamper Evidence

The project includes explicit tampering support through the CLI command `tamper` in [cmd/toychain/main.go](cmd/toychain/main.go). The command changes a transaction amount inside a selected block and saves the modified blockchain without recalculating the block hash. This is designed specifically for research and validation experiments.

If an earlier block’s transaction is changed, the validation process detects the tampering. In the current implementation, the first failure is usually the Merkle-root check because the transaction content changed but the stored Merkle root remained from the original block. The validation logic recomputes the Merkle root from the modified transactions and compares it with the stored one. If the two differ, validation stops and reports the offense. The reason returned is: “stored Merkle root does not match block transactions”.

The intended experiment is:

```bash
go run ./cmd/toychain init -difficulty 2
go run ./cmd/toychain add -from FAUCET -to Alice -amount 100
go run ./cmd/toychain add -from FAUCET -to Bob -amount 50
go run ./cmd/toychain mine
go run ./cmd/toychain tamper -block 1 -tx 0 -amount 999
go run ./cmd/toychain validate
```

The implementation is expected to report a failure similar to:

```text
Chain invalid
First offending block: 1
Reason: stored Merkle root does not match block transactions
```

This demonstrates that tampering with block data is detected by the validation logic.

### 6. Required Investigation 2 – Difficulty vs Mining Effort

The repository does not include a dedicated benchmark harness or committed benchmark output. The code does, however, expose the relevant mining statistics through the `MineResult` structure in [block/block.go](block/block.go): the CLI prints the number of hashes tried and the elapsed time after mining. A practical experiment can therefore be run manually by mining the same transaction at different difficulties and recording the reported values.

Suggested procedure:

```bash
go run ./cmd/toychain init -difficulty <N>
go run ./cmd/toychain add -from FAUCET -to Alice -amount 100
go run ./cmd/toychain mine
```

Repeat for several values of `N` and record the output fields `Hashes tried` and `Time taken`.

A suitable table template is:

| Difficulty | Hashes tried | Time taken |
|------------|-------------:|-----------:|
| 1 |  |  |
| 2 |  |  |
| 3 |  |  |
| 4 |  |  |
| 5 |  |  |

The observed trend follows directly from the implementation: the mining function checks hashes until the required prefix of zeroes is found. Increasing the difficulty means requiring more leading zeroes, which expands the expected search space substantially. In other words, each additional difficulty level makes mining materially more expensive on average.

### 7. Required Investigation 3 – Design Discussion

The project uses a simple but effective integrity strategy. First, individual transactions are hashed and summarized by a Merkle root. Second, the Merkle root is included in the block-hash input along with the previous hash, timestamp, difficulty, height, and nonce. Third, the block hash is mined to satisfy a target prefix. Finally, validation recomputes the Merkle root and the hash and compares them with the stored values.

This design provides integrity because any modification to transaction content or block metadata changes the derived values used for hashing. In addition, each non-genesis block links to the previous block’s hash, so validation also detects broken chain links. The implementation therefore guarantees that an altered block will not pass validation unless its downstream metadata and proof-of-work values are recomputed consistently.

The design decisions are consistent with the project’s educational goals: a single local chain, deterministic hashing, transparent validation, and a simple CLI interface. The code favors clarity and verifiability over scalability or distributed coordination.

### 8. Discussion Questions

1. Why do previous-hash links make tampering difficult?

   Previous-hash links make tampering difficult because each block depends on the hash of the block that came before it. If an earlier block is changed, its hash changes, which breaks the link from the next block unless that next block is also modified and re-mined. In a real blockchain, that requires recomputing work for every downstream block, making tampering expensive and visible.

2. Proof-of-Work versus Proof-of-Stake or Proof-of-Authority.

   Proof-of-Work requires miners to find a hash with a certain prefix. One advantage is that it is simple to verify and makes the work expensive to fake. One disadvantage is that it consumes significant computational energy. Proof-of-Stake replaces hashing competition with staking-based selection, which is more energy efficient but can concentrate influence among wealthier participants. The current project uses Proof-of-Work only.

3. Three differences between this toy blockchain and a production blockchain.

   - This project has no peer-to-peer networking or distributed consensus among independent nodes.
   - It does not implement Merkle proofs for individual transactions, only Merkle roots for whole blocks.
   - It does not provide smart contracts, finality guarantees, or a full wallet and key-management ecosystem.

   One improvement that could be added naturally to this project is Merkle proofs. The implementation already computes Merkle roots, so proofs could be added to let a client verify that a specific transaction is included in a block without hashing the entire transaction set.

### 9. Testing

The test suite is concentrated in the package-specific test files. The tests cover several core behaviors:

- Hash determinism tests in [block/block_test.go](block/block_test.go) and [merkle/merkle_test.go](merkle/merkle_test.go)
- Mining tests that check whether mined hashes satisfy the requested difficulty in [block/block_test.go](block/block_test.go)
- Validation tests that confirm honest chains validate and tampered chains fail in [chain/chain_test.go](chain/chain_test.go)
- Transaction rejection tests for overspending and negative amounts in [chain/chain_test.go](chain/chain_test.go)
- Tamper detection tests for Merkle-root mismatches in [chain/chain_test.go](chain/chain_test.go)
- Fork-resolution tests in [chain/fork_test.go](chain/fork_test.go)
- Difficulty-retargeting tests in [chain/retarget_test.go](chain/retarget_test.go)
- Persistence tests in [storage/storage_test.go](storage/storage_test.go)

The suite is solid for a toy implementation, but it does not include CLI-level integration tests or dedicated tests for signature-verification failures. Those areas remain uncovered by the current automated tests.

### 10. Limitations

The current implementation has several real limitations:

- It is not a distributed blockchain and has no peer-to-peer network.
- It has no consensus algorithm beyond local validation and fork selection.
- It does not implement Merkle proofs or lightweight transaction verification.
- It uses local JSON files for persistence rather than a database or replicated storage layer.
- Difficulty retargeting is simple and based on coarse timestamps.
- The transaction model is lightweight and does not include a full wallet, account model, or advanced key management system.
- The lower-level validator still allows unsigned non-faucet transactions when both signature fields are empty, so signature enforcement is only partially strict at the API boundary.

### 11. Possible Future Improvements

A few improvements would extend the project in a natural way:

- Add Merkle proofs for individual transactions.
- Add explicit CLI flags for worker count in concurrent mining.
- Add CLI integration tests for init, add, mine, validate, and tamper flows.
- Improve the retargeting logic with more robust timing rules.
- Add stronger transaction authentication enforcement so all non-faucet transactions must be signed.
- Add networking and a simple distributed consensus mechanism if the goal is to move beyond a local simulator.

### 12. Conclusion

The project successfully implements a compact blockchain simulator in Go. It demonstrates the core ideas of block creation, pending transaction pools, proof-of-work mining, Merkle roots, chain validation, persistence, fork resolution, and optional digital signatures. The validation logic is strong enough to detect tampering, and the code is organized clearly around the packages responsible for blocks, chain state, transaction validation, hashing, storage, and the CLI. The implementation is educational and deliberately simple, but it is also complete enough to show how integrity and validation are handled in a blockchain-like system.
