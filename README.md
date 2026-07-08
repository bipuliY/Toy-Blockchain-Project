# Toy Blockchain CLI

A simple blockchain command line application built with **Go**.

This project demonstrates the basic working concepts of a blockchain, including blocks, transactions, pending transaction pool, proof-of-work mining, hash linking, blockchain validation, balances, and tamper detection.

---

## Features

* Create a new blockchain with a genesis block
* Add transactions to a pending transaction pool
* Mine pending transactions into new blocks
* Use proof-of-work difficulty during mining
* Print the full blockchain
* Validate the blockchain
* Show account balances
* Show pending transactions
* Deliberately tamper with old data for testing
* Store blockchain data in a JSON file

---

## Project Structure

```text
toy-blockchain/
│
├── block/
│   ├── block.go
│   └── block_test.go
│
├── chain/
│   ├── chain.go
│   └── chain_test.go
│
├── cmd/
│   └── toychain/
│       └── main.go
│
├── data/
│   └── chain.json
│
├── internal/
│   └── transaction/
│       └── transaction.go
│
├── ledger/
│   └── ledger.go
│
├── storage/
│   ├── storage.go
│   └── storage_test.go
│
├── go.mod
└── README.md
```

---

## Requirements

Before running this project, install:

* Go 1.20 or later
* Git, optional but recommended

Check Go installation:

```bash
go version
```

---

## Installation

Clone the repository:

```bash
git clone <your-repository-url>
cd toy-blockchain
```

Or open your existing project folder:

```bash
cd path/to/toy-blockchain
```

Download dependencies if needed:

```bash
go mod tidy
```

---

## How to Run

The main CLI application is inside:

```text
cmd/toychain/main.go
```

Run commands using:

```bash
go run ./cmd/toychain <command>
```

Example:

```bash
go run ./cmd/toychain help
```

---

## Available Commands

| Command    | Description                                |
| ---------- | ------------------------------------------ |
| `init`     | Create a new blockchain file               |
| `add`      | Add a transaction to the pending pool      |
| `mine`     | Mine pending transactions into a new block |
| `print`    | Print the full blockchain                  |
| `validate` | Validate the blockchain                    |
| `balances` | Show account balances                      |
| `pending`  | Show pending transactions                  |
| `tamper`   | Deliberately modify old data for testing   |
| `help`     | Show help message                          |

---

## Common Flags

| Flag          | Description                              | Default           |
| ------------- | ---------------------------------------- | ----------------- |
| `-file`       | Path to blockchain JSON file             | `data/chain.json` |
| `-difficulty` | Proof-of-work difficulty for a new chain | Project default   |
| `-block-size` | Maximum transactions per block           | Project default   |

Example:

```bash
go run ./cmd/toychain init -difficulty 2 -block-size 3
```

---

## Usage Guide

### 1. Create a New Blockchain

```bash
go run ./cmd/toychain init -difficulty 2
```

This creates a new blockchain with a genesis block and saves it into:

```text
data/chain.json
```

Example output:

```text
New blockchain created
File: data/chain.json
Difficulty: 2
Block size: 5
Genesis hash: ...
```

---

### 2. Add a Transaction

```bash
go run ./cmd/toychain add -from FAUCET -to Alice -amount 100
```

This adds a transaction to the pending transaction pool.

Example output:

```text
Transaction added to pending pool: FAUCET -> Alice amount 100
Pending transactions: 1
```

---

### 3. View Pending Transactions

```bash
go run ./cmd/toychain pending
```

Example output:

```text
Pending transactions
0. FAUCET -> Alice : 100
```

---

### 4. Mine Pending Transactions

```bash
go run ./cmd/toychain mine
```

This mines the pending transaction into a new block.

Example output:

```text
Block mined successfully
Height: 1
Nonce: ...
Hash: ...
Hashes tried: ...
Time taken: ... ms
Remaining pending transactions: 0
```

---

### 5. Add Another Transaction

```bash
go run ./cmd/toychain add -from Alice -to Bob -amount 30
```

Then mine again:

```bash
go run ./cmd/toychain mine
```

---

### 6. Show Balances

```bash
go run ./cmd/toychain balances
```

Example output:

```text
Balances
Alice: 70
Bob: 30
FAUCET: -100
```

To include pending transactions in the balance view:

```bash
go run ./cmd/toychain balances -pending
```

---

### 7. Print the Blockchain

```bash
go run ./cmd/toychain print
```

This displays all blocks, hashes, previous hashes, nonce values, and transactions.

---

