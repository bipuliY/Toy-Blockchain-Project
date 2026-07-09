package main
//imported packages
import (
	"errors"   //use to check file loading errors
	"flag"    // use to parse command line flags like -from , -to , -amount
	"fmt"   
	"os" 	//use to read terminal arguments and exit the program
	"sort"	//use to sort acc names before printing balances

	"toy-blockchain/chain"  //Handles blockchain, blocks, mining, validation, balances
	"toy-blockchain/internal/transaction" //Creates transaction objects
	"toy-blockchain/storage"  //Saves and loads blockchain data from JSON file
)

const defaultDataFile = "data/chain.json" //changes savee dto this path

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

	command := os.Args[1]
	args := os.Args[2:]

	var err error

	switch command {
	case "init":
		err = runInit(args)
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
		err = fmt.Errorf("unknown command: %s", command)
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

func loadOrCreate(opts *commonOptions) (*chain.Blockchain, error) {
	bc, err := storage.Load(opts.file)
	if err == nil {
		return bc, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return chain.NewBlockchain(opts.difficulty, opts.blockSize), nil
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
	if err := fs.Parse(args); err != nil {
		return err
	}

	bc, err := loadOrCreate(opts)
	if err != nil {
		return err
	}

	tx := transaction.New(*from, *to, *amount)
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

	bc, err := loadOrCreate(opts)
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

	bc, err := loadOrCreate(opts)
	if err != nil {
		return err
	}

	fmt.Println("Blockchain")
	fmt.Println("Difficulty:", bc.Difficulty)
	fmt.Println("Block size:", bc.BlockSize)
	fmt.Println("Blocks:", len(bc.Blocks))
	fmt.Println()

	for _, blk := range bc.Blocks {
		fmt.Println("----------------------------------------")
		fmt.Println("--------Toy Blockchain CLI Project------")
		fmt.Println("----------------------------------------")
		fmt.Println("Height:", blk.Height)
		fmt.Println("Timestamp:", blk.Timestamp)
		fmt.Println("Previous hash:", blk.PrevHash)
		fmt.Println("Nonce:", blk.Nonce)
		fmt.Println("Hash:", blk.Hash)
		fmt.Println("Transactions:")
		fmt.Println("----------------------------------------")

		if len(blk.Transactions) == 0 {
			fmt.Println("  none")
		}

		for _, tx := range blk.Transactions {
			fmt.Printf("  %s -> %s : %d\n", tx.From, tx.To, tx.Amount)
		}
	}

	fmt.Println("----------------------------------------")
	return nil
}

func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	opts := addCommonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}

	bc, err := loadOrCreate(opts)
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

	bc, err := loadOrCreate(opts)
	if err != nil {
		return err
	}

	balances := bc.Balances()
	if *includePending {
		balances = bc.BalancesIncludingPending()
	}

	if len(balances) == 0 {
		fmt.Println("No balances yet")
		return nil
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

	return nil
}

func runPending(args []string) error {
	fs := flag.NewFlagSet("pending", flag.ExitOnError)
	opts := addCommonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}

	bc, err := loadOrCreate(opts)
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

	bc, err := loadOrCreate(opts)
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

func printUsage() {
	fmt.Println("Toy Blockchain CLI")
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
	fmt.Println("  go run ./cmd/toychain mine")
	fmt.Println("  go run ./cmd/toychain add -from Alice -to Bob -amount 30")
	fmt.Println("  go run ./cmd/toychain mine")
	fmt.Println("  go run ./cmd/toychain balances")
	fmt.Println("  go run ./cmd/toychain validate")
}
