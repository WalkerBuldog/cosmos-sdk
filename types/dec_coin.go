package types

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// ----------------------------------------------------------------------------
// Decimal Coin

// Coins which can have additional decimal points
type DecCoin struct {
	Denom  string `json:"denom"`
	Amount Dec    `json:"amount"`
}

// NewDecCoin returns a new coin with a denomination and amount.
// It will panic if the amount is negative.
func NewDecCoin(denom string, amount int64) DecCoin {
	c := DecCoin{Denom: denom, Amount: NewDec(amount)}
	if err := c.Validate(false); err != nil {
		panic(err)
	}
	return c
}

// NewDecCoin returns a new coin with a denomination and amount.
// It will panic if the amount is less than or equal to zero.
func NewPositiveDecCoin(denom string, amount int64) DecCoin {
	c := DecCoin{Denom: denom, Amount: NewDec(amount)}
	if err := c.Validate(true); err != nil {
		panic(err)
	}
	return c
}

// NewDecCoinFromDec returns a new coin with a denomination and
// decimal amount. It will panic if the amount is negative.
func NewDecCoinFromDec(denom string, amount Dec) DecCoin {
	c := DecCoin{Denom: denom, Amount: amount}
	if err := c.Validate(false); err != nil {
		panic(err)
	}
	return c
}

// NewDecCoinFromDec returns a new coin with a denomination and
// It will panic if the amount is less than or equal to zero.
func NewPositiveDecCoinFromDec(denom string, amount Dec) DecCoin {
	c := DecCoin{Denom: denom, Amount: amount}
	if err := c.Validate(true); err != nil {
		panic(err)
	}
	return c
}

// NewDecCoinFromCoin returns a new DecCoin from a Coin.
// It will panic if the amount is negative.
func NewDecCoinFromCoin(coin Coin) DecCoin {
	c := DecCoin{Denom: coin.Denom, Amount: NewDecFromInt(coin.Amount)}
	if err := c.Validate(false); err != nil {
		panic(err)
	}
	return c
}

// NewDecCoinFromCoin returns a new DecCoin from a Coin.
// It will panic if the amount is negative.
func NewPositiveDecCoinFromCoin(coin Coin) DecCoin {
	c := DecCoin{Denom: coin.Denom, Amount: NewDecFromInt(coin.Amount)}
	if err := c.Validate(true); err != nil {
		panic(err)
	}
	return c
}

// Validate validates coin's Amount and Denom. If strict is true,
// then it returns an error if Amount less than or equal to zero.
// If strict is false, then it returns an error if and only if
// Amount is less than zero.
func (coin DecCoin) Validate(strict bool) error {
	if err := validateDecCoinAmount(coin.Amount, strict); err != nil {
		return fmt.Errorf("%s: %s", err, coin.Amount)
	}

	if err := validateCoinDenomContainsSpace(coin.Denom); err != nil {
		return fmt.Errorf("%s: %s", err, coin.Denom)
	}

	if err := validateCoinDenomCase(coin.Denom); err != nil {
		return fmt.Errorf("%s: %s", err, coin.Denom)
	}

	return nil
}

// Adds amounts of two coins with same denom
func (coin DecCoin) Plus(coinB DecCoin) DecCoin {
	if coin.Denom != coinB.Denom {
		panic(fmt.Sprintf("coin denom different: %v %v\n", coin.Denom, coinB.Denom))
	}
	return DecCoin{coin.Denom, coin.Amount.Add(coinB.Amount)}
}

// Subtracts amounts of two coins with same denom
func (coin DecCoin) Minus(coinB DecCoin) DecCoin {
	if coin.Denom != coinB.Denom {
		panic(fmt.Sprintf("coin denom different: %v %v\n", coin.Denom, coinB.Denom))
	}
	return DecCoin{coin.Denom, coin.Amount.Sub(coinB.Amount)}
}

// return the decimal coins with trunctated decimals, and return the change
func (coin DecCoin) TruncateDecimal() (Coin, DecCoin) {
	truncated := coin.Amount.TruncateInt()
	change := coin.Amount.Sub(NewDecFromInt(truncated))
	return NewCoin(coin.Denom, truncated), NewDecCoinFromDec(coin.Denom, change)
}

// IsPositive returns true if coin amount is positive.
//
// TODO: Remove once unsigned integers are used.
func (coin DecCoin) IsPositive() bool {
	return coin.Amount.IsPositive()
}