### 8. Validate the Blockchain

```bash
go run ./cmd/toychain validate
```

Expected output:

```text
Chain valid
```

This means no block data has been incorrectly changed.

---

### 9. Tamper with Old Data

```bash
go run ./cmd/toychain tamper -block 1 -tx 0 -amount 999
```

This deliberately changes an old transaction amount without recalculating the hash.

Example output:

```text
Tampered block 1 transaction 0 amount: 100 -> 999
Important: hash was not recalculated, so validation should fail now.
```

Now validate again:

```bash
go run ./cmd/toychain validate
```

Expected output:

```text
Chain invalid
First offending block: 1
Reason: ...
```

This demonstrates blockchain tamper detection.

---

## Full Demo Command Sequence

```bash
go run ./cmd/toychain init -difficulty 2

go run ./cmd/toychain add -from FAUCET -to Alice -amount 100

go run ./cmd/toychain pending

go run ./cmd/toychain mine

go run ./cmd/toychain add -from Alice -to Bob -amount 30

go run ./cmd/toychain mine

go run ./cmd/toychain balances

go run ./cmd/toychain print

go run ./cmd/toychain validate

go run ./cmd/toychain tamper -block 1 -tx 0 -amount 999

go run ./cmd/toychain validate
```

---

## How It Works

The application follows this simple flow:

```text
User enters CLI command
        ↓
main.go reads the command
        ↓
Blockchain is loaded from JSON file
        ↓
Requested operation is executed
        ↓
Blockchain is saved back to JSON file
        ↓
Result is printed in terminal
```

---

## Main Blockchain Concepts Demonstrated

### Block

A block stores blockchain data. In this project, each block contains:

* Height
* Timestamp
* Previous hash
* Transactions
* Nonce
* Current hash

---

### Transaction

A transaction represents value transfer between two accounts.

Example:

```text
Alice -> Bob : 50
```

---

### Genesis Block

The genesis block is the first block in the blockchain. It is created when the blockchain is initialized.

---

### Pending Transaction Pool

New transactions are first stored in a pending pool. They become confirmed only after mining.

---

### Mining

Mining creates a new block from pending transactions. The system changes the nonce repeatedly until a valid hash is found.

---

### Proof of Work

Proof of work makes mining computationally difficult. In this project, the difficulty controls how many starting zeros the block hash must contain.

Example for difficulty `2`:

```text
00abc123...
```

---

### Blockchain Validation

Validation checks whether:

* Block hashes are correct
* Previous hash links are correct
* Proof of work is valid
* Old data has not been changed

---

### Tamper Detection

If old data is changed, the stored hash no longer matches the block data. The `validate` command detects this and marks the chain as invalid.

---

## Testing

Format the code:

```bash
gofmt -w .
```

Run static checks:

```bash
go vet ./...
```

Run tests:

```bash
go test ./...
```

Example successful test result:

```text
ok      toy-blockchain/block
ok      toy-blockchain/chain
ok      toy-blockchain/storage
```

---

## Example Test Result

```text
ok      toy-blockchain/block    0.484s
ok      toy-blockchain/chain    0.910s
?       toy-blockchain/cmd/toychain     [no test files]
?       toy-blockchain/internal/transaction     [no test files]
?       toy-blockchain/ledger   [no test files]
ok      toy-blockchain/storage  1.420s
```

---

## Limitations

This is a learning project, not a real cryptocurrency system.

Current limitations:

* No peer-to-peer network
* No digital signatures
* No wallet generation
* No real user authentication
* No mining rewards
* No transaction fees
* No Merkle tree
* No database integration
* No web interface
* Not suitable for real financial transactions

---

## Future Improvements

Possible future improvements include:

* Add wallet creation
* Add public/private key cryptography
* Add digital signatures for transactions
* Add mining rewards
* Prevent spending more than available balance
* Add REST API
* Add web dashboard
* Add database storage
* Add peer-to-peer node communication
* Add Merkle root for transaction verification

---

## Project Status

Completed features:

* Blockchain initialization
* Genesis block creation
* Transaction creation
* Pending transaction pool
* Proof-of-work mining
* Block hash generation
* Previous hash linking
* Blockchain validation
* Balance calculation
* JSON file storage
* Tamper detection
* Unit testing

---

## Conclusion

This project successfully demonstrates the basic internal working of a blockchain using Go. It shows how transactions are added, how blocks are mined, how hashes connect blocks, and how validation detects tampering.

Although this is a simplified blockchain, it gives a clear foundation for understanding real blockchain systems.
