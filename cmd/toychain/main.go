package main

// Toychain CLI entrypoint and command handlers.

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"

	"toy-blockchain/chain"
	"toy-blockchain/internal/transaction"
	"toy-blockchain/storage"
)

const defaultDataFile = "data/chain.json"

type commonOptions struct {
	file       string
	difficulty int
	blockSize  int
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "init":
		err = runInit(args)
	case "genkey":
		err = runGenKey(args)
	case "add":
		err = runAdd(args)
	case "mine":
		err = runMine(args)
	case "print":
		err = runPrint(args)
	case "validate":
		err = runValidate(args)
	case "balances":
		err = runBalances(args)
	case "pending":
		err = runPending(args)
	case "tamper":
		err = runTamper(args)
	case "help", "-h", "--help":
		printUsage()
		return
	default:
		err = fmt.Errorf("unknown command: %s", cmd)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func addCommonFlags(fs *flag.FlagSet) *commonOptions {
	opts := &commonOptions{}
	fs.StringVar(&opts.file, "file", defaultDataFile, "path to blockchain JSON file")
	fs.IntVar(&opts.difficulty, "difficulty", chain.DefaultDifficulty, "proof-of-work difficulty for a new chain")
	fs.IntVar(&opts.blockSize, "block-size", chain.DefaultBlockSize, "maximum transactions per block for a new chain")
	return opts
}

// loadExisting attempts to load the chain from disk and returns a clear error
// if the file does not exist so callers can instruct the user to run `init`.
func loadExisting(opts *commonOptions) (*chain.Blockchain, error) {
	bc, err := storage.Load(opts.file)
	if err == nil {
		return bc, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("blockchain file missing: %s (run 'go run ./cmd/toychain init' first)", opts.file)
	}

	return nil, err
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	opts := addCommonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}

	bc := chain.NewBlockchain(opts.difficulty, opts.blockSize)
	if err := storage.Save(opts.file, bc); err != nil {
		return err
	}

	fmt.Println("New blockchain created")
	fmt.Println("File:", opts.file)
	fmt.Println("Difficulty:", bc.Difficulty)
	fmt.Println("Block size:", bc.BlockSize)
	fmt.Println("Genesis hash:", bc.Blocks[0].Hash)
	return nil
}

func runAdd(args []string) error {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	opts := addCommonFlags(fs)
	from := fs.String("from", "", "sender account")
	to := fs.String("to", "", "recipient account")
	amount := fs.Int("amount", 0, "transaction amount")
	sk := fs.String("sk", "", "sender private key in hex (32-byte seed or 64-byte key)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	bc, err := loadExisting(opts)
	if err != nil {
		return err
	}

	tx := transaction.New(*from, *to, *amount)
	if !tx.IsFaucet() {
		if *sk == "" {
			return fmt.Errorf("private key required for non-faucet transaction")
		}
		if err := tx.Sign(*sk); err != nil {
			return err
		}
	}

	if err := bc.AddTransaction(tx); err != nil {
		return err
	}

	if err := storage.Save(opts.file, bc); err != nil {
		return err
	}

	fmt.Printf("Transaction added to pending pool: %s -> %s amount %d\n", tx.From, tx.To, tx.Amount)
	fmt.Printf("Pending transactions: %d\n", len(bc.PendingTransactions))
	return nil
}

func runMine(args []string) error {
	fs := flag.NewFlagSet("mine", flag.ExitOnError)
	opts := addCommonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}

	bc, err := loadExisting(opts)
	if err != nil {
		return err
	}

	newBlock, result, err := bc.MinePending()
	if err != nil {
		return err
	}

	if err := storage.Save(opts.file, bc); err != nil {
		return err
	}

	fmt.Println("Block mined successfully")
	fmt.Println("Height:", newBlock.Height)
	fmt.Println("Nonce:", result.Nonce)
	fmt.Println("Hash:", result.Hash)
	fmt.Println("Hashes tried:", result.HashesTried)
	fmt.Println("Time taken:", result.DurationMillis, "ms")
	fmt.Println("Remaining pending transactions:", len(bc.PendingTransactions))
	return nil
}

func runPrint(args []string) error {
	fs := flag.NewFlagSet("print", flag.ExitOnError)
	opts := addCommonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}

	bc, err := loadExisting(opts)
	if err != nil {
		return err
	}

	printChain(bc)
	return nil
}

func printChain(bc *chain.Blockchain) {
	fmt.Println("--------Toy Blockchain CLI Project------")
	fmt.Println("Blockchain")
	fmt.Println("Difficulty:", bc.Difficulty)
	fmt.Println("Block size:", bc.BlockSize)
	fmt.Println("Blocks:", len(bc.Blocks))
	fmt.Println()

	for _, blk := range bc.Blocks {
		fmt.Println("----------------------------------------")
		fmt.Println("Height:", blk.Height)
		fmt.Println("Timestamp:", blk.Timestamp)
		fmt.Println("Previous hash:", blk.PrevHash)
		fmt.Println("Merkle root:", blk.MerkleRoot)
		fmt.Println("Nonce:", blk.Nonce)
		fmt.Println("Hash:", blk.Hash)
		fmt.Println("Transactions:")

		if len(blk.Transactions) == 0 {
			fmt.Println("  none")
		}

		for _, tx := range blk.Transactions {
			fmt.Printf("  %s -> %s : %d\n", tx.From, tx.To, tx.Amount)
		}
	}

	fmt.Println("----------------------------------------")
	fmt.Println()
}

