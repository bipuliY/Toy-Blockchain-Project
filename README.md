# Toy Blockchain CLI

A local blockchain and ledger simulator written in **Go** and operated through a command-line interface.

The project demonstrates blocks, transactions, a pending transaction pool, proof-of-work mining, SHA-256 hashing, Merkle roots, chain validation, concurrent mining, automatic difficulty retargeting, fork resolution, JSON persistence, balances, tamper detection, and optional Ed25519 transaction signatures.

> **Educational project:** This repository is intended for learning and experimentation. It is not a production blockchain or a real financial system.

---

## Features

* Create a blockchain with a deterministic genesis block
* Add transactions to a pending transaction pool
* Mine transactions into new blocks
* Limit transactions using a configurable block size
* Perform SHA-256 proof-of-work mining
* Mine concurrently using multiple goroutines
* Automatically use the available logical CPU count
* Record the nonce, hash, hashes tried, and mining time
* Create and store deterministic Merkle roots
* Link blocks using previous-block hashes
* Automatically retarget mining difficulty
* Validate the complete blockchain
* Detect changed transactions and broken block links
* Reject invalid and overspending transactions
* Show confirmed or pending-inclusive balances
* Generate Ed25519 keypairs
* Sign and verify non-faucet transactions
* Resolve competing chain files using the longest-valid-chain rule
* Save and load blockchain state as formatted JSON
* Use separate JSON files to simulate different local nodes

---

## Technology

* **Language:** Go 1.22+
* **Hashing:** SHA-256
* **Signatures:** Ed25519
* **Storage:** JSON files
* **Interface:** Command-line application
* **External Go dependencies:** None

---

## Project Structure

```text
Toy-Blockchain-Project/
│
├── block/
│   ├── block.go
│   └── block_test.go
├── chain/
│   ├── chain.go
│   ├── chain_test.go
│   ├── fork.go
│   ├── fork_test.go
│   └── retarget_test.go
├── cmd/
│   └── toychain/
│       └── main.go
├── internal/
│   └── transaction/
│       └── transaction.go
├── ledger/
│   └── ledger.go
├── merkle/
│   ├── merkle.go
│   └── merkle_test.go
├── storage/
│   ├── storage.go
│   └── storage_test.go
├── data/
│   └── chain.json
├── go.mod
├── README.md
└── report.md
```

`data/chain.json` is created when the default blockchain is initialized. Other JSON files can be used to simulate separate nodes.

---

## Default Configuration

| Setting                        |                     Default |
| ------------------------------ | --------------------------: |
| Blockchain file                |           `data/chain.json` |
| Initial difficulty             |                         `2` |
| Maximum transactions per block |                         `5` |
| Target block time              |                `10 seconds` |
| Retarget interval              |            `5 mined blocks` |
| Minimum difficulty             |                         `1` |
| Maximum difficulty             |                         `6` |
| Mining workers                 | Available logical CPU count |

The `init` command allows the initial difficulty and block size to be changed. The other values are currently constants in the chain package.

---

## Installation

```bash
git clone https://github.com/bipuliY/Toy-Blockchain-Project.git
cd Toy-Blockchain-Project
go mod tidy
go test ./...
```

Check your Go version:

```bash
go version
```

---

## Running the CLI

General format:

```bash
go run ./cmd/toychain <command> [flags]
```

Show help:

```bash
go run ./cmd/toychain help
```

Build an executable:

```bash
go build -o toychain ./cmd/toychain
```

Linux or macOS:

```bash
./toychain help
```

Windows PowerShell:

```powershell
.\toychain.exe help
```

---

## Available Commands

| Command    | Description                               |
| ---------- | ----------------------------------------- |
| `init`     | Create a new blockchain and genesis block |
| `genkey`   | Generate an Ed25519 keypair               |
| `add`      | Add a transaction to the pending pool     |
| `mine`     | Mine pending transactions into a block    |
| `print`    | Print the blockchain                      |
| `validate` | Validate the complete chain               |
| `resolve`  | Resolve competing blockchain files        |
| `balances` | Show account balances                     |
| `pending`  | Show pending transactions                 |
| `tamper`   | Deliberately change an old transaction    |
| `help`     | Show CLI usage                            |

---

## Command Reference

### Initialize a Blockchain

```bash
go run ./cmd/toychain init \
  -file data/chain.json \
  -difficulty 2 \
  -block-size 5
```

| Flag          |           Default | Description                          |
| ------------- | ----------------: | ------------------------------------ |
| `-file`       | `data/chain.json` | JSON file used to store the chain    |
| `-difficulty` |               `2` | Initial proof-of-work difficulty     |
| `-block-size` |               `5` | Maximum transactions per mined block |

