package error

import "fmt"

// ArithDivZeroError - for A/B, when B = 0
func ArithDivZeroError() *Error {
	return arithError.NewError(0x01, Error{
		text: "被除数不得为0",
	})
}

// ParseFromStringError -
func ParseFromStringError(raw string) *Error {
	return arithError.NewError(0x02, Error{
		text: fmt.Sprintf("解析「%s」错误", raw),
	})
}

// ToIntegerError -
func ToIntegerError(raw string) *Error {
	return arithError.NewError(0x03, Error{
		text: fmt.Sprintf("转换 %s 成整数错误", raw),
	})
}

const (
	// ErrCodeArithDivZero -
	ErrCodeArithDivZero = (ArithErrorClass << 16) & 0x01
)
