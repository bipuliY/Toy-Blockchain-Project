# Toy Blockchain and Ledger Simulator

## Research Report

## 1. Introduction

This project is a small blockchain and ledger simulator developed in Go. It runs locally as a command-line application and does not communicate with external blockchain networks.

The system supports transactions, account balances, a pending transaction pool, proof-of-work mining, JSON persistence, full-chain validation, and deliberate tampering for testing. As an additional stretch goal, a Merkle root was implemented to summarise the transactions contained in each block.

The purpose of this report is to describe the hashing and validation design, present experiments conducted using the implementation, and discuss how the toy blockchain differs from a production blockchain.

---

## 2. Implementation and Design

### 2.1 Block structure

Each block contains:

1. Height
2. Unix timestamp
3. List of transactions
4. Merkle root
5. Previous block hash
6. Nonce
7. Block hash

The first block is a deterministic genesis block at height `0`. Its previous hash is a fixed string containing 64 zero characters.

Each later block stores the hash of the block before it. This creates a link between consecutive blocks.

### 2.2 Transaction model

A transaction contains:

* sender,
* recipient,
* amount,
* optional public key, and
* optional digital signature.

The ledger rejects transactions with:

* an empty sender,
* an empty recipient,
* a non-positive amount,
* the same sender and recipient, or
* an amount larger than the sender's available balance.

A special sender named `FAUCET` is used to introduce initial funds. Faucet transactions credit the recipient without requiring an existing sender balance.

### 2.3 Pending transactions and block size

New transactions are first stored in a pending transaction pool.

When mining begins, the application selects transactions from this pool up to the configured maximum block size. The selected transactions are copied into a new block and removed from the pending pool only after mining succeeds.

---

## 3. Hashing Scheme

### 3.1 Transaction hashing

Each transaction is serialised using Go's `encoding/json` package and hashed using SHA-256.

Conceptually:

```text
Transaction hash = SHA-256(serialised transaction)
```

The same transaction data produces the same transaction hash. Changing the sender, recipient, amount, public key, or signature changes the resulting hash.

### 3.2 Merkle-root construction

The hash of every transaction becomes a leaf in a Merkle tree.

For four transactions:

```text
TX1 → Hash A
TX2 → Hash B
TX3 → Hash C
TX4 → Hash D
```

Adjacent hashes are joined and hashed:

```text
Hash AB = SHA-256(Hash A + Hash B)
Hash CD = SHA-256(Hash C + Hash D)
```

The parent hashes are then joined and hashed again:

```text
Merkle root = SHA-256(Hash AB + Hash CD)
```

The resulting structure is:

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

When a level contains an odd number of hashes, the final hash is duplicated. For example:

```text
A, B, C
```

is processed as:

```text
A, B, C, C
```

The genesis block has no transactions. Its Merkle root is therefore the SHA-256 hash of an empty byte sequence:

```text
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
```

### 3.3 Block hashing

The block hash is calculated from a stable JSON structure containing the following fields in this order:

1. Height
2. Timestamp
3. Merkle root
4. Previous hash
5. Nonce

Conceptually:

```text
Block hash = SHA-256(
    height +
    timestamp +
    Merkle root +
    previous hash +
    nonce
)
```

The raw transaction list is not directly included in the block-hash input. It is represented by the Merkle root.

The block's own hash is excluded because including it would create a circular dependency.

Go's standard `crypto/sha256` implementation was used because it is deterministic, widely standardised, available without external dependencies, and suitable for demonstrating blockchain hashing.

---

## 4. Proof-of-Work Mining

Mining searches for a nonce that causes the block hash to begin with the configured number of zero hexadecimal characters.

The process is:

```text
1. Start with nonce 0.
2. Calculate the block hash.
3. Check whether the hash begins with enough zeroes.
4. If not, increase the nonce.
5. Repeat until a valid hash is found.
```

For difficulty `2`, a valid result may look like:

```text
007143e8d37312ffd3d5942c0a96910f93c0f18b4d0fec5c29a855f2262090a6
```

Before mining starts, the Merkle root is recalculated from the block's transactions. Therefore, the proof of work is based on the correct transaction summary.

---