The command prints the genesis hash and current blockchain settings.

### Generate a Keypair

```bash
go run ./cmd/toychain genkey
```

The command prints:

* A 64-byte Ed25519 private key in hexadecimal
* A 32-byte Ed25519 public key in hexadecimal

Keep the private key secret. Do not use educational test keys for real assets.

### Add a Faucet Transaction

```bash
go run ./cmd/toychain add \
  -from FAUCET \
  -to Alice \
  -amount 100
```

Faucet transactions create test funds and do not require a signature.

### Add a Signed Transaction

```bash
go run ./cmd/toychain add \
  -from <public-key-hex> \
  -to Bob \
  -amount 30 \
  -sk <private-key-hex>
```

The `-sk` flag accepts either:

* A 32-byte Ed25519 seed in hexadecimal
* A 64-byte Ed25519 private key in hexadecimal

When signing, the program derives the public key and uses it as the sender address.

### View Pending Transactions

```bash
go run ./cmd/toychain pending
```

### Mine Pending Transactions

```bash
go run ./cmd/toychain mine
```

Mining output includes:

* Block height
* Block difficulty
* Nonce
* Block hash
* Hashes tried
* Time taken
* Next-block difficulty
* Remaining pending transactions

Only the configured number of transactions is included in one block. Extra transactions remain pending.

### Print the Blockchain

```bash
go run ./cmd/toychain print
```

The output includes blockchain settings, block metadata, Merkle roots, hashes, nonces, and transactions.

### Show Balances

Confirmed balances only:

```bash
go run ./cmd/toychain balances
```

Include pending transactions:

```bash
go run ./cmd/toychain balances -pending
```

### Validate the Blockchain

```bash
go run ./cmd/toychain validate
```

Valid output:

```text
Chain valid
```

Invalid output:

```text
Chain invalid
First offending block: <height>
Reason: <failure reason>
```

### Tamper with a Block

```bash
go run ./cmd/toychain tamper \
  -block 1 \
  -tx 0 \
  -amount 999
```

| Flag      |           Default | Description                     |
| --------- | ----------------: | ------------------------------- |
| `-block`  |               `1` | Block height to modify          |
| `-tx`     |               `0` | Transaction index in that block |
| `-amount` |             `999` | New transaction amount          |
| `-file`   | `data/chain.json` | Chain file to modify            |

The command intentionally changes the transaction without recalculating the Merkle root or block hash.

### Resolve Competing Chains

```bash
go run ./cmd/toychain resolve \
  -file data/nodeA.json \
  -candidates data/nodeB.json,data/nodeC.json
```

A candidate replaces the local chain only when it is valid, compatible, and strictly longer.

---

## Correct Signed-Transaction Workflow

A generated key can only spend funds that belong to its public-key address. Sending faucet funds to the name `Alice` does not fund a generated public key.

### 1. Generate a Keypair

```bash
go run ./cmd/toychain genkey
```

Copy the printed public and private keys.

### 2. Initialize the Chain

```bash
go run ./cmd/toychain init -difficulty 2
```

### 3. Fund the Public-Key Address

```bash
go run ./cmd/toychain add \
  -from FAUCET \
  -to <public-key-hex> \
  -amount 100

go run ./cmd/toychain mine
```

### 4. Add and Mine a Signed Transfer

```bash
go run ./cmd/toychain add \
  -from <public-key-hex> \
  -to Bob \
  -amount 30 \
  -sk <private-key-hex>

go run ./cmd/toychain mine
```

### 5. Check the Result

```bash
go run ./cmd/toychain balances
go run ./cmd/toychain validate
```

Expected balance relationship:

```text
<public-key-hex>: 70
Bob: 30
```

---

## Simple Faucet Workflow

```bash
go run ./cmd/toychain init -difficulty 2

go run ./cmd/toychain add -from FAUCET -to Alice -amount 100
go run ./cmd/toychain add -from FAUCET -to Bob -amount 50

go run ./cmd/toychain pending
go run ./cmd/toychain mine
go run ./cmd/toychain print
go run ./cmd/toychain balances
go run ./cmd/toychain validate
```

---

## Block-Size Example

Create a chain that allows two transactions per block:

```bash
go run ./cmd/toychain init -difficulty 2 -block-size 2
```

Add three transactions:

```bash
go run ./cmd/toychain add -from FAUCET -to Alice -amount 10
go run ./cmd/toychain add -from FAUCET -to Bob -amount 20
go run ./cmd/toychain add -from FAUCET -to Carol -amount 30
```

Mine once:

```bash
go run ./cmd/toychain mine
```

The first two transactions are mined and one remains pending:

