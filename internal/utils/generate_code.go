package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateCode(length int) string {
	if length <= 0 {
		length = 6
	}

	// Calculate the max number (e.g. for 6 → 999999)
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		// fallback to a default code
		return "000000"
	}

	// Format with leading zeros, e.g. 000123
	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, n.Int64())
}
