package common

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

var (
	pattern, _ = regexp.Compile("[0-9]+\\.{0,1}[0-9]*[eE][-+]{0,1}[0-9]+")
)

func Pow(base common.Dec, exp int) common.Dec {
	if exp < 0 {
		return Pow(common.NewDec(1).Quo(base), -exp)
	}
	result := common.NewDec(1)
	for {
		if exp%2 == 1 {
			result = result.Mul(base)
		}
		exp = exp >> 1
		if exp == 0 {
			break
		}
		base = base.Mul(base)
	}
	return result
}

func NewDecFromString(i string) (common.Dec, error) {
	if strings.HasPrefix(i, "-") {
		return common.ZeroDec(), errors.New(fmt.Sprintf("can not be negative: %s", i))
	}
	if pattern.FindString(i) != "" {
		if tokens := strings.Split(i, "e"); len(tokens) > 1 {
			a, _ := common.NewDecFromStr(tokens[0])
			b, _ := strconv.Atoi(tokens[1])
			return a.Mul(Pow(common.NewDec(10), b)), nil
		}
		tokens := strings.Split(i, "E")
		a, _ := common.NewDecFromStr(tokens[0])
		b, _ := strconv.Atoi(tokens[1])
		return a.Mul(Pow(common.NewDec(10), b)), nil
	} else {
		if strings.HasPrefix(i, ".") {
			i = "0" + i
		}
		return common.NewDecFromStr(i)
	}
}

// Assumes Hex string input
// Split into 2 64 bit integers to guarentee 128 bit precision
func NewDecFromHex(str string) common.Dec {
	str = strings.TrimPrefix(str, "0x")
	half := len(str) / 2
	right := str[half:]
	r, _ := big.NewInt(0).SetString(right, 16)
	if half == 0 {
		return common.NewDecFromBigInt(r)
	}
	left := str[:half]
	l, _ := big.NewInt(0).SetString(left, 16)
	return common.NewDecFromBigInt(l).Mul(
		Pow(common.NewDec(16), len(right)),
	).Add(common.NewDecFromBigInt(r))
}
