package clob

import "github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"

import (
	"fmt"
	"strings"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"
)

func buildOrderPayload(order *clobtypes.SignedOrder) (map[string]interface{}, error) {
	if order == nil {
		return nil, fmt.Errorf("order is required")
	}
	orderType := normalizeOrderType(order.OrderType, clobtypes.OrderTypeGTC)
	if order.PostOnly != nil && *order.PostOnly && orderType != clobtypes.OrderTypeGTC && orderType != clobtypes.OrderTypeGTD {
		return nil, fmt.Errorf("postOnly is only supported for GTC and GTD orders")
	}
	orderMap, err := orderWithSignature(order)
	if err != nil {
		return nil, err
	}

	postOnly := false
	if order.PostOnly != nil {
		postOnly = *order.PostOnly
	}
	deferExec := false
	if order.DeferExec != nil {
		deferExec = *order.DeferExec
	}

	payload := map[string]interface{}{
		"order":     orderMap,
		"owner":     order.Owner,
		"orderType": orderType,
		"postOnly":  postOnly,
		"deferExec": deferExec,
	}
	return payload, nil
}

func buildOrdersPayload(orders *clobtypes.SignedOrders) ([]map[string]interface{}, error) {
	if orders == nil {
		return nil, fmt.Errorf("orders are required")
	}
	payloads := make([]map[string]interface{}, 0, len(orders.Orders))
	for idx := range orders.Orders {
		order := orders.Orders[idx]
		payload, err := buildOrderPayload(&order)
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, payload)
	}
	return payloads, nil
}

func orderWithSignature(order *clobtypes.SignedOrder) (map[string]interface{}, error) {
	if order == nil {
		return nil, fmt.Errorf("order is required")
	}
	if order.Signature == "" {
		return nil, fmt.Errorf("signature is required")
	}
	if order.Owner == "" {
		return nil, fmt.Errorf("owner is required")
	}

	sigType := 0
	if order.Order.SignatureType != nil {
		sigType = *order.Order.SignatureType
	}

	side := strings.ToUpper(order.Order.Side)
	if side != "BUY" && side != "SELL" {
		return nil, fmt.Errorf("invalid order side %q", order.Order.Side)
	}

	salt, err := saltToJSON(order.Order.Salt)
	if err != nil {
		return nil, err
	}

	if orderVersion(&order.Order, 2) == 2 {
		metadata, err := normalizeBytes32(order.Order.Metadata)
		if err != nil {
			return nil, fmt.Errorf("metadata: %w", err)
		}
		builderCode, err := normalizeBytes32(order.Order.Builder)
		if err != nil {
			return nil, fmt.Errorf("builder: %w", err)
		}
		return map[string]interface{}{
			"salt":          salt,
			"maker":         order.Order.Maker.Hex(),
			"signer":        order.Order.Signer.Hex(),
			"taker":         order.Order.Taker.Hex(),
			"tokenId":       u256String(order.Order.TokenID),
			"makerAmount":   decimalString(order.Order.MakerAmount),
			"takerAmount":   decimalString(order.Order.TakerAmount),
			"side":          side,
			"expiration":    u256String(order.Order.Expiration),
			"signatureType": sigType,
			"timestamp":     u256String(order.Order.Timestamp),
			"metadata":      metadata,
			"builder":       builderCode,
			"signature":     order.Signature,
		}, nil
	}

	return map[string]interface{}{
		"salt":          salt,
		"maker":         order.Order.Maker.Hex(),
		"signer":        order.Order.Signer.Hex(),
		"taker":         order.Order.Taker.Hex(),
		"tokenId":       u256String(order.Order.TokenID),
		"makerAmount":   decimalString(order.Order.MakerAmount),
		"takerAmount":   decimalString(order.Order.TakerAmount),
		"side":          side,
		"expiration":    u256String(order.Order.Expiration),
		"nonce":         u256String(order.Order.Nonce),
		"feeRateBps":    decimalString(order.Order.FeeRateBps),
		"signatureType": sigType,
		"signature":     order.Signature,
	}, nil
}

func u256String(value types.U256) string {
	if value.Int == nil {
		return "0"
	}
	return value.Int.String()
}

func decimalString(value types.Decimal) string {
	return value.String()
}

func saltToJSON(value types.U256) (interface{}, error) {
	if value.Int == nil {
		return uint64(0), nil
	}
	if value.Int.Sign() < 0 {
		return nil, fmt.Errorf("salt must be non-negative")
	}
	if value.Int.BitLen() > 53 {
		return nil, fmt.Errorf("salt is too large (max 53 bits)")
	}
	return value.Int.Uint64(), nil
}

func normalizeOrderType(orderType clobtypes.OrderType, fallback clobtypes.OrderType) clobtypes.OrderType {
	trimmed := strings.TrimSpace(string(orderType))
	if trimmed == "" {
		return fallback
	}
	upper := strings.ToUpper(trimmed)
	return clobtypes.OrderType(upper)
}
