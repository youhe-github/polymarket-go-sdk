package clob

import (
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/shopspring/decimal"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

type recordingSigner struct {
	auth.Signer
	domain      *apitypes.TypedDataDomain
	typesDef    apitypes.Types
	message     apitypes.TypedDataMessage
	primaryType string
}

func (s *recordingSigner) SignTypedData(domain *apitypes.TypedDataDomain, typesDef apitypes.Types, message apitypes.TypedDataMessage, primaryType string) ([]byte, error) {
	s.domain = domain
	s.typesDef = typesDef
	s.message = message
	s.primaryType = primaryType
	return []byte{1, 2, 3}, nil
}

func TestOrderManagementMethods(t *testing.T) {
	signer, _ := auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	apiKey := &auth.APIKey{Key: "k1", Secret: "s1", Passphrase: "p1"}
	ctx := context.Background()

	t.Run("PostOrder", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/order": `{"orderID":"o1","status":"OK"}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
			signer:     signer,
			apiKey:     apiKey,
		}
		order := &clobtypes.SignedOrder{
			Order:     clobtypes.Order{Side: "BUY"},
			Signature: "0x123",
			Owner:     "0xabc",
		}
		resp, err := client.PostOrder(ctx, order)
		if err != nil || resp.ID != "o1" {
			t.Errorf("PostOrder failed: %v", err)
		}
	})

	t.Run("CancelAll", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/cancel-all": `{"canceled":["o1","o2"]}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.CancelAll(ctx)
		if err != nil {
			t.Errorf("CancelAll failed: %v", err)
		}
		if len(resp.Canceled) != 2 {
			t.Errorf("expected 2 canceled, got %d", len(resp.Canceled))
		}
	})

	t.Run("CancelOrder", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/order": `{"canceled":["o1"]}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.CancelOrder(ctx, &clobtypes.CancelOrderRequest{OrderID: "o1"})
		if err != nil {
			t.Errorf("CancelOrder failed: %v", err)
		}
		if len(resp.Canceled) != 1 || resp.Canceled[0] != "o1" {
			t.Errorf("expected canceled [o1], got %v", resp.Canceled)
		}
	})

	t.Run("CancelOrders", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/orders": `{"canceled":["o1"]}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.CancelOrders(ctx, &clobtypes.CancelOrdersRequest{OrderIDs: []string{"o1"}})
		if err != nil {
			t.Errorf("CancelOrders failed: %v", err)
		}
		if len(resp.Canceled) != 1 || resp.Canceled[0] != "o1" {
			t.Errorf("expected canceled [o1], got %v", resp.Canceled)
		}
	})

	t.Run("CancelMarketOrders", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/cancel-market-orders": `{"canceled":["o1"]}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.CancelMarketOrders(ctx, &clobtypes.CancelMarketOrdersRequest{Market: "m1"})
		if err != nil {
			t.Errorf("CancelMarketOrders failed: %v", err)
		}
		if len(resp.Canceled) != 1 || resp.Canceled[0] != "o1" {
			t.Errorf("expected canceled [o1], got %v", resp.Canceled)
		}
	})

	t.Run("BuilderTrades", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/builder/trades": `{"data":[]}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.BuilderTrades(ctx, nil)
		if err != nil {
			t.Errorf("BuilderTrades failed: %v", err)
		}
		if resp.Data == nil {
			t.Errorf("expected empty slice")
		}
	})

	t.Run("OrderLookup", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/data/order/o1": `{"orderID":"o1","status":"OK"}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.Order(ctx, "o1")
		if err != nil || resp.ID != "o1" {
			t.Errorf("Order lookup failed: %v", err)
		}
	})

	t.Run("OrdersList", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/data/orders": `{"data":[{"id":"o1"}],"next_cursor":"LTE="}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.Orders(ctx, nil)
		if err != nil {
			t.Fatalf("Orders list failed: %v", err)
		}
		if len(resp.Data) == 0 {
			t.Fatal("Orders list returned no data")
		}
		if resp.Data[0].ID != "o1" {
			t.Errorf("Orders list ID = %s, want o1", resp.Data[0].ID)
		}
	})

	t.Run("OrdersListNumericCreatedAt", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/data/orders": `{"data":[{"orderID":"o1","created_at":1700000000,"timestamp":1700000001}],"next_cursor":"LTE="}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.Orders(ctx, nil)
		if err != nil {
			t.Fatalf("Orders list failed: %v", err)
		}
		if len(resp.Data) != 1 {
			t.Fatalf("len(resp.Data) = %d, want 1", len(resp.Data))
		}
		if resp.Data[0].CreatedAt != "1700000000" {
			t.Errorf("CreatedAt = %s, want 1700000000", resp.Data[0].CreatedAt)
		}
		if resp.Data[0].Timestamp != "1700000001" {
			t.Errorf("Timestamp = %s, want 1700000001", resp.Data[0].Timestamp)
		}
	})

	t.Run("OrderScoring", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/order-scoring?order_id=o1": `{"scoring":true}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.OrderScoring(ctx, &clobtypes.OrderScoringRequest{ID: "o1"})
		if err != nil || !resp.Scoring {
			t.Errorf("OrderScoring failed: %v", err)
		}
	})

	t.Run("OrdersScoring", func(t *testing.T) {
		doer := &staticDoer{
			responses: map[string]string{"/orders-scoring": `{"o1":true,"o2":false}`},
		}
		client := &clientImpl{
			httpClient: transport.NewClient(doer, "http://example"),
		}
		resp, err := client.OrdersScoring(ctx, &clobtypes.OrdersScoringRequest{IDs: []string{"o1", "o2"}})
		if err != nil || resp["o1"] != true || resp["o2"] != false {
			t.Errorf("OrdersScoring failed: %v", err)
		}
	})
}

