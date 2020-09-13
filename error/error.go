package error

import (
	"fmt"
	"regexp"
	"strings"
)

// Error model
type Error struct {
	code   uint16
	text   string
	cursor Cursor
	// custom field to store additional data
	// info string format: key1=(value1) key2=(value2)
	// example: path=(/tmp/new-data) code=(2000)
	info string
	// extra data (any type)
	extra interface{}

	displayMask uint16
}

// RawError <--> builtin 'error' type
// this is to resolve naming confliction.
//
// usage:
// OLD:    func A(int) error
// NEW:    func A(int) RawError
type RawError interface {
	Error() string
}

// Error - display error text
func (e *Error) Error() string {
	return e.text
}

// SetCursor - set error occurance location
// for better readability
func (e *Error) SetCursor(cursor Cursor) {
	e.cursor = cursor
}

// GetCursor - get error cursor
func (e *Error) GetCursor() Cursor {
	return e.cursor
}

// GetCode - get error code
func (e *Error) GetCode() uint16 {
	return e.code
}

// GetExtra - get extra info
func (e *Error) GetExtra() interface{} {
	return e.extra
}

// GetErrorClass - get error class
func (e *Error) GetErrorClass() int {
	return int(e.code >> 8)
}

// GetInfo - get (parsed) info
func (e *Error) GetInfo() map[string]string {
	var infoMap = map[string]string{}

	items := strings.Split(e.info, " ")
	r := regexp.MustCompile(`^(\w+)=\((.+)\)`)

	for _, item := range items {
		match := r.FindStringSubmatch(item)
		if len(match) > 0 {
			infoMap[match[1]] = match[2]
		}
	}

	return infoMap
}

// Display - display detailed error info to user
// general format:
//
// 在 [FILE] 中，位于第 [LINE] 行：
//     [ LINE TEXT WITHOUT INDENTS AND CRLF ]
// ‹[ERRCODE]› [ERRCLASS]：[ERRTEXT]
//
// example error:
//
// 在 draft/example.zn 中，位于第 12 行：
//     如果代码不为空：
//    ^
// ‹2021› 语法错误：此行现行缩进类型为「TAB」，与前设缩进类型「空格」不符！
func (e *Error) Display() string {
	var line1, line2, line3, line4 string
	// line1
	if e.onMask(dpHideFileName) {
		if e.onMask(dpHideLineNum) {
			line1 = "发现错误："
		} else {
			line1 = fmt.Sprintf("在第 %d 行发现错误：", e.cursor.LineNum)
		}
	} else if e.onMask(dpHideLineNum) {
		line1 = fmt.Sprintf("在「%s」中发现错误：", e.cursor.File)
	} else {
		line1 = fmt.Sprintf("在「%s」中，位于第 %d 行发现错误：", e.cursor.File, e.cursor.LineNum)
	}
	// line2
	if e.onMask(dpHideLineText) {
		line2 = ""
	} else {
		line2 = fmt.Sprintf("    %s", e.cursor.Text)
	}
	// line3
	if e.onMask(dpHideLineText) || e.onMask(dpHideLineCursor) {
		line3 = ""
		if !e.onMask(dpHideLineText) {
			line3 = "    "
		}
	} else {
		line3 = fmt.Sprintf("   %s^", strings.Repeat(" ", calcCursorOffset(e.cursor.Text, e.cursor.ColNum)+1))
	}
	// line4
	if e.onMask(dpHideErrClass) {
		line4 = e.text
	} else {
		errClassText := fmt.Sprintf("‹%04X› %s", e.code, errClassMap[e.code>>8])
		line4 = fmt.Sprintf("%s：%s", errClassText, e.text)
	}

	lines := []string{line1, line2, line3, line4}
	texts := []string{}
	for _, line := range lines {
		if line != "" {
			texts = append(texts, line)
		}
	}
	return strings.Join(texts, "\n")
}

func calcCursorOffset(text string, col int) int {
	if col < 0 {
		return col
	}
	widthBorders := []int32{
		126, 159, 687, 710, 711, 727, 733, 879, 1154, 1161,
		4347, 4447, 7467, 7521, 8369, 8426, 9000, 9002, 11021, 12350,
		12351, 12438, 12442, 19893, 19967, 55203, 63743, 64106, 65039, 65059,
		65131, 65279, 65376, 65500, 65510, 120831, 262141, 1114109,
	}

	widths := []int{
		1, 0, 1, 0, 1, 0, 1, 0, 1, 0,
		1, 2, 1, 0, 1, 0, 1, 2, 1, 2,
		1, 2, 0, 2, 1, 2, 1, 2, 1, 0,
		2, 1, 2, 1, 2, 1, 2, 1,
	}

	offsets := 0

	getOffset := func(t rune) int {
		if t == 0xE || t == 0xF {
			return 0
		}
		for idx, b := range widthBorders {
			if t <= b {
				return widths[idx]
			}
		}
		return 1
	}
	for _, t := range []rune(text)[:col] {
		offsets = offsets + getOffset(t)
	}

	return offsets
}

