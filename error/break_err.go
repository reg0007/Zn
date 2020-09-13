package error

// ReturnBreakError - breaks when return statement is executed
func ReturnBreakError(extra interface{}) *Error {
	return breakError.NewError(0x01, Error{
		text:  "未处理之「返回」中断",
		extra: extra,
	})
}

// ContinueBreakError - breaks when "此之（继续）" statement is executed
func ContinueBreakError() *Error {
	return breakError.NewError(0x02, Error{
		text: "未处理之「继续」中断",
	})
}

// BreakBreakError - breaks when "此之（结束）" statement fis executed
func BreakBreakError() *Error {
	return breakError.NewError(0x03, Error{
		text: "未处理之「结束」中断",
	})
}
