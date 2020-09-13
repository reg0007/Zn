package error

import (
	"fmt"
	"io"
)

// FileNotFound - file not found
func FileNotFound(path string) *Error {
	info := fmt.Sprintf("path=(%s)", path)
	return lexError.NewError(0x10, Error{
		text: fmt.Sprintf("未能找到文件 %s，请检查它是否存在！", path),
		info: info,
	})
}

// FileOpenError -
func FileOpenError(filePath string, oriError error) *Error {
	info := fmt.Sprintf("path=(%s) error=(%s)", filePath, oriError)
	return lexError.NewError(0x11, Error{
		text: fmt.Sprintf("未能读取文件 %s，请检查其是否存在及有无读取权限！", filePath),
		info: info,
	})
}

// ReadFileError -
func ReadFileError(e error) *Error {
	errTextMap := map[error]string{
		io.ErrShortBuffer:   "需要更大的缓冲区",
		io.ErrUnexpectedEOF: "未知文件结束符",
		io.ErrNoProgress:    "多次尝试读取，皆无数据或返回错误",
		io.ErrShortWrite:    "操作写入的数据比提供的少",
	}

	errText := e.Error()
	if v, ok := errTextMap[e]; ok {
		errText = fmt.Sprintf("%s (%s)", v, e.Error())
	}
	return lexError.NewError(0x12, Error{
		text: fmt.Sprintf("读取I/O流失败：%s！", errText),
		info: fmt.Sprintf("error=(%s)", errText),
	})
}
