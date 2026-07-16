# Optional Blockchain Features Report

## 1. Introduction

This report explains the optional blockchain features implemented in the **Toy Blockchain CLI** project. The project is written in Go and extends the basic blockchain implementation with four advanced features:

1. Digital signatures
2. Merkle roots
3. Concurrent mining
4. Automatic difficulty retargeting

These features improve transaction authenticity, block integrity, mining performance, and proof-of-work control.

---

## 2. Feature Completion Summary

| Optional Feature | Status | Main Implementation Files |
| ---              | ---    | ---                       |
| Digital signatures | Implemented | `internal/transaction/transaction.go`, `cmd/toychain/main.go`, `ledger/ledger.go` |
| Merkle root | Implemented | `merkle/merkle.go`, `block/block.go`, `chain/chain.go` |
| Concurrent mining | Implemented | `block/block.go`, `chain/chain.go` |
| Difficulty retargeting | Implemented | `chain/chain.go`, `block/block.go`, `cmd/toychain/main.go` |

---

## 3. Digital Signatures

### 3.1 Purpose

A digital signature proves that a transaction was authorised by the owner of a private key. It also helps detect changes made to the signed transaction after signing.

The project uses the **Ed25519** digital-signature algorithm provided by Go's `crypto/ed25519` package.

### 3.2 Key-pair generation

The CLI provides the following command:

```bash
go run ./cmd/toychain genkey
```

This command generates:

- A private key used to sign transactions
- A public key used as the sender identity and to verify signatures

The keys are printed in hexadecimal format.

### 3.3 Transaction signing

A non-faucet transaction is created using a private key:

```bash
go run ./cmd/toychain add   -from <public-key-hex>   -to Bob   -amount 30   -sk <private-key-hex>
```

The `Sign()` method:

1. Decodes the private key from hexadecimal.
2. Accepts either a 32-byte seed or a 64-byte Ed25519 private key.
3. Derives the public key from the private key.
4. Creates deterministic signing data from `From`, `To`, and `Amount`.
5. Signs the data using the private key.
6. Stores the public key and signature in the transaction.
7. Sets the sender address to the public-key value.

The transaction structure stores:

```go
PubKeyHex string
SigHex    string
```

### 3.4 Signature verification

Before a signed transaction is accepted, validation:

1. Decodes the public key.
2. Decodes the signature.
3. Checks that the sender matches the supplied public key.
4. Recreates the original signing data.
5. Calls `ed25519.Verify()`.

The transaction is rejected when the key format is invalid, the sender does not match the key, or the signature verification fails.

The CLI requires a private key for every non-faucet transaction. However, the lower-level transaction validator retains support for older unsigned transactions when both signature fields are empty. This backward-compatibility behaviour should be considered a limitation if strict signature enforcement is required outside the CLI.

### 3.5 Benefits

Digital signatures provide:

- Transaction authentication
- Sender ownership verification
- Protection against unauthorised transaction creation
- Detection of signed-data modification

---

## 4. Merkle Root

### 4.1 Purpose

A Merkle root is one hash that summarises all transactions in a block. Instead of hashing the complete raw transaction list directly inside the block hash, the project calculates a Merkle root and includes it in the block-hash input.

Each block stores:

```go
Transactions []transaction.Transaction
MerkleRoot   string
```

### 4.2 Merkle-root calculation

The `merkle.CalculateRoot()` function performs the following steps:

1. Serialises each transaction using JSON.
2. Hashes every transaction using SHA-256.
3. Treats the transaction hashes as the leaf nodes of a Merkle tree.
4. Combines adjacent hashes and hashes them again.
5. Duplicates the final hash when a tree level contains an odd number of hashes.
6. Repeats the process until one root hash remains.

For an empty transaction list, such as the genesis block, the project returns the SHA-256 hash of empty bytes. This creates a deterministic Merkle root.

### 4.3 Use in block hashing

The block hash is calculated from:

