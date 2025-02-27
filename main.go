package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/kr/pretty"
	"github.com/santegoeds/nbx"
)

func main() {
	side := flag.String("side", "", "Order side: 'buy' or 'sell' (required)")
	market := flag.String("market", "", "Market symbol (e.g., BTC-NOK) (required)")
	fiatAmount := flag.Float64("fiatAmount", 0, "Fiat amount to spend (required for buy orders)")
	quantity := flag.Float64("quantity", 0, "Asset quantity to sell (required for sell orders)")
	showHelp := flag.Bool("h", false, "Show usage information")

	flag.Parse()

	if *showHelp {
		printUsageAndExit(nil)
	}

	if *side == "" || *market == "" {
		printUsageAndExit(errors.New("arguments 'side' and 'market' are required"))
	}

	if *side == "buy" && *fiatAmount <= 0 {
		printUsageAndExit(errors.New("argument 'fiatAmount' must be greater than zero for buy orders"))
	}

	if *side == "sell" && *quantity <= 0 {
		printUsageAndExit(errors.New("argument 'quantity' must be greater than zero for sell orders"))
	}

	ctx := context.Background()
	client := nbx.NewClient()

	accountID := os.Getenv("NBX_ACCOUNT_ID")
	keyID := os.Getenv("NBX_KEY")
	secret := os.Getenv("NBX_SECRET")
	passphrase := os.Getenv("NBX_PASSPHRASE")

	if err := client.Authenticate(ctx, accountID, keyID, secret, passphrase, nbx.Minute); err != nil {
		fmt.Printf("Failed to authenticate: %v\n", err)
		os.Exit(1)
	}

	switch *side {
	case "buy":
		orderID, err := client.MarketBuy(ctx, *market, 0, *fiatAmount) // Fix: Use fiatAmount for buy orders
		if err != nil {
			fmt.Printf("Failed to place buy order: %v\n", err)
			os.Exit(1)
		}
		printOrder(ctx, client, orderID)

	case "sell":
		orderID, err := client.MarketSell(ctx, *market, *quantity)
		if err != nil {
			fmt.Printf("Failed to place sell order: %v\n", err)
			os.Exit(1)
		}
		printOrder(ctx, client, orderID)

	default:
		printUsageAndExit(errors.New("invalid side, must be 'buy' or 'sell'"))
	}
}

func printUsageAndExit(err error) {
	programName := os.Args[0]
	if err != nil {
		fmt.Printf("Error: %v\n\n", err)
	}
	fmt.Println("Usage:")
	fmt.Printf("  %s --side=<buy|sell> --market=<market> --fiatAmount=<fiat> --quantity=<amount>\n", programName)
	fmt.Println("\nRequired Arguments:")
	fmt.Println("  --side         Order side ('buy' or 'sell')")
	fmt.Println("  --market       Market symbol (e.g., BTC-NOK)")
	fmt.Println("  --fiatAmount   Amount of fiat currency to spend (required for buy orders)")
	fmt.Println("  --quantity     Asset amount to sell (required for sell orders)")

	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  NBX_ACCOUNT_ID    Your NBX account ID")
	fmt.Println("  NBX_KEY           Your NBX API key")
	fmt.Println("  NBX_SECRET        Your NBX API secret")
	fmt.Println("  NBX_PASSPHRASE    Your NBX API passphrase")

	fmt.Println("\nExamples:")
	fmt.Printf("  %s --side=buy --market=BTC-NOK --fiatAmount=1000\n", programName)
	fmt.Printf("  %s --side=sell --market=BTC-NOK --quantity=0.1\n", programName)
	os.Exit(1)
}

func printOrder(ctx context.Context, client *nbx.Client, orderID string) {
	order, err := client.GetOrder(ctx, orderID)
	if err != nil {
		fmt.Printf("Failed to retrieve order: %v\n", err)
		os.Exit(1)
	}
	pretty.Println(order)
}
