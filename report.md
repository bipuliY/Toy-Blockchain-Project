# Toy Blockchain CLI Project Report

## Abstract

This report documents a simplified blockchain implemented in Go as a command-line application. The project was built to meet the requirements of the assessment by demonstrating the core ideas behind a blockchain: deterministic hashing, proof-of-work mining, transaction validation, ledger balances, chain validation, and tamper detection. The implementation is intentionally small and educational, so it focuses on correctness and clarity rather than networking or production-grade consensus.

---

## 1. Introduction

The assignment required a single-process Go application that could create blocks, mine them with proof of work, store transactions, validate the chain, and report tampering. This project satisfies those requirements through a compact CLI that can be run locally from the terminal.

The implementation is organised into separate packages for blocks, the chain, transactions, storage, and the CLI entry point. The chain is persisted to JSON so it can be reused between runs.

---

## 2. Problem Statement

Blockchain systems are often difficult to understand because they combine cryptography, distributed systems, and financial logic. The goal of this project was to simplify the core concepts so they could be studied in a small, understandable example. In particular, the project focuses on:

- how transactions are created and checked,
- how blocks are mined,
- how each block is linked to the previous one,
- how a chain can detect tampering, and
- how account balances can be derived from confirmed transactions.

---

## 3. Project Objectives

The main objectives were to:

1. implement a working toy blockchain in Go,
2. create a CLI for common blockchain actions,
3. add transactions to a pending pool,
4. mine pending transactions into blocks,
5. enforce proof-of-work difficulty,
6. validate the integrity of the full chain,
7. reject invalid or overspending transactions, and
8. document the design choices and experimental results in a short report.

---

## 4. Scope of the Project

### Included in scope

- a single-process command-line application,
- block and chain data structures,
- deterministic SHA-256 hashing,
- a genesis block,
- a basic transaction and ledger model,
- configurable proof-of-work mining,
- full-chain validation and tamper detection,
- unit tests, and
- a written research report.

### Not included in scope

- peer-to-peer networking,
- real wallets or digital signatures,
- smart contracts,
- a web interface, or
- any external blockchain node or SDK.

---

## 5. Implementation Summary

The project uses the following main packages:

- block: defines the block structure and hashing/mining logic,
- chain: manages the blockchain, pending transactions, mining, validation, and balances,
- internal/transaction: defines the transaction model and basic validation,
- ledger: applies transactions to balances and checks whether a sender has enough funds,
- storage: saves and loads the chain as JSON,
- cmd/toychain: exposes the CLI commands.

A typical workflow is:

1. initialise the chain,
2. add one or more transactions,
3. mine the pending transactions into a new block,
4. view balances or the chain,
5. validate the chain, and
6. optionally tamper with an old block to observe validation failure.

---

## 6. Hashing Scheme and Why SHA-256 Was Used

### 6.1 Hashing design

Each block uses SHA-256 to calculate a deterministic hash. The hash is computed from a stable serialisation of the block data using Go's standard library. The exact fields included in the hash input are:

1. Height,
2. Timestamp,
3. Transactions,
4. Previous hash,
5. Nonce.

The current block's own hash field is intentionally excluded. Including it would create a circular dependency because the hash is the value being calculated.

The transaction data is also serialized in a stable way. In the implementation, the transaction fields are stored in the order `From`, `To`, and `Amount`.

The hash input is therefore derived from a stable JSON structure rather than a manually formatted string. This makes the operation deterministic: the same block data will always produce the same hash.

### 6.2 Why SHA-256 was chosen

SHA-256 was selected because it is:

- deterministic, so the same input always produces the same output,
- fast enough for a toy blockchain on a laptop,
- standardised and widely used in real blockchain systems,
- easy to use with Go's `crypto/sha256` package.

For this project, SHA-256 is sufficient because the goal is to demonstrate how hashing links blocks together and makes tampering observable. It is not intended to be a production-grade security implementation.

---

## 7. Experiments and Results

### 7.1 Tamper-evidence experiment

