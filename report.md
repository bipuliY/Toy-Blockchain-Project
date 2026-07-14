# Toy Blockchain and Ledger Simulator

## Research Report

## 1. Introduction

This project is a small blockchain and ledger simulator implemented in Go. It runs locally as a command-line tool and does not communicate with external blockchain networks. It supports transactions, account balances, a pending transaction pool, proof-of-work mining, JSON persistence, full-chain validation, and deliberate tampering for testing. A Merkle root is also calculated for each block so the project can demonstrate how a block's transactions can be summarised and later validated.

The purpose of this report is to describe the implementation, present experiments run against the code, and explain how the toy system differs from a production blockchain.

## 2. Implementation Summary

The core design is intentionally simple:

- A blockchain is stored as a JSON file and reloaded on demand.
- The first block is a deterministic genesis block with a fixed previous hash.
- New transactions are appended to a pending pool.
- Mining selects transactions from the pool and builds a new block.
- Each block stores its transactions, a Merkle root, a previous-block hash, a nonce, and its final hash.
- Validation checks the block height, Merkle root, block hash, previous-hash link, proof-of-work requirement, and transaction validity.

The CLI also supports optional Ed25519 signatures for non-faucet transactions. If a sender is not the special `FAUCET` account, the project requires a private key to sign the transaction.

## 3. Hashing and Validation Design

### 3.1 Transaction hashing

Each transaction is serialised using Go's standard JSON encoder and hashed using SHA-256. The resulting digest is deterministic for a given transaction payload. Changing the sender, recipient, amount, or signature changes the hash.

### 3.2 Merkle root construction

Each transaction hash becomes a leaf in a Merkle tree. Adjacent hashes are concatenated and hashed again until a single root remains. If the number of hashes at a level is odd, the last hash is duplicated so the pairing can continue. This means the Merkle root is a compact summary of the transaction set and is sensitive to transaction order and content.

### 3.3 Block hashing

The block hash is calculated from a stable JSON structure containing:

1. Height
2. Timestamp
3. Merkle root
4. Previous hash
5. Nonce

The raw transaction list is not directly included in the block-hash input. Instead, the Merkle root summarises the transactions. The block's own hash is excluded to avoid a circular dependency.

### 3.4 Why validation is meaningful

Validation is meaningful because the chain integrity check uses the same transaction data that the block was mined with. If any transaction changes, the recalculated Merkle root changes. That causes the stored Merkle root and recalculated Merkle root to diverge, and validation immediately reports the block as invalid.

## 4. Research Component

### 4.1 Required investigation: tamper evidence

#### Objective

The objective was to determine whether a transaction changed inside an already mined block would be detected by validation.

#### Procedure

A new chain was created, a faucet transaction was added, a second transaction was added, and the block was mined. The first transaction in the mined block was then changed from `100` to `999` using the CLI tamper command. Validation was run afterwards.

#### Commands run

```bash
go run ./cmd/toychain init -difficulty 2
go run ./cmd/toychain add -from FAUCET -to Alice -amount 100
go run ./cmd/toychain add -from FAUCET -to Bob -amount 50
go run ./cmd/toychain mine
go run ./cmd/toychain tamper -block 1 -tx 0 -amount 999
go run ./cmd/toychain validate
```

#### Before and after output

Before tampering:

```text
Tampered block 1 transaction 0 amount: 100 -> 999
Important: hash was not recalculated, so validation should fail now.
```

After validation:

```text
Chain invalid
First offending block: 1
Reason: stored Merkle root does not match block transactions
```

#### Explanation

The change to the transaction altered the serialised transaction bytes, so the transaction hash changed. Because the Merkle root is derived from the transaction hashes, the recalculated Merkle root also changed. The block still contained the old Merkle root from before the tampering, so the stored Merkle root no longer matched the block transactions. The validation code catches this first and reports the block as the first offending block.

### 4.2 Required investigation: difficulty versus effort

#### Objective

The objective was to observe how increasing proof-of-work difficulty affected the amount of work needed to mine a block.

#### Experiment

A block was mined at several difficulty levels. The number of hashes tried and the time taken were recorded.

| Difficulty | Hashes tried | Time taken |
| --- | ---: | ---: |
| 2 | 663 | 0 ms |
| 3 | 2038 | 0 ms |
| 4 | 54472 | 18 ms |
| 5 | 2190554 | 713 ms |

#### Interpretation

The growth is much faster than linear. Increasing the number of required leading zeroes makes the expected search space grow exponentially. In practice, this means that each additional zero character can make mining dramatically harder. The results are consistent with the idea that a difficulty of $d$ requires a hash prefix of roughly $d$ zeroes, which makes the expected work grow on the order of $16^d$ in the simplest model.

### 4.3 Design write-up

The project uses a straightforward integrity model:

1. Transactions are hashed individually.
2. A Merkle root is built from those hashes.
3. The Merkle root is included in the block hash input.
4. The block hash is then mined to satisfy the required difficulty.
5. Validation recalculates the Merkle root and the block hash and compares them with the stored values.

This design ensures that any change to a transaction or block content will break the expected hash chain. Because a later block stores the previous block's hash, changing an older block would also break the chain linkage unless every following block were recomputed and re-mined.

## 5. Discussion Questions

### 5.1 How does the previous-hash link make tampering with an old block impractical in a real chain?

In a real blockchain, each block stores the hash of the previous block. If an old block is changed, its hash changes and the next block's previous-hash link becomes invalid. To hide the modification, an attacker would need to recompute the changed block, redo its proof of work, and update every later block. In a distributed system, this becomes very difficult because honest nodes continue extending the valid chain while the attacker tries to recreate the old work.

### 5.2 What is an alternative to proof of work?

One alternative is proof of stake. In proof of stake, validators are chosen based on the stake they hold rather than by repeatedly hashing. A benefit is that it uses far less computational energy. A drawback is that participants with more stake may gain more influence, so the system must carefully design incentives and penalties.

### 5.3 Three concrete ways this toy differs from a production blockchain

1. It has no peer-to-peer network or distributed consensus among independent nodes.
2. It does not provide full Merkle proofs for individual transactions.
3. It stores the complete chain in one local JSON file rather than distributing it across many peers.

A practical improvement would be to add Merkle proofs so a client could verify that one transaction belongs to a block without downloading and hashing every transaction in the block.

## 6. Sources

- Satoshi Nakamoto, “Bitcoin: A Peer-to-Peer Electronic Cash System” (white paper).
- The project source code and test suite in this repository.
