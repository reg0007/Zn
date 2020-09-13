package lex

import (
	"github.com/reg0007/Zn/error"
	"github.com/reg0007/Zn/util"
)

// LineStack - store line source and its indent info
type LineStack struct {
	IndentType
	CurrentLine int
	in          *InputStream
	lines       []LineInfo
	scanCursor
	lineBuffer []rune
}

// LineInfo - stores the (absolute) start & end index of this line
// this should be added to the scanner after all parsing is done
type LineInfo struct {
	// the indent number (at the beginning) of this line.
	// all lines should have indents to differentiate scopes.
	indents int
	// startIdx - start index of lineBuffer
	startIdx int
	// endIdx - end index of lineBuffer
	// NOTE: endIdx value IS DIFFERENT FROM the index of last char of that line.
	// usually, endIdx = lastChar + 1
	endIdx int
}

type scanCursor struct {
	startIdx int
	indents  int
	scanState
}

// IndentType - only TAB or SPACE (U+0020) are supported
type IndentType = uint8

// scanState - internal line scan state
type scanState uint8

// define IndentTypes
const (
	IdetUnknown IndentType = 0
	IdetTab     IndentType = 9
	IdetSpace   IndentType = 32
)

// define scanStates
// example:
//
// [ IDENTS ] [ TEXT TEXT TEXT ] [ CR LF ]
// ^         ^                  ^
// 0         1                  2
//
// 0: scanInit
// 1: scanIndent
// 2: scanEnd
const (
	scanInit   scanState = 0
	scanIndent scanState = 1
	scanEnd    scanState = 2
)

// NewLineStack - new line stack
func NewLineStack(in *InputStream) *LineStack {
	return &LineStack{
		IndentType:  IdetUnknown,
		lines:       []LineInfo{},
		CurrentLine: 1,
		in:          in,
		scanCursor:  scanCursor{0, 0, scanIndent},
		lineBuffer:  []rune{},
	}
}

// SetIndent - set current line's indent (for counting the consecutive intent chars, it's the task of lexer)
// notice: for IndentType = SPACE, only 4 * N chars as indent is valid!
// and change scanState from 0 -> 1
//
// possible errors:
// 1. inconsist indentType
// 2. when IndentType = SPACE, the count is not 4 * N chars
func (ls *LineStack) SetIndent(count int, t IndentType) *error.Error {
	switch t {
	case IdetUnknown:
		if count > 0 && ls.IndentType != t {
			return error.InvalidIndentType(ls.IndentType, t)
		}
	case IdetSpace, IdetTab:
		// init ls.IndentType
		if ls.IndentType == IdetUnknown {
			ls.IndentType = t
		}
		// when t = space, the character count must be 4 * N
		if t == IdetSpace && count%4 != 0 {
			return error.InvalidIndentSpaceCount(count)
		}
		// when t does not match indentType, throw error
		if ls.IndentType != t {
			return error.InvalidIndentType(ls.IndentType, t)
		}
	}
	// when indentType = TAB, count = indents
	// otherwise, count = indents * 4
	indentNum := count
	if ls.IndentType == IdetSpace {
		indentNum = count / 4
	}

	// set scanCursor
	ls.scanCursor.indents = indentNum
	ls.scanCursor.scanState = scanIndent
	ls.scanCursor.startIdx += count
	return nil
}

// PushLine - push effective line text into line info
// change scanState from 1 -> 2
func (ls *LineStack) PushLine(lastIndex int) {
	idets := ls.scanCursor.indents

	// push index
	line := LineInfo{
		indents:  idets,
		startIdx: ls.scanCursor.startIdx,
		endIdx:   lastIndex,
	}

	ls.lines = append(ls.lines, line)
	ls.scanCursor.scanState = scanEnd
}

// NewLine - reset scanCursor
// change scanState from 2 -> 0
func (ls *LineStack) NewLine(index int) {
	// reset start index
	ls.scanCursor = scanCursor{index, 0, scanInit}

	// add CurrentLine
	ls.CurrentLine++
}

// AppendLineBuffer - push data to lineBuffer
func (ls *LineStack) AppendLineBuffer(data []rune) {
	ls.lineBuffer = append(ls.lineBuffer, data...)
}

