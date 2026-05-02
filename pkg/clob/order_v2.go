package clob

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
)

func normalizeOrderVersion(version int) int {
	if version == 0 {
		return 2
	}
	return version
}

func orderVersion(order *clobtypes.Order, fallback int) int {
	if order != nil && order.Version != 0 {
		return normalizeOrderVersion(order.Version)
	}
	return normalizeOrderVersion(fallback)
}

func normalizeBytes32(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Bytes32Zero, nil
	}
	if len(value) >= 2 && value[:2] == "0X" {
		value = "0x" + value[2:]
	}
	if !strings.HasPrefix(value, "0x") {
		return "", fmt.Errorf("value must be 0x-prefixed")
	}
	raw := value[2:]
	if len(raw) != 64 {
		return "", fmt.Errorf("value must be 32 bytes")
	}
	if _, err := hex.DecodeString(raw); err != nil {
		return "", fmt.Errorf("value must be hex: %w", err)
	}
	return "0x" + strings.ToLower(raw), nil
}

func orderTimestamp(order *clobtypes.Order) *big.Int {
	if order == nil || order.Timestamp.Int == nil || order.Timestamp.Int.Sign() == 0 {
		return big.NewInt(0)
	}
	return order.Timestamp.Int
}

func orderNegRisk(order *clobtypes.Order) bool {
	return order != nil && order.NegRisk != nil && *order.NegRisk
}

func exchangeAddressForOrder(signer auth.Signer, order *clobtypes.Order, version int, defaultExchange, defaultNegRiskExchange string) string {
	if order != nil && strings.TrimSpace(order.VerifyingContract) != "" {
		return strings.TrimSpace(order.VerifyingContract)
	}

	negRisk := orderNegRisk(order)
	if negRisk && strings.TrimSpace(defaultNegRiskExchange) != "" {
		return strings.TrimSpace(defaultNegRiskExchange)
	}
	if !negRisk && strings.TrimSpace(defaultExchange) != "" {
		return strings.TrimSpace(defaultExchange)
	}

	chainID := int64(0)
	if signer != nil && signer.ChainID() != nil {
		chainID = signer.ChainID().Int64()
	}
	switch chainID {
	case auth.AmoyChainID:
		if version == 1 {
			if negRisk {
				return AmoyNegRiskExchangeV1
			}
			return AmoyExchangeV1
		}
		if negRisk {
			return AmoyNegRiskExchangeV2
		}
		return AmoyExchangeV2
	default:
		if version == 1 {
			if negRisk {
				return PolygonNegRiskExchangeV1
			}
			return PolygonExchangeV1
		}
		if negRisk {
			return PolygonNegRiskExchangeV2
		}
		return PolygonExchangeV2
	}
}