One of the key experiments was to tamper with an existing transaction and then run validation.

#### Procedure

1. Create a new chain.
2. Add a faucet transaction and mine it.
3. Validate the chain successfully.
4. Tamper with the amount in the first mined block.
5. Validate the chain again.

#### Observed output

Before tampering:

```text
$ go run ./cmd/toychain validate
Chain valid
```

After tampering:

```text
$ go run ./cmd/toychain tamper -block 1 -tx 0 -amount 999
Tampered block 1 transaction 0 amount: 100 -> 999
Important: hash was not recalculated, so validation should fail now.

$ go run ./cmd/toychain validate
Chain invalid
First offending block: 1
Reason: stored hash does not match recalculated hash
```

#### Interpretation

The chain becomes invalid because the stored hash of block 1 no longer matches the hash recomputed from the altered data. Validation detects this immediately and identifies the first offending block. This shows that even a small change in an early block breaks the chain's integrity.

### 7.2 Difficulty versus mining effort

The mining step uses proof of work by searching for a nonce such that the block hash begins with the required number of leading zero hex digits. Higher difficulty means more hashes must be tried before success.

The following results were observed on this machine:

| Difficulty | Hashes tried | Approx. time |
| --- | ---: | ---: |
| 4 | 119,724 | 0.08 s |
| 5 | 371,601 | 0.16 s |
| 6 | 44,060,072 | 14.77 s |

The trend is not linear. For a difficulty target of $d$ leading zero hex digits, the expected work grows roughly like $16^d$ because each extra zero makes the target much smaller. This is why the mining time rises quickly as difficulty increases.

---

## 8. Ledger Behaviour and Example Balances

The ledger model is intentionally simple. A normal transaction subtracts the amount from the sender and adds it to the recipient. A faucet transaction is treated as a special case that credits the recipient directly without requiring an existing balance.

Example:

```bash
go run ./cmd/toychain init -difficulty 2
go run ./cmd/toychain add -from FAUCET -to Alice -amount 100
go run ./cmd/toychain mine
go run ./cmd/toychain add -from Alice -to Bob -amount 30
go run ./cmd/toychain mine
go run ./cmd/toychain balances
```

Observed output:

```text
Balances
Alice: 70
Bob: 30
```

This example also shows that overspending is rejected. For instance:

```bash
go run ./cmd/toychain add -from Alice -to Bob -amount 150
```

The command fails with:

```text
Error: insufficient balance: Alice has 100 but tried to send 150
```

---

## 9. Discussion Questions

### How does the previous-hash link help security?

Each block stores the hash of the previous block. If an old block is modified, the hash of that block changes and the next block's stored previous-hash value no longer matches. This makes tampering visible, especially when validation checks the whole chain.

### What is an alternative to proof-of-work?

Proof-of-stake is one alternative. It lets a participant create the next block based on the amount of stake they hold rather than on solving a computational puzzle. A benefit is lower energy consumption, while a drawback is that wealth can influence block production more heavily.

### How does this toy chain differ from a production blockchain?

Three concrete differences are:

1. there is no network of peers or consensus among nodes,
2. there are no transaction signatures or real wallet authentication,
3. there is no Merkle tree or full economic finality model.

A natural improvement would be to add digital signatures so that transactions can be verified by public/private keys rather than by a simple local ledger.

---

## 10. Limitations and Future Improvements

The project is intentionally educational and therefore has several limitations:

- it does not support a distributed network,
- it does not provide real cryptographic signatures,
- it does not implement Merkle trees,
- it uses a simple local JSON persistence model, and
- it does not support forks or chain selection rules.

Future work could focus on digital signatures, Merkle roots, improved validation messages, and a richer CLI.

---

## 11. Conclusion

The toy blockchain project successfully demonstrates the main ideas behind blockchain technology in a compact and understandable form. It shows how hashing, proof of work, ledger rules, and validation together produce a simple tamper-evident chain. Although it is not a production blockchain, it is a solid educational implementation that meets the assessment requirements and provides a clear foundation for further experimentation.
