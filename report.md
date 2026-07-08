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