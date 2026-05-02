package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	polymarket "github.com/GoPolymarket/polymarket-go-sdk"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	ctx := context.Background()

	signer, apiKey := loadAuth()
	client := polymarket.NewClient(
		polymarket.WithUseServerTime(true),
		polymarket.WithCLOBBuilderCode(strings.TrimSpace(os.Getenv("POLYMARKET_BUILDER_CODE"))),
	)

	clobClient := client.CLOB.WithAuth(signer, apiKey)
	clobClient = applyWalletOptions(clobClient)

	mode := strings.ToLower(strings.TrimSpace(getenv("POLYMARKET_DEMO_MODE", "limit")))
	switch mode {
	case "limit":
		resp := placeLimitOrder(ctx, clobClient, signer)
		fmt.Printf("limit order: id=%s status=%s\n", resp.ID, resp.Status)
	case "market":
		resp := placeMarketOrder(ctx, clobClient, signer)
		fmt.Printf("market order: id=%s status=%s\n", resp.ID, resp.Status)
	case "cancel":
		cancelOrder(ctx, clobClient, requiredEnv("POLYMARKET_CANCEL_ORDER_ID"))
	case "all":
		limitResp := placeLimitOrder(ctx, clobClient, signer)
		fmt.Printf("limit order: id=%s status=%s\n", limitResp.ID, limitResp.Status)

		marketResp := placeMarketOrder(ctx, clobClient, signer)
		fmt.Printf("market order: id=%s status=%s\n", marketResp.ID, marketResp.Status)

		cancelID := strings.TrimSpace(os.Getenv("POLYMARKET_CANCEL_ORDER_ID"))
		if cancelID == "" {
			cancelID = limitResp.ID
		}
		if cancelID == "" {
			log.Fatal("no order id available to cancel; set POLYMARKET_CANCEL_ORDER_ID")
		}
		cancelOrder(ctx, clobClient, cancelID)
	default:
		log.Fatalf("unsupported POLYMARKET_DEMO_MODE %q (use limit, market, cancel, or all)", mode)
	}
}

func loadAuth() (auth.Signer, *auth.APIKey) {
	chainID := int64Env("POLYMARKET_CHAIN_ID", auth.PolygonChainID)
	signer, err := auth.NewPrivateKeySigner(requiredEnv("POLYMARKET_PK"), chainID)
	if err != nil {
		log.Fatalf("create signer: %v", err)
	}
	apiKey := &auth.APIKey{
		Key:        requiredEnv("POLYMARKET_API_KEY"),
		Secret:     requiredEnv("POLYMARKET_API_SECRET"),
		Passphrase: requiredEnv("POLYMARKET_API_PASSPHRASE"),
	}
	return signer, apiKey
}

func applyWalletOptions(client clob.Client) clob.Client {
	sigType := auth.SignatureType(int64Env("POLYMARKET_SIGNATURE_TYPE", int64(auth.SignatureEOA)))
	client = client.WithSignatureType(sigType)

	if funder := strings.TrimSpace(os.Getenv("POLYMARKET_FUNDER")); funder != "" {
		client = client.WithFunder(types.Address(common.HexToAddress(funder)))
	}
	return client
}

func placeLimitOrder(ctx context.Context, client clob.Client, signer auth.Signer) clobtypes.OrderResponse {
	tokenID := requiredEnv("POLYMARKET_TOKEN_ID")
	side := strings.ToUpper(getenv("POLYMARKET_LIMIT_SIDE", "BUY"))
	price := floatEnv("POLYMARKET_LIMIT_PRICE", 0)
	size := floatEnv("POLYMARKET_LIMIT_SIZE", 0)
	if price <= 0 || size <= 0 {
		log.Fatal("POLYMARKET_LIMIT_PRICE and POLYMARKET_LIMIT_SIZE must be positive")
	}

	signable, err := clob.NewOrderBuilder(client, signer).
		TokenID(tokenID).
		Side(side).
		Price(price).
		Size(size).
		OrderType(clobtypes.OrderTypeGTC).
		PostOnly(boolEnv("POLYMARKET_LIMIT_POST_ONLY", false)).
		BuildSignableWithContext(ctx)
	if err != nil {
		log.Fatalf("build limit order: %v", err)
	}

	resp, err := client.CreateOrderFromSignable(ctx, signable)
	if err != nil {
		log.Fatalf("post limit order: %v", err)
	}
	return resp
}

func placeMarketOrder(ctx context.Context, client clob.Client, signer auth.Signer) clobtypes.OrderResponse {
	tokenID := requiredEnv("POLYMARKET_TOKEN_ID")
	side := strings.ToUpper(getenv("POLYMARKET_MARKET_SIDE", "BUY"))

	builder := clob.NewOrderBuilder(client, signer).
		TokenID(tokenID).
		Side(side).
		OrderType(clobtypes.OrderTypeFAK)

	if side == "SELL" {
		amountShares := floatEnv("POLYMARKET_MARKET_AMOUNT_SHARES", 0)
		if amountShares <= 0 {
			log.Fatal("SELL market orders require POLYMARKET_MARKET_AMOUNT_SHARES")
		}
		builder = builder.AmountShares(amountShares)
	} else {
		amountUSDC := floatEnv("POLYMARKET_MARKET_AMOUNT_USDC", 0)
		if amountUSDC <= 0 {
			log.Fatal("BUY market orders require POLYMARKET_MARKET_AMOUNT_USDC")
		}
		builder = builder.AmountUSDC(amountUSDC)
	}

	signable, err := builder.BuildMarketWithContext(ctx)
	if err != nil {
		log.Fatalf("build market order: %v", err)
	}

	resp, err := client.CreateOrderFromSignable(ctx, signable)
	if err != nil {
		log.Fatalf("post market order: %v", err)
	}
	return resp
}

func cancelOrder(ctx context.Context, client clob.Client, orderID string) {
	resp, err := client.CancelOrder(ctx, &clobtypes.CancelOrderRequest{OrderID: orderID})
	if err != nil {
		log.Fatalf("cancel order %s: %v", orderID, err)
	}
	fmt.Printf("cancel order: id=%s canceled=%v not_canceled=%v\n", orderID, resp.Canceled, resp.NotCanceled)
}

func requiredEnv(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		log.Fatalf("%s is required", key)
	}
	return value
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func int64Env(key string, fallback int64) int64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		log.Fatalf("%s must be an integer: %v", key, err)
	}
	return value
}

func floatEnv(key string, fallback float64) float64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		log.Fatalf("%s must be a number: %v", key, err)
	}
	return value
}

func boolEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		log.Fatalf("%s must be a bool: %v", key, err)
	}
	return value
}