func (e *Error) onMask(mask uint16) bool {
	return (e.displayMask & mask) > 0
}

// declare display masks
//                    16 8 4 2 1
// X X X X X X X X X X O O O O O
const (
	dpHideFileName   uint16 = 0x0001
	dpHideLineCursor uint16 = 0x0002
	dpHideLineNum    uint16 = 0x0004
	dpHideLineText   uint16 = 0x0008
	dpHideErrClass   uint16 = 0x0010
)

// Cursor denotes the indicator where the error occurs
type Cursor struct {
	File    string
	LineNum int
	ColNum  int
	Text    string
}

// ErrorClass defines the prefix of error code
type errorClass struct {
	prefix   uint16
	baseMask uint16
}

// NewError - new error with subcode
func (ec *errorClass) NewError(subcode uint16, model Error) *Error {
	var code uint16
	// code = prefix << 8 + subcode
	code = ec.prefix
	code = code*256 + subcode

	model.code = code
	model.displayMask = model.displayMask | ec.baseMask
	return &model
}

// define some error classes
const (
	LexErrorClass      = 0x20
	IOErrorClass       = 0x21
	SyntaxErrorClass   = 0x22
	TypeErrorClass     = 0x23
	IndexErrorClass    = 0x24
	NameErrorClass     = 0x25
	ArithErrorClass    = 0x26
	ParamErrorClass    = 0x27
	BreakErrorClass    = 0x50
	InternalErrorClass = 0x60
)

// NewErrorSLOT - a tmp placeholder for adding errors quickly while the
// details has not been thought carefully.
// 简单来说就是临时错误加个坑位，等到正式写代码的时候再用
func NewErrorSLOT(text string) *Error {
	return &Error{
		code: 0xFFFE,
		text: text,
	}
}

// definitions of all error classes inside the Zn Programming language
var (
	// 0x20 - lexError
	// this error class displays all errors occur during lexing stage. (including input data)
	lexError = errorClass{LexErrorClass, 0}
	// 0x21 - ioError
	// I/O related error (e.g. FileNotFound, OpenFileError, ReadFileError)
	ioError = errorClass{
		IOErrorClass,
		dpHideLineText | dpHideFileName | dpHideLineCursor | dpHideLineNum,
	}
	// 0x22 - syntaxError
	// show errors occured on parsing stage
	syntaxError = errorClass{SyntaxErrorClass, 0}
	// 0x23 - typeError
	// show errors of type validation
	typeError = errorClass{TypeErrorClass, dpHideLineCursor}
	// 0x24 - indexError
	// show errors when retrieving data by index from arrayList
	indexError = errorClass{IndexErrorClass, dpHideLineCursor}
	// 0x25 - nameError
	// show errors while identifier not found or others related to identifiers.
	nameError = errorClass{NameErrorClass, dpHideLineCursor}
	// 0x26 - arithError
	// trigger error happened on arithmetic operation (such as div)
	arithError = errorClass{ArithErrorClass, dpHideLineCursor}
	// 0x27 - paramError
	// trigger error when input parameters doesn't satisfy the requirements
	paramError = errorClass{ParamErrorClass, dpHideLineCursor}
	// 0x50 - breakError
	// send a virtual BREAK interrupt to stop the process.
	// NOTICE: breakError is NOT a true error!
	breakError = errorClass{BreakErrorClass, dpHideLineCursor}
	// 0x60 - internalError
	// Zn interpreter internal logic error, rarly happen
	internalError = errorClass{InternalErrorClass, dpHideLineCursor}

	errClassMap = map[uint16]string{
		LexErrorClass:      "语法错误", // from lex
		IOErrorClass:       "I/O错误",
		SyntaxErrorClass:   "语法错误", // from parser
		TypeErrorClass:     "类型错误",
		IndexErrorClass:    "索引错误",
		NameErrorClass:     "标识错误",
		ArithErrorClass:    "算术错误",
		ParamErrorClass:    "参数错误",
		BreakErrorClass:    "中断信号",
		InternalErrorClass: "内部错误",
	}
)

// declare some "breaks" error code
const (
	ReturnBreakSignal   = 0x5001
	ContinueBreakSignal = 0x5002
	BreakBreakSignal    = 0x5003
)
