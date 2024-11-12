package util

import (
	"fmt"
	"math/big"
)

func HexToInt(hexStr string) int64 {
	n := new(big.Int)
	n.SetString(hexStr[2:], 16)
	return n.Int64()
}

func IntToHex(n int64) string {
	return fmt.Sprintf("0x%x", n)
}

func HexToBigInt(hexStr string) *big.Int {
	n := new(big.Int)
	n.SetString(hexStr[2:], 16)
	return n
}

func WeiToEth(wei *big.Int) *big.Float {
	// 1 ETH = 1,000,000,000,000,000,000 Wei
	eth := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
	return eth
}

func TrimQuotes(str string) string {
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		return str[1 : len(str)-1]
	}
	return str
}
