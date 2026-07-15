# Toy Blockchain CLI

A small Go-based blockchain and ledger simulator that runs locally as a command-line application. The project demonstrates the core ideas behind blockchain design: blocks, transactions, a pending transaction pool, proof-of-work mining, Merkle roots, chain validation, JSON persistence, and optional Ed25519 transaction signatures.

## Features

- Create a new blockchain with a genesis block
- Add transactions to a pending transaction pool
- Mine pending transactions into new blocks
- Configure proof-of-work difficulty and block size
- Build a deterministic Merkle root from block transactions
- Store the Merkle root inside each block
- Print blocks, transactions, hashes, and Merkle roots
- Validate the full chain and detect tampering
- Show account balances and pending transactions
- Reject overspending transactions
- Persist the chain to JSON for later inspection
- Support optional Ed25519 signatures for non-faucet transactions

## Requirements

- Go 1.22 or later
- Git (optional)

Check the installed Go version:

```bash
go version
```

## Installation

```bash
git clone https://github.com/bipuliY/Toy-Blockchain-Project.git
cd Toy-Blockchain-Project
go mod tidy
```

## Running the CLI

Use the following command format:

```bash
go run ./cmd/toychain <command>
```

### Available commands

| Command | Description |
| --- | --- |
| `init` | Create a new blockchain |
| `genkey` | Generate a new Ed25519 keypair |
| `add` | Add a transaction to the pending pool |
| `mine` | Mine pending transactions into a new block |
| `print` | Print the blockchain |
| `validate` | Validate the blockchain |
| `balances` | Show account balances |
| `pending` | Show pending transactions |
| `tamper` | Deliberately modify an old transaction for testing |
| `help` | Show available commands |

> Note: non-faucet transactions require a private key. The CLI accepts a 32-byte seed or a 64-byte private key in hex using the `-sk` flag.

## Example workflow

Create a new blockchain:

```bash
go run ./cmd/toychain init -difficulty 2
```

Add an initial faucet transaction:

```bash
go run ./cmd/toychain add -from FAUCET -to Alice -amount 100
```

Mine the pending transaction:

```bash
go run ./cmd/toychain mine
```

Generate a keypair for a signed transfer:

```bash
go run ./cmd/toychain genkey
```

Add a signed transaction from the generated key to another account:

```bash
go run ./cmd/toychain add -from <pubkey-hex> -to Bob -amount 30 -sk <private-key-hex>
```

Mine again and inspect balances:

```bash
go run ./cmd/toychain mine
go run ./cmd/toychain balances
```

## Overspending protection

The blockchain rejects transactions when the sender does not have enough available balance. For example, a transfer larger than the current balance is rejected before it reaches the pending pool.

## Merkle root and hashing

Each block stores a deterministic Merkle root created from the block's transactions. The block hash is calculated from the height, timestamp, Merkle root, previous hash, and nonce. The raw transaction list is not directly included in the block hash; instead, the Merkle root acts as a compact summary of the transactions.

## Validation and tamper detection

Validation checks each block in order and verifies:

1. The block height is correct
2. The stored Merkle root matches the block transactions
3. The stored block hash matches a recalculated hash
4. The previous-hash link is correct
5. The block satisfies the required proof-of-work difficulty
6. Each transaction is valid and the balances remain consistent

A tampered block fails validation with a clear reason.

### Example tamper experiment

```bash
go run ./cmd/toychain tamper -block 1 -tx 0 -amount 999
go run ./cmd/toychain validate
```

Expected output:

```text
Chain invalid
First offending block: 1
Reason: stored Merkle root does not match block transactions
```

## Research highlights

The repository also includes a short research-style write-up in [report.md](report.md). The report documents two concrete experiments:

- Tamper evidence: changing a transaction in an already mined block causes validation to fail because the recalculated Merkle root no longer matches the stored one.
- Difficulty vs effort: mining was tested at several difficulties. The number of hashes tried grew much faster than linearly as difficulty increased.

### Example difficulty data

| Difficulty | Hashes tried | Time taken |
| --- | ---: | ---: |
| 2 | 663 | 0 ms |
| 3 | 2038 | 0 ms |
| 4 | 54472 | 18 ms |
| 5 | 2190554 | 713 ms |

## Testing

Format the Go files:

```bash
gofmt -w .
```

Run the complete test suite:

```bash
go test ./...
```

## Limitations

This is a learning-oriented toy blockchain and is not intended for production use. It does not provide peer-to-peer networking, full consensus among nodes, Merkle proofs for individual transactions, or a complete wallet and security model.

## Run the project - example

rm data/chain.json

go run ./cmd/toychain init -difficulty 2

go run ./cmd/toychain add -from FAUCET -to Alice -amount 100

go run ./cmd/toychain add -from FAUCET -to Bob -amount 50

go run ./cmd/toychain pending

go run ./cmd/toychain mine

go run ./cmd/toychain print

go run ./cmd/toychain validate

go run ./cmd/toychain tamper -block 1 -tx 0 -amount 999

go run ./cmd/toychain validate