// String implements the Stringer interface for DecCoin. It returns a
// human-readable representation of a decimal coin or an empty string
// if the amount is 0.
func (coin DecCoin) String() string {
	if !coin.IsPositive() {
		return ""
	}
	return fmt.Sprintf("%v%v", coin.Amount, coin.Denom)
}

// ----------------------------------------------------------------------------
// Decimal Coins

// coins with decimal
type DecCoins []DecCoin

func NewDecCoins(coins Coins) DecCoins {
	dcs := make(DecCoins, len(coins))
	for i, coin := range coins {
		dcs[i] = NewDecCoinFromCoin(coin)
	}
	return dcs
}

// String implements the Stringer interface for DecCoins. It returns a
// human-readable representation of decimal coins.
func (coins DecCoins) String() string {
	if len(coins) == 0 {
		return ""
	}

	out := []string{}
	for _, coin := range coins {
		if coin.IsPositive() {
			out = append(out, coin.String())
		}
	}

	return strings.Join(out, ",")
}

// return the coins with trunctated decimals, and return the change
func (coins DecCoins) TruncateDecimal() (Coins, DecCoins) {
	changeSum := DecCoins{}
	out := make(Coins, len(coins))
	for i, coin := range coins {
		truncated, change := coin.TruncateDecimal()
		out[i] = truncated
		changeSum = changeSum.Plus(DecCoins{change})
	}
	return out, changeSum
}

// Plus combines two sets of coins
// CONTRACT: Plus will never return Coins where one Coin has a 0 amount.
func (coins DecCoins) Plus(coinsB DecCoins) DecCoins {
	sum := ([]DecCoin)(nil)
	indexA, indexB := 0, 0
	lenA, lenB := len(coins), len(coinsB)
	for {
		if indexA == lenA {
			if indexB == lenB {
				return sum
			}
			return append(sum, coinsB[indexB:]...)
		} else if indexB == lenB {
			return append(sum, coins[indexA:]...)
		}
		coinA, coinB := coins[indexA], coinsB[indexB]
		switch strings.Compare(coinA.Denom, coinB.Denom) {
		case -1:
			sum = append(sum, coinA)
			indexA++
		case 0:
			if coinA.Amount.Add(coinB.Amount).IsZero() {
				// ignore 0 sum coin type
			} else {
				sum = append(sum, coinA.Plus(coinB))
			}
			indexA++
			indexB++
		case 1:
			sum = append(sum, coinB)
			indexB++
		}
	}
}

// Negative returns a set of coins with all amount negative
func (coins DecCoins) Negative() DecCoins {
	res := make([]DecCoin, 0, len(coins))
	for _, coin := range coins {
		res = append(res, DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.Neg(),
		})
	}
	return res
}

// Minus subtracts a set of coins from another (adds the inverse)
func (coins DecCoins) Minus(coinsB DecCoins) DecCoins {
	return coins.Plus(coinsB.Negative())
}

// multiply all the coins by a decimal
func (coins DecCoins) MulDec(d Dec) DecCoins {
	res := make([]DecCoin, len(coins))
	for i, coin := range coins {
		product := DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.Mul(d),
		}
		res[i] = product
	}
	return res
}

// multiply all the coins by a decimal, truncating
func (coins DecCoins) MulDecTruncate(d Dec) DecCoins {
	res := make([]DecCoin, len(coins))
	for i, coin := range coins {
		product := DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.MulTruncate(d),
		}
		res[i] = product
	}
	return res
}

// divide all the coins by a decimal
func (coins DecCoins) QuoDec(d Dec) DecCoins {
	res := make([]DecCoin, len(coins))
	for i, coin := range coins {
		quotient := DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.Quo(d),
		}
		res[i] = quotient
	}
	return res
}

// divide all the coins by a decimal, truncating
func (coins DecCoins) QuoDecTruncate(d Dec) DecCoins {
	res := make([]DecCoin, len(coins))
	for i, coin := range coins {
		quotient := DecCoin{
			Denom:  coin.Denom,
			Amount: coin.Amount.QuoTruncate(d),
		}
		res[i] = quotient
	}
	return res
}

// returns the amount of a denom from deccoins
func (coins DecCoins) AmountOf(denom string) Dec {
	switch len(coins) {
	case 0:
		return ZeroDec()
	case 1:
		coin := coins[0]
		if coin.Denom == denom {
			return coin.Amount
		}
		return ZeroDec()
	default:
		midIdx := len(coins) / 2 // binary search
		coin := coins[midIdx]
		if denom < coin.Denom {
			return coins[:midIdx].AmountOf(denom)
		} else if denom == coin.Denom {
			return coin.Amount
		} else {
			return coins[midIdx+1:].AmountOf(denom)
		}
	}
}

