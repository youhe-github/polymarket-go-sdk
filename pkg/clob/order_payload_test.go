package clob

import (
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

func TestBuildOrderPayloadCasingAndOptions(t *testing.T) {
	sigType := 0
	order := clobtypes.SignedOrder{
		Order: clobtypes.Order{
			Salt:          types.U256{Int: big.NewInt(1)},
			Maker:         common.HexToAddress("0x0000000000000000000000000000000000000001"),
			Signer:        common.HexToAddress("0x0000000000000000000000000000000000000002"),
			Taker:         common.HexToAddress("0x0000000000000000000000000000000000000000"),
			TokenID:       types.U256{Int: big.NewInt(123)},
			MakerAmount:   decimal.NewFromInt(100),
			TakerAmount:   decimal.NewFromInt(50),
			Side:          "BUY",
			Expiration:    types.U256{Int: big.NewInt(0)},
			FeeRateBps:    decimal.NewFromInt(0),
			Nonce:         types.U256{Int: big.NewInt(0)},
			SignatureType: &sigType,
		},
		Signature: "0xsig",
		Owner:     "builder-owner",
		OrderType: clobtypes.OrderTypeGTC,
		PostOnly:  boolPtr(true),
	}

	payload, err := buildOrderPayload(&order)
	if err != nil {
		t.Fatalf("buildOrderPayload failed: %v", err)
	}

	if payload["owner"] != "builder-owner" {
		t.Fatalf("owner mismatch: got %v", payload["owner"])
	}
	if got := payload["orderType"]; got != clobtypes.OrderTypeGTC {
		t.Fatalf("orderType mismatch: got %v", got)
	}

	orderMap, ok := payload["order"].(map[string]interface{})
	if !ok {
		t.Fatalf("order payload missing order map")
	}
	if orderMap["tokenId"] != "123" {
		t.Fatalf("tokenId mismatch: got %v", orderMap["tokenId"])
	}
	if orderMap["makerAmount"] == nil || orderMap["takerAmount"] == nil {
		t.Fatalf("maker/taker amounts missing in order payload")
	}
	if orderMap["signature"] != "0xsig" {
		t.Fatalf("signature mismatch: got %v", orderMap["signature"])
	}
}

func TestBuildOrderPayloadV2Fields(t *testing.T) {
	sigType := 0
	order := clobtypes.SignedOrder{
		Order: clobtypes.Order{
			Version:       2,
			Salt:          types.U256{Int: big.NewInt(1)},
			Maker:         common.HexToAddress("0x0000000000000000000000000000000000000001"),
			Signer:        common.HexToAddress("0x0000000000000000000000000000000000000002"),
			Taker:         common.HexToAddress("0x0000000000000000000000000000000000000000"),
			TokenID:       types.U256{Int: big.NewInt(123)},
			MakerAmount:   decimal.NewFromInt(100),
			TakerAmount:   decimal.NewFromInt(50),
			Side:          "BUY",
			Expiration:    types.U256{Int: big.NewInt(99)},
			Timestamp:     types.U256{Int: big.NewInt(1710000000123)},
			Metadata:      Bytes32Zero,
			Builder:       "0x1111111111111111111111111111111111111111111111111111111111111111",
			SignatureType: &sigType,
		},
		Signature: "0xsig",
		Owner:     "owner",
		OrderType: clobtypes.OrderTypeGTC,
	}

	payload, err := buildOrderPayload(&order)
	if err != nil {
		t.Fatalf("buildOrderPayload failed: %v", err)
	}
	orderMap := payload["order"].(map[string]interface{})
	for _, key := range []string{"timestamp", "metadata", "builder", "expiration", "signature"} {
		if _, ok := orderMap[key]; !ok {
			t.Fatalf("v2 payload missing %s", key)
		}
	}
	for _, key := range []string{"nonce", "feeRateBps"} {
		if _, ok := orderMap[key]; ok {
			t.Fatalf("v2 payload should not include %s", key)
		}
	}
	if orderMap["timestamp"] != "1710000000123" {
		t.Fatalf("timestamp mismatch: got %v", orderMap["timestamp"])
	}
	if orderMap["metadata"] != Bytes32Zero {
		t.Fatalf("metadata mismatch: got %v", orderMap["metadata"])
	}
	if orderMap["builder"] != strings.ToLower(order.Order.Builder) {
		t.Fatalf("builder mismatch: got %v", orderMap["builder"])
	}
}

func TestBuildOrderPayloadPostOnlyValidation(t *testing.T) {
	sigType := 0
	order := clobtypes.SignedOrder{
		Order: clobtypes.Order{
			Salt:          types.U256{Int: big.NewInt(1)},
			Maker:         common.HexToAddress("0x0000000000000000000000000000000000000001"),
			Signer:        common.HexToAddress("0x0000000000000000000000000000000000000002"),
			Taker:         common.HexToAddress("0x0000000000000000000000000000000000000000"),
			TokenID:       types.U256{Int: big.NewInt(123)},
			MakerAmount:   decimal.NewFromInt(100),
			TakerAmount:   decimal.NewFromInt(50),
			Side:          "BUY",
			Expiration:    types.U256{Int: big.NewInt(0)},
			FeeRateBps:    decimal.NewFromInt(0),
			Nonce:         types.U256{Int: big.NewInt(0)},
			SignatureType: &sigType,
		},
		Signature: "0xsig",
		Owner:     "builder-owner",
		OrderType: clobtypes.OrderTypeFAK,
		PostOnly:  boolPtr(true),
	}

	_, err := buildOrderPayload(&order)
	if err == nil || !strings.Contains(err.Error(), "postOnly") {
		t.Fatalf("expected postOnly validation error, got %v", err)
	}
}

func TestBuildOrderPayloadRequiresSignatureAndOwner(t *testing.T) {
	order := clobtypes.SignedOrder{
		Order: clobtypes.Order{
			Salt:        types.U256{Int: big.NewInt(1)},
			Maker:       common.HexToAddress("0x0000000000000000000000000000000000000001"),
			Signer:      common.HexToAddress("0x0000000000000000000000000000000000000002"),
			Taker:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
			TokenID:     types.U256{Int: big.NewInt(123)},
			MakerAmount: decimal.NewFromInt(100),
			TakerAmount: decimal.NewFromInt(50),
			Side:        "BUY",
			Expiration:  types.U256{Int: big.NewInt(0)},
			FeeRateBps:  decimal.NewFromInt(0),
			Nonce:       types.U256{Int: big.NewInt(0)},
		},
	}

	_, err := buildOrderPayload(&order)
	if err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("expected signature validation error, got %v", err)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
