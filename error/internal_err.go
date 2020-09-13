package error

import "fmt"

// Internal Error Class, for Zn Internal exception (rare to happen)
// e.g. Unexpected switch-case

// UnExpectedCase -
func UnExpectedCase(tag string, value string) *Error {
	return internalError.NewError(0x01, Error{
		text: fmt.Sprintf("未定义的条件项：「%s」的值为「%s」", tag, value),
	})
}
