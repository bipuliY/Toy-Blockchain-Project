# Toy Blockchain CLI

A simple blockchain command-line application built with **Go**.

This project demonstrates the core concepts behind a blockchain, including blocks, transactions, a pending transaction pool, proof-of-work mining, Merkle trees, deterministic hashing, ledger balances, chain validation, JSON persistence, and tamper detection.

---

## Features

* Create a new blockchain with a genesis block
* Add transactions to a pending transaction pool
* Mine pending transactions into new blocks
* Use configurable proof-of-work difficulty
* Summarise block transactions using a SHA-256 Merkle root
* Store the Merkle root inside each block
* Print blocks, transactions, hashes, and Merkle roots
* Validate the complete blockchain
* Show account balances
* Reject overspending transactions
* Detect modified transactions using Merkle-root validation
* Deliberately tamper with old transaction data for testing
* Persist the blockchain to a JSON file
* Support optional Ed25519 transaction signatures

---

## Requirements

* Go 1.22 or later
* Git, optional but recommended

Check the installed Go version:

```bash
go version
```

---

## Installation

Clone the repository:

```bash
git clone https://github.com/bipuliY/Toy-Blockchain-Project.git
cd Toy-Blockchain-Project
```

Install or update the required Go modules:

```bash
go mod tidy
```

---

## Run the CLI

Use the following command format:

```bash
go run ./cmd/toychain <command>
```

### Available commands

| Command    | Description                                       |
| ---------- | ------------------------------------------------- |
| `init`     | Create a new blockchain                           |
| `add`      | Add a transaction to the pending transaction pool |
| `mine`     | Mine pending transactions into a new block        |
| `print`    | Print the complete blockchain                     |
| `validate` | Validate the blockchain                           |
| `balances` | Show account balances                             |
| `pending`  | Show pending transactions                         |
| `tamper`   | Modify an old transaction for testing             |
| `help`     | Show available commands                           |

---

## Example workflow

Create a new blockchain:

```bash
go run ./cmd/toychain init -difficulty 2
```

Add an initial transaction from the faucet:

```bash
go run ./cmd/toychain add -from FAUCET -to Alice -amount 100
```

Mine the pending transaction:

```bash
go run ./cmd/toychain mine
```

Transfer funds from Alice to Bob:

```bash
go run ./cmd/toychain add -from Alice -to Bob -amount 30
go run ./cmd/toychain mine
```

Display account balances:

```bash
go run ./cmd/toychain balances
```

Expected output:

```text
Balances
Alice: 70
Bob: 30
```

---

## Overspending protection

The blockchain checks the sender's available balance before accepting a transaction.

For example:

```bash
go run ./cmd/toychain add -from Alice -to Bob -amount 150
```

Expected output:

```text
Error: insufficient balance: Alice has 100 but tried to send 150
```

The invalid transaction is not added to the pending transaction pool.

---

## Merkle root implementation

Each block contains a list of transactions.

Instead of directly including the complete raw transaction list in the block-hash calculation, the transactions are first summarised using a **Merkle root**.

The process is:

```text
Transactions
     ↓
Hash every transaction using SHA-256
     ↓
Combine adjacent transaction hashes
     ↓
Hash each combined pair
     ↓
Continue until one hash remains
     ↓
Merkle root
```

For example, with four transactions:

```text
                     Merkle Root
                          |
                  -----------------
                  |               |
               Hash AB         Hash CD
               /    \          /    \
           Hash A  Hash B  Hash C  Hash D
              |       |       |       |
             TX1     TX2     TX3     TX4
```

If a Merkle-tree level contains an odd number of hashes, the final hash is duplicated before generating the next level.

For example:

```text
A, B, C
```

becomes:

```text
A, B, C, C
```

The resulting pairs are:

```text
Hash(A + B)
Hash(C + C)
```

This process continues until one final hash remains.

---

## Block structure

Each block stores the following information:

1. Block height
2. Timestamp
3. Transactions
4. Merkle root
5. Previous block hash
6. Nonce
7. Current block hash

A printed block looks similar to:

```text
Height: 1
Timestamp: 1784002436
Previous hash: 97dd171d...
Merkle root: 69b353b4...
Nonce: 24
Hash: 007143e8...
Transactions:
  FAUCET -> Alice : 100
  FAUCET -> Bob : 50
```

---

## Block hashing

Each block hash is calculated using SHA-256 over a stable serialisation of these fields:

1. Height
2. Timestamp
3. Merkle root
4. Previous block hash
5. Nonce

The raw transaction list is not directly included in the block-hash input.

Conceptually:

```text
Block Hash = SHA-256(
    Height +
    Timestamp +
    Merkle Root +
    Previous Hash +
    Nonce
)
```

The block's own hash field is not included in the calculation because that would create a circular dependency.

The same block data always produces the same hash. Changing any included field produces a different block hash.

---

## Blockchain validation

The validation process checks every block in the chain.

For each block, the program:

1. Checks that the block height is correct
2. Recalculates the Merkle root from the stored transactions
3. Compares the recalculated Merkle root with the stored Merkle root
4. Recalculates the block hash
5. Compares the recalculated block hash with the stored block hash
6. Checks the previous-block hash link
7. Checks that timestamps are in the correct order
8. Checks that the block hash meets the proof-of-work difficulty
9. Validates every transaction
10. Updates the temporary account balances

The important Merkle-root validation is:

```text
Stored Merkle Root
        compared with
Merkle Root Recalculated from Transactions
```

If someone modifies even one transaction, the recalculated Merkle root changes and validation fails.

---

## Tamper detection

To deliberately modify a transaction:

```bash
go run ./cmd/toychain tamper -block 1 -tx 0 -amount 999
```

Then validate the blockchain:

```bash
go run ./cmd/toychain validate
```

Expected result:

```text
Chain invalid
First offending block: 1
Reason: stored Merkle root does not match block transactions
```

This happens because:

```text
Original transaction
        ↓
Original transaction hash
        ↓
Stored Merkle root
```

After tampering:

```text
Modified transaction
        ↓
Different transaction hash
        ↓
Different recalculated Merkle root
```

The new Merkle root no longer matches the Merkle root stored when the block was mined.

---

## Proof-of-work mining

Mining repeatedly changes the block nonce and recalculates the block hash.

Mining continues until the block hash begins with the required number of zeroes.

For difficulty `2`, a valid hash may look like:

```text
007143e8d37312ffd3d5942c0a96910f93c0f18b4d0fec5c29a855f2262090a6
```

The first two characters are zeroes:

```text
00
```

A higher difficulty requires more leading zeroes and normally requires more hash attempts.

---

## JSON persistence

The blockchain is stored in:

```text
data/chain.json
```

This allows the application to reload the blockchain between separate command executions.

Because the Merkle-root implementation changed the block structure and hashing algorithm, old blockchain files created before this implementation may not be compatible.

Delete the old chain before starting a new test:

```bash
rm data/chain.json
```

This is the correct command for macOS and Linux terminals.

---

## Complete project run example

Remove the old blockchain file:

```bash
rm data/chain.json
```

Create a new blockchain:

```bash
go run ./cmd/toychain init -difficulty 2
```

Add transactions:

```bash
go run ./cmd/toychain add -from FAUCET -to Alice -amount 100
go run ./cmd/toychain add -from FAUCET -to Bob -amount 50
```

Show pending transactions:

```bash
go run ./cmd/toychain pending
```

Mine the pending transactions:

```bash
go run ./cmd/toychain mine
```

Print the blockchain:

```bash
go run ./cmd/toychain print
```

Validate the blockchain:

```bash
go run ./cmd/toychain validate
```

Expected result:

```text
Chain valid
```

Tamper with a transaction:

```bash
go run ./cmd/toychain tamper -block 1 -tx 0 -amount 999
```

Validate again:

```bash
go run ./cmd/toychain validate
```

Expected result:

```text
Chain invalid
First offending block: 1
Reason: stored Merkle root does not match block transactions
```

---

## Testing

Format all Go files:

```bash
gofmt -w .
```

Run the complete test suite:

```bash
go test ./...
```

Run only the chain tests with detailed output:

```bash
go test ./chain -v
```

The Merkle-root tampering test should produce output similar to:

```text
=== RUN   TestValidationDetectsMerkleRootMismatch
--- PASS: TestValidationDetectsMerkleRootMismatch
PASS
```

The project includes tests for:

* Deterministic block hashing
* Deterministic Merkle-root generation
* Genesis-block creation
* Proof-of-work mining
* Honest-chain validation
* Transaction tampering detection
* Merkle-root mismatch detection
* Overspending rejection
* Invalid transaction rejection
* Blockchain storage and loading

---

## Design choices

* The project uses Go's standard library without external blockchain SDKs.
* SHA-256 is used for transaction hashes, Merkle-tree nodes, and block hashes.
* Transactions remain stored in each block so they can be displayed and validated.
* The Merkle root provides one deterministic summary of all block transactions.
* The transaction order affects the Merkle root.
* The final hash is duplicated when a Merkle-tree level contains an odd number of nodes.
* The genesis block uses the SHA-256 hash of an empty byte sequence as its Merkle root.
* The chain is stored as JSON so it can be inspected and reloaded.
* Proof-of-work difficulty and block size are configurable.
* Validation separately checks the Merkle root and block hash.

---

## Limitations

This is a learning-oriented toy blockchain and should not be used as a production cryptocurrency or financial system.

It does not currently include:

* Peer-to-peer networking
* Distributed blockchain nodes
* Full consensus rules
* Merkle proofs for individual transactions
* Wallet management
* Transaction fees
* Mining rewards
* Production-level key security
* Protection against all real-world blockchain attacks

---

## Project purpose

The purpose of this project is to demonstrate how important blockchain concepts work together:

```text
Transactions
     ↓
Pending transaction pool
     ↓
Merkle-root generation
     ↓
Proof-of-work mining
     ↓
Block creation
     ↓
Previous-hash linking
     ↓
Blockchain validation
     ↓
Tamper detection
```

The project is intended for learning, experimentation, and demonstrating the basic internal behaviour of a blockchain.
