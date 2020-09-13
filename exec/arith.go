package exec

import (
	"math"
	"math/big"

	"github.com/reg0007/Zn/error"
)

// Arith - arithmetic calculation (including + - * /) instance
type Arith struct {
	precision int
}

// NewArith -
func NewArith(precision int) *Arith {
	return &Arith{precision}
}

// Add - A + B + C + D + ... = ?
func (ai *Arith) Add(decimal1 *ZnDecimal, others ...*ZnDecimal) *ZnDecimal {
	var result = copyZnDecimal(decimal1)
	if len(others) == 0 {
		return result
	}

	for _, item := range others {
		r1, r2 := rescalePair(result, item)
		result.co.Add(r1.co, r2.co)
		result.exp = r1.exp
	}
	return result
}

// Sub - A - B - C - D - ... = ?
func (ai *Arith) Sub(decimal1 *ZnDecimal, others ...*ZnDecimal) *ZnDecimal {
	var result = copyZnDecimal(decimal1)
	if len(others) == 0 {
		return result
	}

	for _, item := range others {
		r1, r2 := rescalePair(result, item)
		result.co.Sub(r1.co, r2.co)
		result.exp = r1.exp
	}
	return result
}

// Mul - A * B * C * D * ... = ?, ZnDeicmal value will be copied
func (ai *Arith) Mul(decimal1 *ZnDecimal, others ...*ZnDecimal) *ZnDecimal {
	// init result from decimal1
	var result = copyZnDecimal(decimal1)
	if len(others) == 0 {
		return result
	}

	for _, item := range others {
		result.co.Mul(result.co, item.co)
		result.exp = result.exp + item.exp
	}

	return result
}

// Div - A / B / C / D / ... = ?, ZnDecimal value will be copied
// notice , when one of the dividends are 0, an ArithDivZeroError will be yield
func (ai *Arith) Div(decimal1 *ZnDecimal, others ...*ZnDecimal) (*ZnDecimal, *error.Error) {
	var result = copyZnDecimal(decimal1)
	var num10 = big.NewInt(10)
	var num0 = big.NewInt(0)
	if len(others) == 0 {
		return result, nil
	}

	// if divisor is zero, return 0 directly
	if decimal1.co.Cmp(num0) == 0 {
		return result, nil
	}
	for _, item := range others {
		// check if divident is zero
		if item.co.Cmp(num0) == 0 {
			return nil, error.ArithDivZeroError()
		}
		adjust := 0
		// adjust bits
		// C1 < C2
		if result.co.Cmp(item.co) < 0 {
			var c2_10x = new(big.Int).Mul(item.co, num10)
			for {
				if result.co.Cmp(item.co) >= 0 && result.co.Cmp(c2_10x) < 0 {
					break
				}
				// else, C1 = C1 * 10
				result.co.Mul(result.co, num10)
				adjust = adjust + 1
			}
		} else {
			var c1_10x = new(big.Int).Mul(result.co, num10)
			for {
				if item.co.Cmp(result.co) >= 0 && item.co.Cmp(c1_10x) < 0 {
					break
				}

				item.co.Mul(item.co, num10)
				adjust = adjust - 1
			}
		}

		// exp10x = 10^(precision)
		var precFactor = ai.precision - 1
		if adjust < 0 {
			precFactor = ai.precision
		}
		var exp10x *big.Int
		if ai.precision >= 18 {
			exp10x = new(big.Int).Exp(num10, num10, big.NewInt(int64(precFactor-1)))
		} else {
			exp10x = big.NewInt(int64(math.Pow10(precFactor)))
		}

		// do div
		var mul10x = exp10x.Mul(result.co, exp10x)
		var xr = new(big.Int)
		var xq, _ = result.co.DivMod(mul10x, item.co, xr) // don't use QuoRem here!

		// rounding
		if xr.Mul(xr, big.NewInt(2)).Cmp(item.co) > 0 {
			xq = xq.Add(xq, big.NewInt(1))
		}

		// get final result
		result.co = xq
		result.exp = result.exp - item.exp - adjust - precFactor
	}
	return result, nil
}

//// arith helper

// rescalePair - make exps to be same
func rescalePair(d1 *ZnDecimal, d2 *ZnDecimal) (*ZnDecimal, *ZnDecimal) {
	intTen := big.NewInt(10)

	if d1.exp == d2.exp {
		return d1, d2
	}
	if d1.exp > d2.exp {
		// return new d1
		diff := d1.exp - d2.exp

		expVal := new(big.Int).Exp(intTen, big.NewInt(int64(diff)), nil)
		nD1 := &ZnDecimal{
			co:  new(big.Int).Mul(d1.co, expVal),
			exp: d2.exp,
		}
		return nD1, d2
	}
	// d1.exp < d2.exp
	// return new d2
	diff := d2.exp - d1.exp

	expVal := new(big.Int).Exp(intTen, big.NewInt(int64(diff)), nil)
	nD2 := &ZnDecimal{
		co:  new(big.Int).Mul(d2.co, expVal),
		exp: d1.exp,
	}
	return d1, nD2
}

// copyDecimal - duplicate deicmal value to a new variable
func copyZnDecimal(old *ZnDecimal) *ZnDecimal {
	result, _ := duplicateValue(old).(*ZnDecimal)
	return result
}
