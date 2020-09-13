package error

import "fmt"

// IndexOutOfRange -
func IndexOutOfRange() *Error {
	return indexError.NewError(0x01, Error{
		text: "索引超出此对象可用范围",
	})
}

// IndexKeyNotFound - used in hashmap
func IndexKeyNotFound(key string) *Error {
	return indexError.NewError(0x02, Error{
		text: fmt.Sprintf("索引「%s」并不存在于此对象中", key),
		info: fmt.Sprintf("index=(%s)", key),
	})
}