// GetLineIndent - get current lineNum indent
// NOTE: lineNum starts from 1
// NOTE2: if lineNum not found, return -1
func (ls *LineStack) GetLineIndent(lineNum int) int {
	switch util.Compare(lineNum, ls.CurrentLine) {
	case 1: // a > b
		return -1
	case 0:
		return ls.indents
	default:
		if lineNum > 0 {
			lineInfo := ls.lines[lineNum-1]
			return lineInfo.indents
		}
		return -1
	}
}

// GetLineText -
func (ls *LineStack) GetLineText(lineNum int, slideToLineEnd bool) string {
	return string(ls.getLineRune(lineNum, slideToLineEnd))
}

// GetLineRune -
func (ls *LineStack) GetLineRune(lineNum int, slideToLineEnd bool) []rune {
	return ls.getLineRune(lineNum, slideToLineEnd)
}

// GetLineColumn -
func (ls *LineStack) GetLineColumn(lineNum int, cursor int) int {
	return ls.getLineColumn(lineNum, cursor)
}

//// private helpers
func (ls *LineStack) getCurrentLine() int {
	return ls.CurrentLine
}

func (ls *LineStack) getLineBufferSize() int {
	return len(ls.lineBuffer)
}

// getChar - get value from lineBuffer
func (ls *LineStack) getChar(idx int) rune {
	if idx >= len(ls.lineBuffer) {
		return EOF
	}
	return ls.lineBuffer[idx]
}

// getTextFromIdx - get text from absolute index range of line buffer
func (ls *LineStack) getTextFromIdx(startIdx int, endIdx int) []rune {
	return ls.lineBuffer[startIdx:endIdx]
}

// getLineRune - get text from desinated line number.
// There're two situations need to handle:
//
// I. lineNum < currentLine. line info MUST have been completely parsed
// and stored in ls.lineBuffer. Thus it's easy to fetch by index.
//
// II. lineNum = currentLine. Here, line info MAY NOT parsed completely. (For instance, when a token throws TokenError
// in the middle of line where the rest part of line may not been read by inputStream at all!)
// Therefore, when `slideToLineEnd = true`, it will attempt to read characters afterwards until meets CR, LF, EOF;
// when `slideToLineEnd = false`, get line text until the last char of lineBuffer
//
// III. lineNum > currentLine, return empty []rune{}
func (ls *LineStack) getLineRune(lineNum int, slideToLineEnd bool) []rune {
	switch util.Compare(lineNum, ls.CurrentLine) {
	case 1: // a > b
		return []rune{}
	case 0: // a == b
		sIdx := ls.scanCursor.startIdx
		eIdx := sIdx
		for eIdx < len(ls.lineBuffer) && !util.Contains(ls.lineBuffer[eIdx], []rune{CR, LF, EOF}) {
			eIdx++
			if eIdx >= len(ls.lineBuffer) {
				if !slideToLineEnd {
					break
				}
				// usually the length of a line won't exceed 512 chars, thus for most
				// cases, only one fetch is enough.
				if !ls.in.End() {
					if b, err := ls.in.Read(256); err == nil {
						ls.lineBuffer = append(ls.lineBuffer, b...)
					} else {
						// throw the error globally
						panic(err)
					}
				}
			}
		}
		return ls.lineBuffer[sIdx:eIdx]
	default: // a < b
		if lineNum > 0 {
			info := ls.lines[lineNum-1]
			return ls.getTextFromIdx(info.startIdx, info.endIdx)
		}
		return []rune{}
	}
}

func (ls *LineStack) getLineColumn(lineNum int, cursor int) int {
	countIndentChars := func(indent int) int {
		if ls.IndentType == IdetSpace {
			return indent * 4
		}
		return indent
	}
	switch util.Compare(lineNum, ls.CurrentLine) {
	case -1: // a < b
		if lineNum > 0 {
			line := ls.lines[lineNum-1]
			return cursor - line.startIdx + countIndentChars(line.indents)
		}
		return -1
	case 0:
		return cursor - ls.scanCursor.startIdx + countIndentChars(ls.scanCursor.indents)
	default:
		return -1
	}
}

// onIndentStage - if the incoming SPACE/TAB should be regarded as indents
// or normal spaces.
func (ls *LineStack) onIndentStage() bool {
	return ls.scanState == scanInit
}