func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	opts := addCommonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}

	bc, err := loadExisting(opts)
	if err != nil {
		return err
	}

	result := bc.Validate()
	if result.Valid {
		fmt.Println("Chain valid")
		return nil
	}

	fmt.Println("Chain invalid")
	fmt.Println("First offending block:", result.BlockHeight)
	fmt.Println("Reason:", result.Reason)
	return nil
}

func runBalances(args []string) error {
	fs := flag.NewFlagSet("balances", flag.ExitOnError)
	opts := addCommonFlags(fs)
	includePending := fs.Bool("pending", false, "include pending transactions in balance view")
	if err := fs.Parse(args); err != nil {
		return err
	}

	bc, err := loadExisting(opts)
	if err != nil {
		return err
	}

	printBalances(bc, *includePending)
	return nil
}

func printBalances(bc *chain.Blockchain, includePending bool) {
	balances := bc.Balances()
	if includePending {
		balances = bc.BalancesIncludingPending()
	}

	if len(balances) == 0 {
		fmt.Println("No balances yet")
		return
	}

	accounts := make([]string, 0, len(balances))
	for account := range balances {
		accounts = append(accounts, account)
	}
	sort.Strings(accounts)

	fmt.Println("Balances")
	for _, account := range accounts {
		fmt.Printf("%s: %d\n", account, balances[account])
	}
}

func runPending(args []string) error {
	fs := flag.NewFlagSet("pending", flag.ExitOnError)
	opts := addCommonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}

	bc, err := loadExisting(opts)
	if err != nil {
		return err
	}

	if len(bc.PendingTransactions) == 0 {
		fmt.Println("No pending transactions")
		return nil
	}

	fmt.Println("Pending transactions")
	for i, tx := range bc.PendingTransactions {
		fmt.Printf("%d. %s -> %s : %d\n", i, tx.From, tx.To, tx.Amount)
	}

	return nil
}

func runTamper(args []string) error {
	fs := flag.NewFlagSet("tamper", flag.ExitOnError)
	opts := addCommonFlags(fs)
	blockIndex := fs.Int("block", 1, "block height to tamper")
	txIndex := fs.Int("tx", 0, "transaction index inside the block")
	newAmount := fs.Int("amount", 999, "new amount to write without recalculating hash")
	if err := fs.Parse(args); err != nil {
		return err
	}

	bc, err := loadExisting(opts)
	if err != nil {
		return err
	}

	if *blockIndex <= 0 || *blockIndex >= len(bc.Blocks) {
		return fmt.Errorf("invalid block height %d", *blockIndex)
	}

	if *txIndex < 0 || *txIndex >= len(bc.Blocks[*blockIndex].Transactions) {
		return fmt.Errorf("invalid transaction index %d", *txIndex)
	}

	oldAmount := bc.Blocks[*blockIndex].Transactions[*txIndex].Amount
	bc.Blocks[*blockIndex].Transactions[*txIndex].Amount = *newAmount

	if err := storage.Save(opts.file, bc); err != nil {
		return err
	}

	fmt.Printf("Tampered block %d transaction %d amount: %d -> %d\n", *blockIndex, *txIndex, oldAmount, *newAmount)
	fmt.Println("Important: hash was not recalculated, so validation should fail now.")
	return nil
}

func runGenKey(args []string) error {
	fs := flag.NewFlagSet("genkey", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	pub, priv, err := generateEd25519Keypair()
	if err != nil {
		return err
	}

	fmt.Println("Private key (hex, 32-byte seed or 64-byte key):")
	fmt.Println(priv)
	fmt.Println()
	fmt.Println("Public key (hex):")
	fmt.Println(pub)
	return nil
}

func generateEd25519Keypair() (pubHex, privHex string, err error) {
	// Generate seed and keypair
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		return "", "", err
	}
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)
	privHex = hex.EncodeToString([]byte(priv))
	pubHex = hex.EncodeToString([]byte(pub))
	return pubHex, privHex, nil
}

func printUsage() {
	fmt.Println("--------Toy Blockchain CLI Project------")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init       Create a new blockchain file")
	fmt.Println("  add        Add a transaction to the pending pool")
	fmt.Println("  mine       Mine pending transactions into a new block")
	fmt.Println("  print      Print the blockchain")
	fmt.Println("  validate   Validate the blockchain")
	fmt.Println("  balances   Show account balances")
	fmt.Println("  pending    Show pending transactions")
	fmt.Println("  tamper     Deliberately modify old data for research experiment")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run ./cmd/toychain init -difficulty 2")
	fmt.Println("  go run ./cmd/toychain add -from FAUCET -to Alice -amount 100")
	fmt.Println("  go run ./cmd/toychain add -from <pubkey-hex> -to Bob -amount 30 -sk <private-key-hex>")
	fmt.Println("  go run ./cmd/toychain mine")
	fmt.Println("  go run ./cmd/toychain add -from Alice -to Bob -amount 30")
	fmt.Println("  go run ./cmd/toychain mine")
	fmt.Println("  go run ./cmd/toychain balances")
	fmt.Println("  go run ./cmd/toychain validate")
}
