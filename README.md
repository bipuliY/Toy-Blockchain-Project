# Toy Blockchain CLI

A simple blockchain command-line application built with Go.

This project demonstrates the core ideas behind a blockchain: blocks, transactions, a pending transaction pool, proof-of-work mining, deterministic hashing, ledger balances, chain validation, and tamper detection.

---

## Features

- create a new blockchain with a genesis block,
- add transactions to a pending transaction pool,
- mine pending transactions into new blocks,
- use configurable proof-of-work difficulty,
- print the chain,
- validate the chain,
- show account balances,
- reject overspending transactions,
- deliberately tamper with old data for testing,
- persist the chain to JSON.

---

## Requirements

- Go 1.22 or later
- Git (optional but recommended)

Check your installation:

```bash
go version
```

---

## Installation

```bash
git clone <your-repository-url>
cd toy-blockchain
go mod tidy
```

---

## Run the CLI

```bash
go run ./cmd/toychain <command>
```

Common commands:

| Command | Description |
| --- | --- |
| init | create a new blockchain |
| add | add a transaction to the pending pool |
| mine | mine pending transactions into a new block |
| print | print the blockchain |
| validate | validate the chain |
| balances | show account balances |
| pending | show pending transactions |
| tamper | tamper with an old block for testing |
| help | show help |

---

## Example workflow

```bash
go run ./cmd/toychain init -difficulty 2

go run ./cmd/toychain add -from FAUCET -to Alice -amount 100
go run ./cmd/toychain mine

go run ./cmd/toychain add -from Alice -to Bob -amount 30
go run ./cmd/toychain mine

go run ./cmd/toychain balances
```

Expected balance output:

```text
Balances
Alice: 70
Bob: 30
```

Overspending is rejected:

```bash
go run ./cmd/toychain add -from Alice -to Bob -amount 150
```

That command returns:

```text
Error: insufficient balance: Alice has 100 but tried to send 150
```

Tampering and validation:

```bash
go run ./cmd/toychain tamper -block 1 -tx 0 -amount 999
go run ./cmd/toychain validate
```

Expected validation output after tampering:

```text
Chain invalid
First offending block: 1
Reason: stored hash does not match recalculated hash
```

---

## Hashing and validation

Each block hash is computed with SHA-256 over a stable serialisation of the block contents. The fields used are:

1. height,
2. timestamp,
3. transactions,
4. previous hash,
5. nonce.

The current block's own hash field is not included in the calculation to avoid a circular dependency. The implementation uses the standard library's `crypto/sha256` package because it is deterministic, fast, and suitable for a toy blockchain.

This means the same block data always produces the same hash, and any change in the data causes validation to fail.

---

## Design choices

- The implementation uses the standard library only, with no external blockchain SDKs.
- The chain is stored as JSON so it can be reloaded between runs.
- Proof-of-work difficulty is configurable through flags.
- Validation checks block hashes, previous-hash links, height values, timestamps, and transaction validity.

---

## Testing

Run the test suite with:

```bash
go test ./...
```

---

## Limitations

This is a learning-oriented toy chain, not a production blockchain. It does not include peer-to-peer networking, digital signatures, Merkle trees, or full consensus rules.