func TestSignOrderDefaults(t *testing.T) {
	signer, _ := auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	apiKey := &auth.APIKey{Key: "k1", Secret: "s1", Passphrase: "p1"}

	funder := common.HexToAddress("0x3333333333333333333333333333333333333333")
	client := &clientImpl{
		signer:        signer,
		apiKey:        apiKey,
		signatureType: auth.SignatureProxy,
		funder:        &funder,
		saltGenerator: func() (*big.Int, error) { return big.NewInt(7), nil },
	}

	order := &clobtypes.Order{
		Side:        "BUY",
		TokenID:     types.U256{Int: big.NewInt(1)},
		MakerAmount: decimal.NewFromInt(10),
		TakerAmount: decimal.NewFromInt(5),
		FeeRateBps:  decimal.NewFromInt(0),
		Nonce:       types.U256{Int: big.NewInt(1)},
		Expiration:  types.U256{Int: big.NewInt(0)},
		Taker:       common.Address{},
		Signer:      signer.Address(),
	}

	signed, err := client.signOrder(order)
	if err != nil {
		t.Fatalf("signOrder failed: %v", err)
	}
	if signed.Order.SignatureType == nil || *signed.Order.SignatureType != 1 {
		t.Fatalf("signature type mismatch: %+v", signed.Order.SignatureType)
	}
	if signed.Order.Maker != funder {
		t.Fatalf("maker mismatch: got %s want %s", signed.Order.Maker.Hex(), funder.Hex())
	}
	if signed.Order.Salt.Int == nil || signed.Order.Salt.Int.Int64() != 7 {
		t.Fatalf("salt mismatch: got %v", signed.Order.Salt.Int)
	}
}

func TestSignOrderV2TypedData(t *testing.T) {
	baseSigner, _ := auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	signer := &recordingSigner{Signer: baseSigner}
	apiKey := &auth.APIKey{Key: "k1", Secret: "s1", Passphrase: "p1"}
	sigType := int(auth.SignatureEOA)
	negRisk := true

	order := &clobtypes.Order{
		Salt:          types.U256{Int: big.NewInt(7)},
		Maker:         signer.Address(),
		Signer:        signer.Address(),
		Taker:         common.Address{},
		TokenID:       types.U256{Int: big.NewInt(1)},
		MakerAmount:   decimal.NewFromInt(10),
		TakerAmount:   decimal.NewFromInt(5),
		Side:          "BUY",
		Expiration:    types.U256{Int: big.NewInt(123)},
		FeeRateBps:    decimal.NewFromInt(99),
		Nonce:         types.U256{Int: big.NewInt(11)},
		SignatureType: &sigType,
		Timestamp:     types.U256{Int: big.NewInt(1710000000123)},
		Metadata:      Bytes32Zero,
		Builder:       "0x1111111111111111111111111111111111111111111111111111111111111111",
		NegRisk:       &negRisk,
	}

	signed, err := signOrderWithCreds(signer, apiKey, order, signOptions{})
	if err != nil {
		t.Fatalf("signOrderWithCreds failed: %v", err)
	}
	if signed.Order.Version != 2 {
		t.Fatalf("version mismatch: got %d", signed.Order.Version)
	}
	if signer.domain == nil || signer.domain.Version != "2" {
		t.Fatalf("expected v2 domain, got %+v", signer.domain)
	}
	if signer.domain.VerifyingContract != PolygonNegRiskExchangeV2 {
		t.Fatalf("verifying contract mismatch: got %s", signer.domain.VerifyingContract)
	}
	if _, ok := signer.message["taker"]; ok {
		t.Fatal("v2 typed data should not include taker")
	}
	if _, ok := signer.message["nonce"]; ok {
		t.Fatal("v2 typed data should not include nonce")
	}
	if _, ok := signer.message["feeRateBps"]; ok {
		t.Fatal("v2 typed data should not include feeRateBps")
	}
	if _, ok := signer.message["expiration"]; ok {
		t.Fatal("v2 typed data should not include expiration")
	}
	if signer.message["metadata"] != Bytes32Zero {
		t.Fatalf("metadata mismatch: got %v", signer.message["metadata"])
	}
	if signer.message["builder"] != strings.ToLower(order.Builder) {
		t.Fatalf("builder mismatch: got %v", signer.message["builder"])
	}
}