// has a negative DecCoin amount
func (coins DecCoins) HasNegative() bool {
	for _, coin := range coins {
		if coin.Amount.IsNegative() {
			return true
		}
	}
	return false
}

// return whether all coins are zero
func (coins DecCoins) IsZero() bool {
	for _, coin := range coins {
		if !coin.Amount.IsZero() {
			return false
		}
	}
	return true
}

// Validate asserts the DecCoins are sorted, have positive amount, and Denom
// does not contain upper case characters.
func (coins DecCoins) IsValid() bool {
	switch len(coins) {
	case 0:
		return true
	case 1:
		if err := coins[0].Validate(true); err != nil {
			return false
		}
		return true
	default:
		// check single coin case
		if !(DecCoins{coins[0]}).IsValid() {
			return false
		}

		lowDenom := coins[0].Denom
		for _, coin := range coins[1:] {
			if err := coin.Validate(true); err != nil {
				return false
			}
			if coin.Denom <= lowDenom {
				return false
			}

			// we compare each coin against the last denom
			lowDenom = coin.Denom
		}

		return true
	}
}

//-----------------------------------------------------------------------------
// Sorting

var _ sort.Interface = Coins{}

//nolint
func (coins DecCoins) Len() int           { return len(coins) }
func (coins DecCoins) Less(i, j int) bool { return coins[i].Denom < coins[j].Denom }
func (coins DecCoins) Swap(i, j int)      { coins[i], coins[j] = coins[j], coins[i] }

// Sort is a helper function to sort the set of decimal coins in-place.
func (coins DecCoins) Sort() DecCoins {
	sort.Sort(coins)
	return coins
}

// ----------------------------------------------------------------------------
// Parsing

// ParseDecCoin parses a decimal coin from a string, returning an error if
// invalid. An empty string is considered invalid.
func ParseDecCoin(coinStr string) (DecCoin, error) {
	coin, err := parseDecCoinString(coinStr)
	if err != nil {
		return DecCoin{}, fmt.Errorf("failed to parse decimal coin: %s", err)
	}

	if err := coin.Validate(false); err != nil {
		return DecCoin{}, fmt.Errorf("validation error: %s", err)
	}

	return coin, nil
}

// ParsePositiveDecCoin parses a decimal coin from a string, returning an error if
// invalid. An empty string is considered invalid.
func ParsePositiveDecCoin(coinStr string) (DecCoin, error) {
	coin, err := parseDecCoinString(coinStr)
	if err != nil {
		return DecCoin{}, fmt.Errorf("failed to parse decimal coin: %s", err)
	}

	if err := coin.Validate(true); err != nil {
		return DecCoin{}, fmt.Errorf("validation error: %s", err)
	}

	return coin, nil
}

// ParseDecCoins will parse out a list of decimal coins separated by commas.
// If nothing is provided, it returns nil DecCoins. Returned decimal coins are
// sorted.
func ParseDecCoins(coinsStr string) (coins DecCoins, err error) {
	coinsStr = strings.TrimSpace(coinsStr)
	if len(coinsStr) == 0 {
		return nil, nil
	}

	splitRe := regexp.MustCompile(",|;")
	coinStrs := splitRe.Split(coinsStr, -1)
	for _, coinStr := range coinStrs {
		coin, err := ParseDecCoin(coinStr)
		if err != nil {
			return nil, err
		}

		coins = append(coins, coin)
	}

	// sort coins for determinism
	coins.Sort()

	// validate coins before returning
	if !coins.IsValid() {
		return nil, fmt.Errorf("parsed decimal coins are invalid: %#v", coins)
	}

	return coins, nil
}

func validateDecCoinAmount(amount Dec, strict bool) error {
	if strict && amount.LTE(ZeroDec()) {
		return errors.New("non-positive coin amount")
	}
	if !strict && amount.LT(ZeroDec()) {
		return errors.New("negative coin amount")
	}
	return nil
}

func parseDecCoinString(coinStr string) (DecCoin, error) {
	coinStr = strings.TrimSpace(coinStr)
	matches := reDecCoin.FindStringSubmatch(coinStr)
	if matches == nil {
		return DecCoin{}, fmt.Errorf("invalid decimal coin expression %q", coinStr)
	}

	amountStr, denomStr := matches[1], matches[2]
	amount, err := NewDecFromStr(amountStr)
	if err != nil {
		return DecCoin{}, fmt.Errorf("failed to parse decimal coin amount: %s", amountStr)
	}

	return DecCoin{Denom: denomStr, Amount: amount}, nil
}
