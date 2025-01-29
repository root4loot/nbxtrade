package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kr/pretty"
	"github.com/santegoeds/nbx"
)

func main() {
	side := flag.String("side", "", "Order side: 'buy' or 'sell' (required)")
	market := flag.String("market", "", "Market symbol (e.g., BTC-NOK) (required)")
	fiatAmount := flag.Float64("fiatAmount", 0, "Fiat amount to spend for a buy order (required for 'buy')")
	showHelp := flag.Bool("h", false, "Show usage information")

	flag.Parse()

	if *showHelp {
		printUsageAndExit(nil)
	}

	if *side == "" || *market == "" {
		printUsageAndExit(errors.New("arguments 'side' and 'market' are required"))
	}

	if *side == "buy" && *fiatAmount <= 0 {
		printUsageAndExit(errors.New("argument 'fiatAmount' must be greater than zero for a buy order"))
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
		asset := strings.SplitN(*market, "-", 2)[0]
		var maxQuantity float64

		// Dirty hack to determine max quantity based on the market
		switch asset {
		case "BTC":
			maxQuantity = 1
		case "LTC":
			maxQuantity = 100
		case "ATOM":
			maxQuantity = 1000
		default:
			fmt.Printf("Unsupported market for buy: %s\n", *market)
			os.Exit(1)
		}

		orderID, err := client.MarketBuy(ctx, *market, maxQuantity, *fiatAmount)
		if err != nil {
			fmt.Printf("Failed to place buy order: %v\n", err)
			os.Exit(1)
		}
		printOrder(ctx, client, orderID)

	case "sell":
		orderID, err := client.MarketSell(ctx, *market, *flag.Float64("quantity", 0, "Exact amount to sell (required)"))
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
	fmt.Printf("  %s --side=<buy|sell> --market=<market> --fiatAmount=<fiat>\n", programName)
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  NBX_ACCOUNT_ID    Your NBX account ID")
	fmt.Println("  NBX_KEY           Your NBX API key")
	fmt.Println("  NBX_SECRET        Your NBX API secret")
	fmt.Println("  NBX_PASSPHRASE    Your NBX API passphrase")
	fmt.Println("\nExamples:")
	fmt.Printf("  %s --side=buy --market=BTC-NOK --fiatAmount=30000\n", programName)
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