func TestSignOrderV1StillAvailable(t *testing.T) {
	baseSigner, _ := auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	signer := &recordingSigner{Signer: baseSigner}
	apiKey := &auth.APIKey{Key: "k1", Secret: "s1", Passphrase: "p1"}

	order := &clobtypes.Order{
		Version:     1,
		Salt:        types.U256{Int: big.NewInt(7)},
		Maker:       signer.Address(),
		Signer:      signer.Address(),
		TokenID:     types.U256{Int: big.NewInt(1)},
		MakerAmount: decimal.NewFromInt(10),
		TakerAmount: decimal.NewFromInt(5),
		Side:        "SELL",
		Expiration:  types.U256{Int: big.NewInt(123)},
		Nonce:       types.U256{Int: big.NewInt(11)},
	}

	if _, err := signOrderWithCreds(signer, apiKey, order, signOptions{}); err != nil {
		t.Fatalf("signOrderWithCreds failed: %v", err)
	}
	if signer.domain == nil || signer.domain.Version != "1" {
		t.Fatalf("expected v1 domain, got %+v", signer.domain)
	}
	for _, key := range []string{"taker", "nonce", "feeRateBps", "expiration"} {
		if _, ok := signer.message[key]; !ok {
			t.Fatalf("v1 typed data missing %s", key)
		}
	}
}

func TestPostOrders_BatchSizeValidation(t *testing.T) {
	ctx := context.Background()
	client := &clientImpl{
		httpClient: transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"),
	}

	// Exactly at the limit should not error (would error from server, but not from validation)
	atLimit := &clobtypes.SignedOrders{
		Orders: make([]clobtypes.SignedOrder, clobtypes.MaxPostOrdersBatchSize),
	}
	// This will fail at the HTTP level but should NOT fail at validation
	_, err := client.PostOrders(ctx, atLimit)
	if err != nil && strings.Contains(err.Error(), "batch size") {
		t.Errorf("expected no batch size error at limit, got: %v", err)
	}

	// Over the limit should error immediately
	overLimit := &clobtypes.SignedOrders{
		Orders: make([]clobtypes.SignedOrder, clobtypes.MaxPostOrdersBatchSize+1),
	}
	_, err = client.PostOrders(ctx, overLimit)
	if err == nil {
		t.Fatal("expected error for exceeding batch size")
	}
	if !strings.Contains(err.Error(), "batch size") {
		t.Errorf("expected batch size error, got: %v", err)
	}
}

func TestCancelOrders_BatchSizeValidation(t *testing.T) {
	ctx := context.Background()
	client := &clientImpl{
		httpClient: transport.NewClient(&staticDoer{responses: map[string]string{}}, "http://example"),
	}

	// Over the limit should error immediately
	ids := make([]string, clobtypes.MaxCancelOrdersBatchSize+1)
	for i := range ids {
		ids[i] = "order-" + strings.Repeat("x", 5)
	}
	_, err := client.CancelOrders(ctx, &clobtypes.CancelOrdersRequest{OrderIDs: ids})
	if err == nil {
		t.Fatal("expected error for exceeding cancel batch size")
	}
	if !strings.Contains(err.Error(), "batch size") {
		t.Errorf("expected batch size error, got: %v", err)
	}
}
