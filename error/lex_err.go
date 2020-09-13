package error

import (
	"fmt"
)

// InvalidSingleEllipsis -
func InvalidSingleEllipsis() *Error {
	return lexError.NewError(0x01, Error{
		text: "未能识别单个「…」字符，或许应该是「……」？",
	})
}

// InvalidSingleEqual -
func InvalidSingleEqual() *Error {
	return lexError.NewError(0x02, Error{
		text: "未能识别单个「=」字符，或许应该是「==」？",
	})
}

// DecodeUTF8Fail - decode error
func DecodeUTF8Fail(ch byte) *Error {
	return lexError.NewError(0x20, Error{
		text: fmt.Sprintf("前方有无法解析成UTF-8编码之异常字符'\\x%x'，请确认文件编码之正确性及完整性", ch),
		info: fmt.Sprintf("charcode=(%d)", ch),
	})
}

// InvalidIndentType -
func InvalidIndentType(expect uint8, got uint8) *Error {
	findName := func(idetType uint8) string {
		name := "「空格」"
		if idetType == uint8(9) { // TAB
			name = "「TAB」"
		}
		return name
	}
	return lexError.NewError(0x21, Error{
		text: fmt.Sprintf("此行现行缩进类型为%s，与前设缩进类型%s不符", findName(got), findName(expect)),
		info: fmt.Sprintf("expect=(%d) got=(%d)", expect, got),
	})
}

// InvalidIndentSpaceCount -
func InvalidIndentSpaceCount(count int) *Error {
	return lexError.NewError(0x22, Error{
		text: fmt.Sprintf("当缩进类型为「空格」，其所列字符数应为4之倍数：当前空格字符数为%d", count),
		info: fmt.Sprintf("count=(%d)", count),
	})
}

// QuoteStackFull -
func QuoteStackFull(maxSize int) *Error {
	return lexError.NewError(0x23, Error{
		text: fmt.Sprintf("在文本中嵌套过多引号：最大可以嵌套%d层", maxSize),
		info: fmt.Sprintf("maxsize=(%d)", maxSize),
	})
}

// InvalidIdentifier -
func InvalidIdentifier() *Error {
	return lexError.NewError(0x24, Error{
		text: "标识符不符合规范",
	})
}

// IdentifierExceedLength -
func IdentifierExceedLength(maxLen int32) *Error {
	return lexError.NewError(0x25, Error{
		text: fmt.Sprintf("标识符长度超过限制：最大可用长度为%d个字元", maxLen),
		info: fmt.Sprintf("maxlen=(%d)", maxLen),
	})
}

// InvalidChar -
func InvalidChar(ch rune) *Error {
	return lexError.NewError(0x26, Error{
		text: fmt.Sprintf("未能识别字元「%c」", ch),
		info: fmt.Sprintf("charcode=(%d)", ch),
	})
}