```bash
go run ./cmd/toychain pending
```

Mine again to include the remaining transaction:

```bash
go run ./cmd/toychain mine
```

---

## Concurrent Mining

The CLI calls the concurrent mining implementation.

When the internal worker count is `0`, the program uses:

```go
runtime.NumCPU()
```

Each goroutine searches a different sequence of nonce values. When one worker finds a hash that satisfies the difficulty, the other workers are cancelled.

---

## Difficulty Retargeting

Difficulty is reconsidered after every five mined blocks.

* If recent blocks were produced in less than half the expected duration, difficulty increases by `1`.
* If recent blocks took more than twice the expected duration, difficulty decreases by `1`.
* Otherwise, difficulty remains unchanged.
* Difficulty is always kept between `1` and `6`.

The next difficulty is stored in the blockchain JSON file and checked during validation.

---

## Fork Resolution

The `resolve` command simulates choosing between blockchain data received from other nodes.

Before a candidate is accepted, the program checks that it has:

* The same genesis block
* The same block size
* The same target block time
* The same retarget interval
* The same minimum and maximum difficulty
* The same initial difficulty
* A valid blockchain
* More blocks than the current best valid chain

Equal-length candidates do not replace the local chain.

Candidate pending transactions are not copied. Local pending transactions are checked again after adoption:

* Already confirmed transactions are dropped
* Invalid or overspending transactions are dropped
* Still-valid transactions are retained

### Example

Create node A:

```bash
go run ./cmd/toychain init -file data/nodeA.json -difficulty 2
go run ./cmd/toychain add -file data/nodeA.json -from FAUCET -to Alice -amount 100
go run ./cmd/toychain mine -file data/nodeA.json
```

Create a longer node B chain:

```bash
go run ./cmd/toychain init -file data/nodeB.json -difficulty 2

go run ./cmd/toychain add -file data/nodeB.json -from FAUCET -to Bob -amount 50
go run ./cmd/toychain mine -file data/nodeB.json

go run ./cmd/toychain add -file data/nodeB.json -from FAUCET -to Carol -amount 25
go run ./cmd/toychain mine -file data/nodeB.json
```

Resolve node A against node B:

```bash
go run ./cmd/toychain resolve \
  -file data/nodeA.json \
  -candidates data/nodeB.json
```

Validate the adopted chain:

```bash
go run ./cmd/toychain validate -file data/nodeA.json
```

---

## Transaction Rules

A transaction is rejected when:

* The sender is empty
* The recipient is empty
* The amount is zero or negative
* The sender and recipient are the same
* The public key format is invalid
* The signature format is invalid
* The sender does not match the public key
* Signature verification fails
* A non-faucet sender has insufficient balance

The available balance includes confirmed and already-pending transactions. This prevents multiple pending transactions from collectively overspending the same funds.

### Signature Compatibility Note

The CLI requires `-sk` for every non-faucet transaction.

At the lower package-validation level, the current code still permits a legacy unsigned non-faucet transaction when both signature fields are empty. This is a backward-compatibility path; the normal CLI workflow requires signing.

---

## Merkle Root

Each transaction is serialized to JSON and hashed with SHA-256.

Transaction hashes are combined in pairs until one root remains. When a level contains an odd number of hashes, the final hash is duplicated.

For an empty transaction list, such as the genesis block, the Merkle root is the SHA-256 hash of empty input.

A changed transaction therefore produces a different Merkle root.

---

## Block Hashing

Each block hash is calculated from:

1. Height
2. Timestamp
3. Merkle root
4. Previous hash
5. Difficulty
6. Nonce

The values are serialized to JSON and hashed using SHA-256.

The raw transaction list is not directly included in the block-hash input. The Merkle root acts as the transaction summary.

A SHA-256 hash is displayed as 64 hexadecimal characters.

---

## Genesis Block

The deterministic genesis block contains:

* Height `0`
* Timestamp `0`
* No transactions
* Merkle root calculated from empty input
* A previous hash containing 64 zeroes
* Difficulty `0`
* Nonce `0`

The genesis hash is calculated directly without proof-of-work mining.

---

## Validation Checks

Validation stops at the first detected problem and checks:

1. The chain contains blocks
2. Target block time is positive
3. Retarget interval is valid
4. Block size is positive
5. Difficulty limits are valid
6. Initial difficulty is allowed
7. Block heights are correct
8. Blocks do not exceed the size limit
9. Merkle roots match their transactions
10. Stored hashes match recalculated hashes
11. Genesis settings are correct
12. Previous-hash links are correct
13. Timestamps do not move backwards
14. Block difficulty matches the expected schedule
15. Proof-of-work requirements are satisfied
16. Transactions are valid
17. Balances remain consistent
18. Stored next-block difficulty is correct