- Block height
- Timestamp
- Merkle root
- Previous-block hash
- Difficulty
- Nonce

The raw transaction list is not directly included in the final block-hash input.

### 4.4 Validation

During chain validation, the Merkle root is recalculated from the stored transactions. Validation fails when the recalculated value differs from the stored value.

Example validation error:

```text
stored Merkle root does not match block transactions
```

Therefore, changing a transaction after mining changes its transaction hash and Merkle root, allowing tampering to be detected.

### 4.5 Benefits

The Merkle-root implementation provides:

- A compact summary of all block transactions
- Tamper detection
- Deterministic transaction integrity checking
- A foundation for future Merkle-proof support

The current project calculates and validates Merkle roots but does not yet generate individual Merkle proofs.

---

## 5. Concurrent Mining

### 5.1 Purpose

Proof-of-work mining searches for a nonce that produces a block hash with the required number of leading zeroes. Concurrent mining divides this search across multiple goroutines so that different nonce values can be checked at the same time.

### 5.2 Goroutine-based design

The `MineConcurrent()` method receives:

```go
func (b *Block) MineConcurrent(difficulty int, workerCount int) MineResult
```

When `workerCount` is zero or negative, the function uses:

```go
runtime.NumCPU()
```

This allows the program to use the available logical CPU count.

Each goroutine starts from a different nonce:

```text
Worker 0: 0, workerCount, 2 × workerCount, ...
Worker 1: 1, workerCount + 1, ...
Worker 2: 2, workerCount + 2, ...
```

This prevents workers from repeatedly testing the same nonce values.

### 5.3 Clean cancellation

The implementation uses:

- `context.WithCancel()`
- A buffered result channel
- `sync.WaitGroup`
- Atomic hash-attempt counting

When one goroutine finds a valid hash:

1. It sends the nonce and hash to the result channel.
2. It calls the cancellation function.
3. The remaining goroutines detect the cancelled context and stop.
4. The main function waits for all workers to exit.
5. The winning nonce and hash are saved in the block.

This prevents unnecessary goroutines from continuing after a result is found.

### 5.4 Integration with block mining

`Blockchain.MinePending()` calls:

```go
newBlock.MineConcurrent(bc.Difficulty, 0)
```

Passing `0` selects the default CPU-based worker count. Therefore, normal CLI mining uses concurrent mining automatically.

### 5.5 Recorded mining results

The mining method returns:

- Winning nonce
- Valid block hash
- Mining duration in milliseconds
- Total number of hash attempts

The CLI prints these values after mining.

### 5.6 Benefits

Concurrent mining provides:

- Parallel nonce searching
- Better use of multicore processors
- Clean shutdown of unsuccessful workers
- Useful mining-performance statistics

Mining speed is still affected by CPU availability, operating-system scheduling, and the selected difficulty.

---

## 6. Difficulty Retargeting

### 6.1 Purpose

A fixed proof-of-work difficulty can produce blocks too quickly or too slowly when computing performance changes. Difficulty retargeting adjusts the difficulty so that block production remains closer to a target time.

### 6.2 Default settings

The blockchain stores the following retargeting configuration:

```go
DefaultTargetBlockTimeSeconds = 10
DefaultRetargetInterval       = 5
DefaultMinDifficulty          = 1
DefaultMaxDifficulty          = 6
```

This means:

- Target block time: 10 seconds
- Retarget check: every 5 mined blocks
- Minimum difficulty: 1
- Maximum difficulty: 6

Each mined block also stores the difficulty used to mine that specific block.

### 6.3 Retargeting process

After a new block is mined and added to the chain, the blockchain calls:

```go
bc.Difficulty = bc.CalculateNextDifficulty()
```

The retargeting function:

1. Waits until enough blocks have been mined.
2. Runs only at the configured retarget interval.
3. Calculates the actual time taken by the recent block window.
4. Calculates the expected time using the target block time.
5. Compares the actual duration with the expected duration.

