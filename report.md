Toy Blockchain CLI Project Report
Project Title

Toy Blockchain Command Line Application Using Go

Abstract

This project is a simple blockchain command line application developed using the Go programming language. The main purpose of this project is to demonstrate the basic internal working concepts of a blockchain system in a simple and understandable way.

The application allows users to create a blockchain, add transactions, mine blocks, view blockchain data, check account balances, validate the blockchain, view pending transactions, and deliberately tamper with old data to test blockchain security. The blockchain data is stored in a JSON file, allowing the data to persist even after the program stops.

This project does not aim to build a real cryptocurrency system. Instead, it focuses on the core educational concepts of blockchain, such as blocks, transactions, hashing, proof of work, previous hash linking, chain validation, and tamper detection.

1. Introduction

Blockchain is a distributed data storage technology where data is stored in blocks. Each block is connected to the previous block using a cryptographic hash. Because of this hash connection, changing data in an old block affects all following blocks. This makes blockchain suitable for systems that require data integrity, transparency, and tamper detection.

This project implements a simplified blockchain system as a command line interface. The user interacts with the blockchain through terminal commands. The system supports adding transactions, mining transactions into blocks, viewing balances, validating the blockchain, and testing tampering.

The project is implemented using Go because Go is simple, fast, strongly typed, and suitable for building command line tools and backend systems.

2. Problem Statement

The main problem addressed by this project is understanding how blockchain works internally. Many blockchain systems are complex because they include peer-to-peer networking, cryptographic wallets, smart contracts, distributed consensus, and advanced security mechanisms.

For learning purposes, this project simplifies the blockchain concept and focuses only on the most important internal mechanisms:

    How transactions are created.
    How transactions are stored before mining.
    How a block is created.
    How proof of work is performed.
    How blocks are connected using hashes.
    How blockchain validation detects tampering.
    How account balances can be calculated from transactions.
    
3. Project Objectives

The main objectives of this project are:

    To implement a simple blockchain using Go.
    To create a command line interface for interacting with the blockchain.
    To add transactions into a pending transaction pool.
    To mine pending transactions into new blocks.
    To use proof of work during block mining.
    To store blockchain data in a JSON file.
    To validate the blockchain after each operation.
    To calculate account balances from confirmed transactions.
    To deliberately tamper with blockchain data and observe validation failure.
    To understand how blockchain protects data integrity.

4. Scope of the Project

This project includes the basic blockchain operations required for educational demonstration.

Included in the Scope
    Blockchain initialization.
    Genesis block creation.
    Transaction creation.
    Pending transaction pool.
    Block mining.
    Proof-of-work difficulty.
    Blockchain printing.
    Blockchain validation.
    Balance calculation.
    Pending transaction viewing.
    Tamper testing.
    JSON-based data storage.
    Unit testing for important packages.
    Not Included in the Scope

This project does not include:

    Real cryptocurrency wallets.
    Digital signatures.
    Peer-to-peer network.
    Distributed mining.
    Smart contracts.
    Real financial transactions.
    Merkle trees.
    User authentication.
    Web interface.
    Database integration.

Therefore, this project should be considered a learning-based toy blockchain, not a production blockchain.
5. Project Folder Structure

The project is organized into separate packages. This improves readability and maintainability.

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


6. System Overview

The system works as a command line application. The user enters a command in the terminal. The main.go file reads the command and calls the correct function.

For example:

    go run ./cmd/toychain add -from Alice -to Bob -amount 50

The system then:

    Reads the command.
    Loads the blockchain from data/chain.json.
    Creates a transaction.
    Adds it to the pending transaction pool.
    Saves the updated blockchain.
    Prints a success message.

The blockchain is not stored only in memory. It is saved into a JSON file. Therefore, the chain can be reused later.