---

## Tamper-Detection Experiment

```bash
go run ./cmd/toychain init -difficulty 2
go run ./cmd/toychain add -from FAUCET -to Alice -amount 100
go run ./cmd/toychain mine

go run ./cmd/toychain tamper -block 1 -tx 0 -amount 999
go run ./cmd/toychain validate
```

Expected result:

```text
Chain invalid
First offending block: 1
Reason: stored Merkle root does not match block transactions
```

The transaction changes, but the stored Merkle root remains unchanged. Recalculation therefore exposes the modification.

---

## JSON Persistence

The chain is stored as formatted JSON.

Saved data includes:

* Blocks
* Pending transactions
* Next-block difficulty
* Block size
* Target block time
* Retarget interval
* Minimum difficulty
* Maximum difficulty

When older JSON data does not contain newer configuration fields, the storage layer restores missing or invalid values using current defaults.

Use another file with `-file`:

```bash
go run ./cmd/toychain init -file data/test-chain.json
go run ./cmd/toychain add -file data/test-chain.json -from FAUCET -to Alice -amount 100
go run ./cmd/toychain mine -file data/test-chain.json
go run ./cmd/toychain print -file data/test-chain.json
```

---

## Resetting the Default Chain

Linux or macOS:

```bash
rm -f data/chain.json
```

Windows PowerShell:

```powershell
Remove-Item data/chain.json -ErrorAction SilentlyContinue
```

Initialize again:

```bash
go run ./cmd/toychain init -difficulty 2
```

---

## Testing

Format the code:

```bash
gofmt -w .
```

Run all tests:

```bash
go test ./...
```

Useful alternatives:

```bash
go test -count=1 ./...
go test -v ./...
go test -race ./...
```

The tests cover:

* Deterministic hashing
* Merkle roots
* Sequential and concurrent mining
* Proof-of-work difficulty
* Valid-chain verification
* Tamper detection
* Transaction rejection
* Overspending protection
* Block-size validation
* Difficulty retargeting
* Fork resolution
* JSON persistence

---

## Mining Difficulty Experiment

Mining displays:

```text
Hashes tried: ...
Time taken: ... ms
```

Run the same experiment at several initial difficulties:

```bash
go run ./cmd/toychain init -difficulty <N>
go run ./cmd/toychain add -from FAUCET -to Alice -amount 100
go run ./cmd/toychain mine
```

Record the results:

| Difficulty | Hashes tried | Time taken |
| ---------: | -----------: | ---------: |
|          1 |              |            |
|          2 |              |            |
|          3 |              |            |
|          4 |              |            |
|          5 |              |            |

Mining is probabilistic, so results differ between runs.

---

## Research Report

More detailed design and research discussion is available in:

```text
report.md
```

It covers architecture, hashing, Merkle roots, mining, validation, tamper evidence, concurrent mining, difficulty retargeting, fork resolution, signatures, testing, and limitations.

---

## Limitations

* No peer-to-peer network
* No communication between real independent nodes
* No distributed consensus protocol
* Fork resolution uses local JSON files
* Fork selection uses block count rather than cumulative proof of work
* No Merkle proofs for individual transactions
* No smart contracts
* No transaction fees or mining rewards
* No wallet application or secure key storage
* No database or replicated storage
* Difficulty retargeting uses simple Unix-second timestamps
* No CLI flag for selecting the worker count
* Legacy unsigned transactions remain possible through lower-level package use
* Faucet transactions can create unlimited test funds

---

## Possible Improvements

* Strictly require signatures at every validation layer
* Add a mining worker-count flag
* Select forks using cumulative proof of work
* Add Merkle proofs
* Add fees and mining rewards
* Add wallet and key-management support
* Add CLI integration tests and benchmarks
* Add networking and distributed consensus
* Store blockchain data in a database
* Make retargeting settings configurable
* Add structured logging

---

## Full Demonstration

```bash
rm -f data/chain.json

go run ./cmd/toychain init -difficulty 2 -block-size 5

go run ./cmd/toychain add -from FAUCET -to Alice -amount 100
go run ./cmd/toychain add -from FAUCET -to Bob -amount 50

go run ./cmd/toychain pending
go run ./cmd/toychain mine
go run ./cmd/toychain print
go run ./cmd/toychain balances
go run ./cmd/toychain validate

go run ./cmd/toychain tamper -block 1 -tx 0 -amount 999
go run ./cmd/toychain validate
```

---

## Disclaimer

This repository is an educational blockchain simulator. It must not be used as a cryptocurrency, production ledger, secure wallet, or real financial system.