The rules are:

- If blocks were produced in less than half the expected time, increase difficulty by 1.
- If blocks took more than twice the expected time, decrease difficulty by 1.
- Otherwise, keep the current difficulty unchanged.

The result is clamped between the configured minimum and maximum difficulty.

### 6.4 Validation of historical difficulty

Chain validation does not only check whether a block hash meets its stored difficulty. It also recalculates the expected difficulty at every retarget boundary.

Validation fails when:

- A block stores an unexpected difficulty.
- A block hash does not satisfy its recorded difficulty.
- The blockchain's stored next difficulty is incorrect.
- The minimum, maximum, interval, or target-time settings are invalid.

This prevents a saved blockchain from silently changing its proof-of-work rules.

### 6.5 CLI output

The `init` command displays:

- Starting difficulty
- Target block time
- Retarget interval
- Allowed difficulty range

After mining, the CLI displays either:

```text
Difficulty retargeted: old -> new
```

or:

```text
Next block difficulty: value
```

### 6.6 Benefits

Difficulty retargeting provides:

- Automatic proof-of-work adjustment
- More stable target block times
- Protection against unlimited difficulty growth
- Historical validation of difficulty changes

Because this is a local toy blockchain and timestamps use whole seconds, very fast blocks may have identical timestamps. The retargeting algorithm is therefore educational rather than production-grade.

---

## 7. Combined Operation

The four optional features work together in the following sequence:

1. A user generates an Ed25519 key pair.
2. A non-faucet transaction is signed with the private key.
3. The transaction signature is checked before the transaction is accepted.
4. Pending transactions are selected for a new block.
5. A Merkle root is calculated from the selected transactions.
6. The Merkle root, previous hash, difficulty, and nonce are used to calculate the block hash.
7. Multiple goroutines search different parts of the nonce space.
8. The first valid result stops all remaining workers.
9. The block is appended to the blockchain.
10. The difficulty for the next block is recalculated when a retarget boundary is reached.
11. Full-chain validation checks signatures, balances, Merkle roots, hashes, links, proof of work, and expected difficulty.

---

## 8. Testing and Verification

The repository includes tests confirming important parts of the optional implementation, including:

- Deterministic genesis-block Merkle roots
- Merkle-root creation for new blocks
- Block-hash changes when the Merkle root changes
- Sequential mining difficulty satisfaction
- Concurrent mining difficulty satisfaction
- Correct storage of the winning concurrent nonce
- Positive concurrent hash-attempt counts
- Automatic selection of the default worker count
- Detection of transaction tampering through a Merkle-root mismatch
- Successful validation of an honest blockchain

The complete test suite can be run with:

```bash
go test ./...
```

The source can be formatted before testing with:

```bash
gofmt -w .
```

---

## 9. Limitations and Future Improvements

Although all four optional features are implemented, the project remains a learning-oriented blockchain simulator.

Possible future improvements include:

- Enforcing signatures for all non-faucet transactions at every API level
- Encrypting or securely storing private keys
- Adding wallet files instead of printing private keys directly
- Adding individual Merkle proofs
- Benchmarking sequential mining against concurrent mining
- Allowing the worker count to be configured through a CLI flag
- Using more robust timestamp and retargeting rules
- Adding dedicated automated tests for signature failures and difficulty retarget boundaries
- Adding peer-to-peer networking and distributed consensus

---

## 10. Conclusion

The Toy Blockchain CLI successfully implements all four optional features.

Digital signatures authenticate non-faucet transactions using Ed25519 keys. Merkle roots summarise block transactions and support tamper detection. Concurrent mining divides nonce searching across goroutines and stops workers cleanly after a valid hash is found. Difficulty retargeting adjusts the proof-of-work difficulty according to recent block-production times and verifies those changes during full-chain validation.

Together, these additions make the project more advanced than a basic blockchain demonstration while keeping the implementation understandable for learning and experimentation.