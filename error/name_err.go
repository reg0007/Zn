package error

import "fmt"

// NameNotDefined -
func NameNotDefined(name string) *Error {
	return nameError.NewError(0x01, Error{
		text: fmt.Sprintf("标识「%s」未有定义", name),
		info: fmt.Sprintf("name=(%s)", name),
	})
}

// NameRedeclared -
func NameRedeclared(name string) *Error {
	return nameError.NewError(0x02, Error{
		text: fmt.Sprintf("标识「%s」被重复定义", name),
		info: fmt.Sprintf("name=(%s)", name),
	})
}

// AssignToConstant -
func AssignToConstant() *Error {
	return nameError.NewError(0x03, Error{
		text: "不允许赋值给常变量",
	})
}

// PropertyNotFound -
func PropertyNotFound(name string) *Error {
	return nameError.NewError(0x04, Error{
		text: fmt.Sprintf("未找到属性「%s」", name),
		info: fmt.Sprintf("name=(%s)", name),
	})
}

// MethodNotFound -
func MethodNotFound(name string) *Error {
	return nameError.NewError(0x05, Error{
		text: fmt.Sprintf("未找到方法名「%s」", name),
		info: fmt.Sprintf("name=(%s)", name),
	})
}