## 5. Full-Chain Validation

Validation checks every block from the genesis block to the latest block.

For each block, the application verifies:

1. The height matches the block's position.
2. The Merkle root matches the stored transactions.
3. The stored block hash matches a recalculated hash.
4. The previous-hash link is correct.
5. Timestamps do not move backwards.
6. The block hash satisfies the configured proof-of-work difficulty.
7. Every transaction is valid.
8. The sender has enough balance.
9. A supplied digital signature is valid.

The Merkle-root check is performed before the block-hash check:

```text
Transactions
     ↓
Recalculate Merkle root
     ↓
Compare with stored Merkle root
     ↓
Recalculate block hash
     ↓
Compare with stored block hash
```

This is necessary because the block hash contains the stored Merkle root rather than the raw transactions.

---

## 6. Investigation 1: Tamper Evidence

### 6.1 Objective

The objective was to determine whether changing a transaction inside an already-mined block would be detected during validation.

### 6.2 Procedure

A new chain was created and two transactions were added:

```bash
rm data/chain.json
go run ./cmd/toychain init -difficulty 2
go run ./cmd/toychain add -from FAUCET -to Alice -amount 100
go run ./cmd/toychain add -from FAUCET -to Bob -amount 50
go run ./cmd/toychain mine
```

The honest chain was then validated:

```text
$ go run ./cmd/toychain validate
Chain valid
```

The first transaction in block `1` was changed from `100` to `999`:

```text
$ go run ./cmd/toychain tamper -block 1 -tx 0 -amount 999

Tampered block 1 transaction 0 amount: 100 -> 999
Important: hash was not recalculated, so validation should fail now.
```

Validation was run again:

```text
$ go run ./cmd/toychain validate

Chain invalid
First offending block: 1
Reason: stored Merkle root does not match block transactions
```

### 6.3 Explanation

Before tampering, the Merkle root represented:

```text
FAUCET → Alice : 100
FAUCET → Bob   : 50
```

After tampering, the first transaction became:

```text
FAUCET → Alice : 999
```

This changed the serialised transaction data and therefore changed its SHA-256 transaction hash. Because one leaf hash changed, the calculated Merkle root also changed.

However, the Merkle root stored in the block remained the original value.

Therefore:

```text
Stored Merkle root ≠ Recalculated Merkle root
```

The Merkle-root validation check detected the modification and identified block `1` as the first offending block.

### 6.4 Result

The experiment confirmed that changing even one transaction makes the chain invalid.

The automated test `TestValidationDetectsMerkleRootMismatch` also passed, proving that the behaviour is checked whenever the Go test suite runs.

---

## 7. Investigation 2: Difficulty Versus Mining Effort

### 7.1 Objective

The objective was to observe how increasing proof-of-work difficulty affects:

* the number of hashes attempted, and
* the mining time.

### 7.2 Results

The following results were obtained during mining experiments:

| Difficulty | Hashes tried | Approximate time |
| ---------: | -----------: | ---------------: |
|          4 |      119,724 |     0.08 seconds |
|          5 |      371,601 |     0.16 seconds |
|          6 |   44,060,072 |    14.77 seconds |

### 7.3 Interpretation

The mining effort does not grow linearly.

Each hexadecimal character has 16 possible values:

```text
0, 1, 2, ..., 9, a, b, c, d, e, f
```

The approximate probability that one hash begins with one required zero is:

```text
1/16
```

For two required zeroes, it is:

```text
1/16²
```

For a difficulty of `d`, the expected work grows approximately as:

```text
16^d
```

Therefore, adding one more required zero can make the expected search approximately 16 times harder.

Actual results vary because mining is probabilistic. A valid nonce may sometimes be found quickly, while another block at the same difficulty may require many more attempts.

The large increase at difficulty `6` demonstrates this probabilistic and faster-than-linear growth.

---

## 8. Discussion Questions

### 8.1 How does the previous-hash link make old-block tampering impractical in a real chain?

Each block stores the hash of the block before it.

If an old block is changed, its Merkle root and block hash must also change. The following block still contains the original hash as its `previous hash`, so the link becomes invalid.

To hide the modification, an attacker would need to:

1. recalculate the modified block,
2. redo its proof of work,
3. update and re-mine every following block, and
4. catch up with and surpass the valid chain maintained by the network.

In this local toy blockchain, tampering is easy because there is only one local JSON file and no competing nodes. In a production proof-of-work blockchain, honest nodes continue extending the valid chain while the attacker tries to recreate the old work. This makes rewriting a sufficiently deep block computationally difficult. This principle is described in the Bitcoin white paper's proof-of-work discussion.

### 8.2 What is an alternative to proof of work?

One alternative is **proof of stake**.

In proof of stake, validators are selected partly according to cryptocurrency they lock as stake rather than by repeatedly calculating hashes.

One advantage is that it normally requires much less computational work and energy than proof of work.

One drawback is that participants with larger stakes may have greater influence. The system must also include rules for penalising dishonest validators and handling validator selection fairly.

### 8.3 Three differences from a production blockchain

This toy blockchain differs from a production blockchain in several ways:

1. It has no peer-to-peer network or distributed consensus.
2. Legacy unsigned transactions are accepted.
3. It does not support Merkle proofs for individual transactions.
4. It stores the complete chain in one local JSON file.
5. It has no mining rewards or transaction fees.
6. It has no fork-resolution or finality rules.
7. It does not provide secure wallet or private-key management.

### 8.4 Improvement sketch: Merkle proofs

The project now calculates a Merkle root, but it does not generate Merkle proofs.

A Merkle proof allows a transaction to be checked without downloading or hashing every transaction in the block.

To add this feature, the implementation could:

1. store or reconstruct each level of the Merkle tree,
2. locate the requested transaction's leaf hash,
3. collect the neighbouring hash at each level,
4. return these hashes as the proof, and
5. reconstruct the root using the transaction and proof.

The reconstructed root would then be compared with the Merkle root stored in the block.

This would make it possible to prove that a transaction belongs to a block using only a small number of hashes.

---

## 9. Testing and Reproducibility

The source code was formatted using:

```bash
gofmt -w .
```

The complete test suite was run using:

```bash
go test ./...
```

The tests passed for the block, chain, Merkle, and storage packages.

Detailed chain tests included:

```text
TestHonestChainValidates
TestTamperingIsDetected
TestOverspendingTransactionIsRejected
TestNegativeAmountIsRejected
TestValidationDetectsMerkleRootMismatch
```

The project can be reproduced from a fresh clone using the build and run commands documented in `README.md`.

---

## 10. Limitations and Engineering Decisions

The project was intentionally kept small and understandable.

Important limitations include:

* no distributed network,
* no consensus between independent peers,
* no fork handling,
* no transaction fees,
* no mining rewards,
* no Merkle proofs,
* optional rather than compulsory signatures,
* local JSON persistence, and
* no production-grade key management.

The standard library was preferred throughout the implementation. SHA-256, JSON encoding, Ed25519 signing, file operations, and timing were implemented using Go standard-library packages.

The Merkle root and optional signatures were treated as stretch improvements only after the main blockchain, ledger, mining, validation, persistence, and tests were working.

---

## 11. Conclusion

The project successfully demonstrates the required blockchain and ledger concepts in a compact Go command-line application.

The tampering investigation showed that modifying a transaction changes its transaction hash and the recalculated Merkle root. Validation correctly detected the mismatch and identified the first offending block.

The mining investigation showed that difficulty and mining effort do not have a linear relationship. Requiring additional leading hexadecimal zeroes reduces the probability of success and causes expected work to grow approximately as `16^d`.

The Merkle-root stretch goal improved the original design by replacing the raw transaction list in the block-hash input with one deterministic transaction summary. Together with proof-of-work, previous-hash linking, ledger validation, and automated tests, this creates a clear educational example of a tamper-evident blockchain.

---

## References

1. Satoshi Nakamoto, *Bitcoin: A Peer-to-Peer Electronic Cash System*, 2008.
2. Go Documentation, *crypto/sha256 Package*.
3. Go Documentation, *encoding/json Package*.
4. Go Documentation, *crypto/ed25519 Package*.
5. Ethereum.org, *Proof-of-Stake Documentation*.
