package lex

import (
	"fmt"

	"github.com/reg0007/Zn/error"
	"github.com/reg0007/Zn/util"
)

const (
	defaultBlockSize int = 512
)

// Lexer is a structure that pe provides a set of tools to help tokenizing the code.
type Lexer struct {
	*InputStream // input stream
	*LineStack
	quoteStack *util.RuneStack
	chBuffer   []rune // the buffer for parsing & generating tokens
	cursor     int
	blockSize  int
	beginLex   bool
}

// NewLexer - new lexer
func NewLexer(in *InputStream) *Lexer {
	return &Lexer{
		LineStack:   NewLineStack(in),
		quoteStack:  util.NewRuneStack(32),
		InputStream: in,
		chBuffer:    []rune{},
		cursor:      -1,
		blockSize:   defaultBlockSize,
		beginLex:    true,
	}
}

// TokenType - general token type
type TokenType int

// Token - general token type
type Token struct {
	Type    TokenType
	Literal []rune
	Range   TokenRange
}

// TokenRange locates the absolute position of a token
type TokenRange struct {
	// startLine - line num (start from 1) of first char
	StartLine int
	StartIdx  int
	// endLine - line num (start from 1) of last char
	EndLine int
	EndIdx  int
}

// newTokenRange creates new TokenRange struct
// with startLine & startIdx initialized.
func newTokenRange(l *Lexer) TokenRange {
	return TokenRange{
		StartLine: l.getCurrentLine(),
		StartIdx:  l.cursor,
		EndLine:   l.getCurrentLine(),
		EndIdx:    l.cursor,
	}
}

func (r *TokenRange) setRangeEnd(l *Lexer) {
	r.EndLine = l.CurrentLine
	r.EndIdx = l.cursor + 1
}

// GetStartLine -
func (r *TokenRange) GetStartLine() int {
	return r.StartLine
}

// GetEndLine -
func (r *TokenRange) GetEndLine() int {
	return r.EndLine
}

//// 0. EOF

// EOF - mark as end of file, should only exists at the end of sequence
const EOF rune = 0

//// 1. keywords
// MOVE to lex/keyword.go FILE, check there for more details

//// 2. markers
// declare marks
const (
	Comma             rune = 0xFF0C //，
	Colon             rune = 0xFF1A //：
	Semicolon         rune = 0xFF1B //；
	QuestionMark      rune = 0xFF1F //？
	RefMark           rune = 0x0026 // &
	BangMark          rune = 0xFF01 // ！
	AnnotationMark    rune = 0x0040 // @
	HashMark          rune = 0x0023 // #
	EllipsisMark      rune = 0x2026 // …
	LeftBracket       rune = 0x3010 // 【
	RightBracket      rune = 0x3011 // 】
	LeftParen         rune = 0xFF08 // （
	RightParen        rune = 0xFF09 // ）
	Equal             rune = 0x003D // =
	DoubleArrow       rune = 0x27FA // ⟺
	LeftCurlyBracket  rune = 0x007B // {
	RightCurlyBracket rune = 0x007D // }
)

// MarkLeads -
var MarkLeads = []rune{
	Comma, Colon, Semicolon, QuestionMark, RefMark, BangMark,
	AnnotationMark, HashMark, EllipsisMark, LeftBracket,
	RightBracket, LeftParen, RightParen, Equal, DoubleArrow,
	LeftCurlyBracket, RightCurlyBracket,
}

//// 3. spaces
const (
	SP  rune = 0x0020 // <SP>
	TAB rune = 0x0009 // <TAB>
	CR  rune = 0x000D // \r
	LF  rune = 0x000A // \n
)

// WhiteSpaces - all kinds of valid spaces
var WhiteSpaces = []rune{
	// where 0x0020 <--> SP
	0x0009, 0x000B, 0x000C, 0x0020, 0x00A0,
	0x2000, 0x2001, 0x2002, 0x2003, 0x2004,
	0x2005, 0x2006, 0x2007, 0x2008, 0x2009,
	0x200A, 0x200B, 0x202F, 0x205F, 0x3000,
}

// helpers
func isWhiteSpace(ch rune) bool {
	for _, whiteSpace := range WhiteSpaces {
		if ch == whiteSpace {
			return true
		}
	}

	return false
}

//// 4. quotes
// declare quotes
const (
	LeftQuoteI    rune = 0x300A //《
	RightQuoteI   rune = 0x300B // 》
	LeftQuoteII   rune = 0x300C // 「
	RightQuoteII  rune = 0x300D // 」
	LeftQuoteIII  rune = 0x300E // 『
	RightQuoteIII rune = 0x300F // 』
	LeftQuoteIV   rune = 0x201C // “
	RightQuoteIV  rune = 0x201D // ”
	LeftQuoteV    rune = 0x2018 // ‘
	RightQuoteV   rune = 0x2019 // ’
)

// LeftQuotes -
var LeftQuotes = []rune{
	LeftQuoteI,
	LeftQuoteII,
	LeftQuoteIII,
	LeftQuoteIV,
	LeftQuoteV,
}

// RightQuotes -
var RightQuotes = []rune{
	RightQuoteI,
	RightQuoteII,
	RightQuoteIII,
	RightQuoteIV,
	RightQuoteV,
}

// QuoteMatchMap -
var QuoteMatchMap = map[rune]rune{
	LeftQuoteI:   RightQuoteI,
	LeftQuoteII:  RightQuoteII,
	LeftQuoteIII: RightQuoteIII,
	LeftQuoteIV:  RightQuoteIV,
	LeftQuoteV:   RightQuoteV,
}

//// 5. var quote
const (
	MiddleDot rune = 0x00B7 // ·
)

//// 6. numbers
func isNumber(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

//// 7. identifiers
const maxIdentifierLength = 32

// @params: ch - input char
// @params: isFirst - is the first char of identifier
func isIdentifierChar(ch rune, isFirst bool) bool {
	// CJK unified ideograph
	if ch >= 0x4E00 && ch <= 0x9FFF {
		return true
	}
	// 〇, _
	if ch == 0x3007 || ch == '_' {
		return true
	}
	// A-Z
	if ch >= 'A' && ch <= 'Z' {
		return true
	}
	if ch >= 'a' && ch <= 'z' {
		return true
	}
	if !isFirst {
		if ch >= '0' && ch <= '9' {
			return true
		}
		if util.Contains(ch, []rune{'*', '+', '-', '/'}) {
			return true
		}
	}
	return false
}

//// token consts and constructors (without keyword token)
// token types -
// for special type Tokens, its range varies from 0 - 9
// for keyword types, check lex/keyword.go for details
const (
	TypeEOF        TokenType = 0
	TypeSpace      TokenType = 1 // 空格类Token 备用
	TypeString     TokenType = 2 // 字符串
	TypeVarQuote   TokenType = 3
	TypeNumber     TokenType = 4 // 数值
	TypeIdentifier TokenType = 5 // 标识符

	TypeComment     TokenType = 10 // 注：
	TypeCommaSep    TokenType = 11 // ，
	TypeStmtSep     TokenType = 12 // ；
	TypeFuncCall    TokenType = 13 // ：
	TypeFuncDeclare TokenType = 14 // ？
	TypeObjRef      TokenType = 15 // &
	TypeMustT       TokenType = 16 // ！
	TypeAnnoT       TokenType = 17 // @
	TypeMapHash     TokenType = 18 // #
	TypeMoreParam   TokenType = 19 // ……
	TypeArrayQuoteL TokenType = 20 // 【
	TypeArrayQuoteR TokenType = 21 // 】
	TypeFuncQuoteL  TokenType = 22 // （
	TypeFuncQuoteR  TokenType = 23 // ）
	TypeMapData     TokenType = 24 // ==
	TypeStmtQuoteL  TokenType = 25 // {
	TypeStmtQuoteR  TokenType = 26 // }
	TypeMapQHash    TokenType = 27 // #{
)

// next - return current rune, and move forward the cursor for 1 character.
func (l *Lexer) next() rune {
	l.cursor++

	if l.cursor+2 >= l.getLineBufferSize() {
		if !l.End() {
			if b, err := l.Read(l.blockSize); err == nil {
				l.AppendLineBuffer(b)
			} else {
				// throw the error globally
				// it will be handled (recovered) in NextToken(),
				// similiar with C++'s try-catch statement.
				panic(err)
			}
		}
	}

	// still no data, return EOF directly
	return l.getChar(l.cursor)
}

// peek - get the character of the cursor
func (l *Lexer) peek() rune {
	return l.getChar(l.cursor + 1)
}

// peek2 - get the next next character without moving the cursor
func (l *Lexer) peek2() rune {
	return l.getChar(l.cursor + 2)
}

// rebase - rebase cursor within the same line
func (l *Lexer) rebase(cursor int) {
	l.cursor = cursor
}

func (l *Lexer) clearBuffer() {
	l.chBuffer = []rune{}
}

func (l *Lexer) pushBuffer(ch ...rune) {
	l.chBuffer = append(l.chBuffer, ch...)
}

// SetBlockSize -
func (l *Lexer) SetBlockSize(size int) {
	l.blockSize = size
}

// NextToken - parse and generate the next token (including comments)
func (l *Lexer) NextToken() (tok *Token, err *error.Error) {
	defer func() {
		var ok bool
		if r := recover(); r != nil {
			err, ok = r.(*error.Error)
			// for other kinds of error (e.g. runtime error), panic it directly
			if !ok {
				panic(r)
			}
		}
		handleDeferError(l, err)
	}()

	// For the first line, we use some tricks to determine if this line
	// contains indents
	if l.beginLex {
		l.beginLex = false
		if !util.Contains(l.peek(), []rune{SP, TAB, EOF}) {
			if err := l.SetIndent(0, IdetUnknown); err != nil {
				return nil, err
			}
		}
	}
head:
	var ch = l.next()
	switch ch {
	case EOF:
		l.PushLine(l.cursor)
		tok = NewTokenEOF(l.CurrentLine, l.cursor)
		return
	case SP, TAB:
		// if indent has been scanned, it should be regarded as whitespaces
		// (it's totally ignored)
		if l.onIndentStage() {
			l.parseIndents(ch)
		} else {
			l.consumeWhiteSpace(ch)
		}
		goto head
	case CR, LF:
		l.parseCRLF(ch)
		goto head
	// meet with 注, it may be possibly a lead character of a comment block
	// notice: it would also be a normal identifer (if 注[number]：) does not satisfy.
	case GlyphZHU:
		cursor := l.cursor
		rg := newTokenRange(l)
		isComment, isMultiLine, note := l.validateComment(ch)
		if isComment {
			tok, err = l.parseComment(l.getChar(l.cursor), isMultiLine, note, rg)
			return
		}

		l.rebase(cursor)
		// goto normal identifier
	// left quotes
	case LeftQuoteI, LeftQuoteII, LeftQuoteIII, LeftQuoteIV, LeftQuoteV:
		tok, err = l.parseString(ch)
		return
	case MiddleDot:
		tok, err = l.parseVarQuote(ch)
		return
	default:
		// skip whitespaces
		if isWhiteSpace(ch) {
			l.consumeWhiteSpace(ch)
			goto head
		}
		// parse number
		if isNumber(ch) || util.Contains(ch, []rune{'.', '+', '-'}) {
			tok, err = l.parseNumber(ch)
			return
		}
		if util.Contains(ch, MarkLeads) {
			tok, err = l.parseMarkers(ch)
			return
		}
		// suppose it's a keyword
		if isKeyword, tk := l.parseKeyword(ch, true); isKeyword {
			tok = tk
			return
		}
	}
	tok, err = l.parseIdentifier(ch)
	return
}

func handleDeferError(l *Lexer, err *error.Error) {
	if err != nil {
		if err.GetErrorClass() == error.IOErrorClass {
			// For I/O error, load current line buffer directly
			// instead of moving cursor to line end (since it's impossible to retrieve line end)
			err.SetCursor(error.Cursor{
				File:    l.InputStream.GetFile(),
				LineNum: l.CurrentLine,
				Text:    l.GetLineText(l.CurrentLine, false),
				ColNum:  0,
			})
		} else {
			l.moveAndSetCursor(err)
		}
	}
}

//// parsing logics
func (l *Lexer) parseIndents(ch rune) {
	count := 1
	for l.peek() == ch {
		count++
		l.next()
	}
	// determine indentType
	indentType := IdetUnknown
	switch ch {
	case TAB:
		indentType = IdetTab
	case SP:
		indentType = IdetSpace
	}
	if err := l.SetIndent(count, indentType); err != nil {
		panic(err)
	}
}

// parseCRLF and return the newline chars by the way
func (l *Lexer) parseCRLF(ch rune) []rune {
	var rtn = []rune{}
	p := l.peek()
	// for CRLF <windows type> or LFCR
	if (ch == CR && p == LF) || (ch == LF && p == CR) {
		// skip one char since we have judge two chars
		l.next()
		l.PushLine(l.cursor - 1)

		rtn = []rune{ch, p}
	} else {
		// for LF or CR only
		// LF: <linux>, CR:<old mac>
		l.PushLine(l.cursor)
		rtn = []rune{ch}
	}

	// new line and reset cursor
	l.NewLine(l.cursor + 1)

	// to see if next line contains (potential) indents
	if !util.Contains(l.peek(), []rune{SP, TAB, EOF}) {
		if err := l.SetIndent(0, IdetUnknown); err != nil {
			panic(err)
		}
	}
	return rtn
}

// validate if the coming block is a comment block
// valid comment block are listed below:
// (single-line)
// 1. 注：
// 2. 注123456：
//
// (multi-line)
//
// 3. 注：“
// 4. 注123456：“
//
// @returns (isValid, isMultiLine)
func (l *Lexer) validateComment(ch rune) (bool, bool, []rune) {
	note := []rune{}
	// “ or 「
	lquotes := []rune{LeftQuoteIV, LeftQuoteII}
	for {
		ch = l.next()
		// match pattern 1, 2
		if ch == Colon {
			// match pattern 3, 4
			if util.Contains(l.peek(), lquotes) {
				l.next()
				return true, true, note
			}
			return true, false, note
		}
		if isNumber(ch) || isWhiteSpace(ch) {
			note = append(note, ch)
			continue
		}
		return false, false, note
	}
}

// parseComment until its end
func (l *Lexer) parseComment(ch rune, isMultiLine bool, note []rune, rg TokenRange) (*Token, *error.Error) {
	// setup
	l.clearBuffer()
	if isMultiLine {
		if !l.quoteStack.Push(ch) {
			return nil, error.QuoteStackFull(l.quoteStack.GetMaxSize())
		}
		l.pushBuffer(ch)
	}
	// iterate
	for {
		ch = l.next()
		switch ch {
		case EOF:
			l.rebase(l.cursor - 1)
			rg.setRangeEnd(l)
			return NewCommentToken(l.chBuffer, note, rg), nil
		case CR, LF:
			// parse CR,LF first
			nl := l.parseCRLF(ch)
			if !isMultiLine {
				return NewCommentToken(l.chBuffer, note, rg), nil
			}
			// for multi-line comment blocks, CRLF is also included
			l.pushBuffer(nl...)

			// manually set no indents
			if err := l.SetIndent(0, IdetUnknown); err != nil {
				return nil, err
			}
		default:
			// for mutli-line comment, calculate quotes is necessary.
			if isMultiLine {
				// push left quotes
				if util.Contains(ch, LeftQuotes) {
					if !l.quoteStack.Push(ch) {
						return nil, error.QuoteStackFull(l.quoteStack.GetMaxSize())
					}
				}
				// pop right quotes if possible
				if util.Contains(ch, RightQuotes) {
					currentL, _ := l.quoteStack.Current()
					if QuoteMatchMap[currentL] == ch {
						l.quoteStack.Pop()
					}
					// stop quoting
					if l.quoteStack.IsEmpty() {
						rg.setRangeEnd(l)
						l.pushBuffer(ch)
						return NewCommentToken(l.chBuffer, note, rg), nil
					}
				}
			}
			l.pushBuffer(ch)
			// cache rangeEnd location as the position of last char
			// Example: (for single-line comment token)
			// 注： ABCDEFG \r\n
			//
			// Here, the rangeEnd should be on char `G`, but it will stop iff the following `\r\n` is parsed!
			rg.setRangeEnd(l)
		}
	}
}

// parseString -
func (l *Lexer) parseString(ch rune) (*Token, *error.Error) {
	// start up
	l.clearBuffer()
	l.quoteStack.Push(ch)
	firstChar := ch
	rg := newTokenRange(l)

	// iterate
	for {
		ch := l.next()
		switch ch {
		case EOF:
			l.rebase(l.cursor - 1)
			// after meeting with EOF
			rg.setRangeEnd(l)
			return NewStringToken(l.chBuffer, firstChar, rg), nil
		// push quotes
		case LeftQuoteI, LeftQuoteII, LeftQuoteIII, LeftQuoteIV, LeftQuoteV:
			l.pushBuffer(ch)
			if !l.quoteStack.Push(ch) {
				return nil, error.QuoteStackFull(l.quoteStack.GetMaxSize())
			}
		// pop quotes if match
		case RightQuoteI, RightQuoteII, RightQuoteIII, RightQuoteIV, RightQuoteV:
			currentL, _ := l.quoteStack.Current()
			if QuoteMatchMap[currentL] == ch {
				l.quoteStack.Pop()
			}
			// stop quoting
			if l.quoteStack.IsEmpty() {
				rg.setRangeEnd(l)
				return NewStringToken(l.chBuffer, firstChar, rg), nil
			}
			l.pushBuffer(ch)
		case CR, LF:
			nl := l.parseCRLF(ch)
			// push buffer & mark new line
			l.pushBuffer(nl...)
			if err := l.SetIndent(0, IdetUnknown); err != nil {
				return nil, err
			}
		default:
			l.pushBuffer(ch)
		}
	}
}

func (l *Lexer) parseVarQuote(ch rune) (*Token, *error.Error) {
	// setup
	l.clearBuffer()
	rg := newTokenRange(l)
	// iterate
	count := 0
	for {
		ch = l.next()
		// we should ensure the following chars to satisfy the condition
		// of an identifier
		switch ch {
		case EOF:
			l.rebase(l.cursor - 1)
			rg.setRangeEnd(l)
			return NewVarQuoteToken(l.chBuffer, rg), nil
		case MiddleDot:
			rg.setRangeEnd(l)
			return NewVarQuoteToken(l.chBuffer, rg), nil
		default:
			// ignore white-spaces!
			if isWhiteSpace(ch) {
				continue
			}
			if isIdentifierChar(ch, count == 0) {
				l.pushBuffer(ch)
				count++
				if count > maxIdentifierLength {
					return nil, error.IdentifierExceedLength(maxIdentifierLength)
				}
			} else {
				return nil, error.InvalidIdentifier()
			}
		}
	}
}

// regex: ^[-+]?[0-9]*\.?[0-9]+((([eE][-+])|(\*(10)?\^[-+]?))[0-9]+)?$
// ref: https://github.com/reg0007/Zn/issues/4
func (l *Lexer) parseNumber(ch rune) (*Token, *error.Error) {
	// setup
	l.clearBuffer()
	rg := newTokenRange(l)

	// hand-written regex parser
	// ref: https://cyberzhg.github.io/toolbox/min_dfa?regex=Rj9QP0QqLj9EKygoKEVQKXwocygxMCk/dVA/KSlEKyk/
	// hand-drawn min-DFA:
	// https://github.com/reg0007/Zn/issues/6
	const (
		sBegin      = 1
		sDot        = 2
		sIntEnd     = 3
		sIntPMFlag  = 5
		sDotDecEnd  = 6
		sEFlag      = 7
		sSFlag      = 8
		sExpPMFlag  = 9
		sSciI       = 10
		sSciEndFlag = 11
		sExpEnd     = 12
		sSciII      = 13
	)
	var state = sBegin
	var endStates = []int{sIntEnd, sDotDecEnd, sExpEnd}

	for {
		switch ch {
		case EOF:
			goto end
		case 'e', 'E':
			switch state {
			case sDotDecEnd, sIntEnd:
				state = sEFlag
			default:
				goto end
			}
		case '.':
			switch state {
			case sBegin, sIntPMFlag, sIntEnd:
				state = sDot
			default:
				goto end
			}
		case '-', '+':
			switch state {
			case sBegin:
				state = sIntPMFlag
			case sEFlag, sSciEndFlag:
				state = sExpPMFlag
			default:
				goto end
			}
		case '_':
			ch = l.next()
			continue
		case '*':
			switch state {
			case sDotDecEnd, sIntEnd:
				state = sSFlag
			default:
				goto end
			}
		case '1':
			switch state {
			case sSFlag:
				state = sSciI
				// same with other numbers
			case sBegin, sIntEnd, sIntPMFlag:
				state = sIntEnd
			case sDot, sDotDecEnd:
				state = sDotDecEnd
			case sExpPMFlag, sSciEndFlag, sExpEnd:
				state = sExpEnd
			default:
				goto end
			}
		case '0':
			switch state {
			case sSciI:
				state = sSciII
			case sBegin, sIntEnd, sIntPMFlag:
				state = sIntEnd
			case sDot, sDotDecEnd:
				state = sDotDecEnd
			case sExpPMFlag, sSciEndFlag, sExpEnd:
				state = sExpEnd
			default:
				goto end
			}
		case '2', '3', '4', '5', '6', '7', '8', '9':
			switch state {
			case sBegin, sIntEnd, sIntPMFlag:
				state = sIntEnd
			case sDot, sDotDecEnd:
				state = sDotDecEnd
			case sExpPMFlag, sSciEndFlag, sExpEnd:
				state = sExpEnd
			default:
				goto end
			}
		case '^':
			switch state {
			case sSFlag, sSciII:
				state = sSciEndFlag
			default:
				goto end
			}
		default:
			goto end
		}
		l.pushBuffer(ch)
		ch = l.next()
	}

end:
	if util.ContainsInt(state, endStates) {
		// back to last available char
		l.rebase(l.cursor - 1)
		rg.setRangeEnd(l)
		return NewNumberToken(l.chBuffer, rg), nil
	}
	return nil, error.InvalidChar(ch)
}

// parseMarkers -
func (l *Lexer) parseMarkers(ch rune) (*Token, *error.Error) {
	// setup
	l.clearBuffer()
	l.pushBuffer(ch)

	startR := newTokenRange(l)
	// switch
	switch ch {
	case Comma:
		return NewMarkToken(l.chBuffer, TypeCommaSep, startR, 1), nil
	case Colon:
		return NewMarkToken(l.chBuffer, TypeFuncCall, startR, 1), nil
	case Semicolon:
		return NewMarkToken(l.chBuffer, TypeStmtSep, startR, 1), nil
	case QuestionMark:
		return NewMarkToken(l.chBuffer, TypeFuncDeclare, startR, 1), nil
	case RefMark:
		return NewMarkToken(l.chBuffer, TypeObjRef, startR, 1), nil
	case BangMark:
		return NewMarkToken(l.chBuffer, TypeMustT, startR, 1), nil
	case AnnotationMark:
		return NewMarkToken(l.chBuffer, TypeAnnoT, startR, 1), nil
	case HashMark:
		if l.peek() == LeftCurlyBracket {
			l.pushBuffer(l.next())
			return NewMarkToken(l.chBuffer, TypeMapQHash, startR, 2), nil
		}
		return NewMarkToken(l.chBuffer, TypeMapHash, startR, 1), nil
	case EllipsisMark:
		if l.peek() == EllipsisMark {
			l.pushBuffer(l.next())
			return NewMarkToken(l.chBuffer, TypeMoreParam, startR, 2), nil
		}
		return nil, error.InvalidSingleEllipsis()
	case LeftBracket:
		return NewMarkToken(l.chBuffer, TypeArrayQuoteL, startR, 1), nil
	case RightBracket:
		return NewMarkToken(l.chBuffer, TypeArrayQuoteR, startR, 1), nil
	case LeftParen:
		return NewMarkToken(l.chBuffer, TypeFuncQuoteL, startR, 1), nil
	case RightParen:
		return NewMarkToken(l.chBuffer, TypeFuncQuoteR, startR, 1), nil
	case LeftCurlyBracket:
		return NewMarkToken(l.chBuffer, TypeStmtQuoteL, startR, 1), nil
	case RightCurlyBracket:
		return NewMarkToken(l.chBuffer, TypeStmtQuoteR, startR, 1), nil
	case Equal:
		if l.peek() == Equal {
			l.pushBuffer(l.next())
			return NewMarkToken(l.chBuffer, TypeMapData, startR, 2), nil
		}
		return nil, error.InvalidSingleEqual()
	case DoubleArrow:
		return NewMarkToken(l.chBuffer, TypeMapData, startR, 1), nil
	}
	return nil, error.InvalidChar(ch)
}

// for (l *Lexer) parseKeyword(), CHECK lex/keyword.go for details

// consume (and skip) whitespaces
func (l *Lexer) consumeWhiteSpace(ch rune) {
	for isWhiteSpace(l.peek()) {
		l.next()
	}
}

// parseIdentifier
func (l *Lexer) parseIdentifier(ch rune) (*Token, *error.Error) {
	// setup
	l.clearBuffer()
	var count = 0
	var terminators = append([]rune{
		EOF, CR, LF, LeftQuoteI, LeftQuoteII, LeftQuoteIII,
		LeftQuoteIV, LeftQuoteV, MiddleDot,
	}, MarkLeads...)

	if !isIdentifierChar(ch, true) {
		return nil, error.InvalidIdentifier()
	}

	rg := newTokenRange(l)
	// range.EndIdx = range.StartIdx + 1
	// i.e. range at least include one char
	rg.setRangeEnd(l)
	// push first char
	l.pushBuffer(ch)
	count++
	// iterate
	for {
		prev := l.cursor
		ch = l.next()

		if isWhiteSpace(ch) {
			continue
		}
		// if the following chars are a keyword,
		// then terminate the identifier parsing process.
		if isKeyword, _ := l.parseKeyword(ch, false); isKeyword {
			l.rebase(prev)
			rg.setRangeEnd(l)
			return NewIdentifierToken(l.chBuffer, rg), nil
		}
		// parse 注
		if ch == GlyphZHU {
			if validComment, _, _ := l.validateComment(ch); validComment {
				l.rebase(prev)
				rg.setRangeEnd(l)
				return NewIdentifierToken(l.chBuffer, rg), nil
			}
			l.rebase(prev + 1)
		}
		// other char as terminator
		if util.Contains(ch, terminators) {
			l.rebase(prev)
			return NewIdentifierToken(l.chBuffer, rg), nil
		}

		if isIdentifierChar(ch, false) {
			if count >= maxIdentifierLength {
				return nil, error.IdentifierExceedLength(maxIdentifierLength)
			}
			l.pushBuffer(ch)
			rg.setRangeEnd(l)
			count++
			continue
		}
		return nil, error.InvalidChar(ch)
	}
}

// moveAndSetCursor - retrieve full text of the line and set the current cursor
// to display errors
func (l *Lexer) moveAndSetCursor(err *error.Error) {
	cursor := error.Cursor{
		File:    l.InputStream.GetFile(),
		ColNum:  l.cursor - l.scanCursor.startIdx,
		LineNum: l.CurrentLine,
		Text:    l.GetLineText(l.CurrentLine, true),
	}
	err.SetCursor(cursor)
}

// NewTokenEOF - new EOF token
func NewTokenEOF(line int, col int) *Token {
	return &Token{
		Type:    TypeEOF,
		Literal: []rune{},
		Range: TokenRange{
			StartLine: line,
			StartIdx:  col,
			EndLine:   line,
			EndIdx:    col,
		},
	}
}

// NewStringToken -
func NewStringToken(buf []rune, quoteType rune, rg TokenRange) *Token {
	literal := append([]rune{quoteType}, util.Copy(buf)...)
	switch quoteType {
	case LeftQuoteI:
		literal = append(literal, RightQuoteI)
	case LeftQuoteII:
		literal = append(literal, RightQuoteII)
	case LeftQuoteIII:
		literal = append(literal, RightQuoteIII)
	case LeftQuoteIV:
		literal = append(literal, RightQuoteIV)
	case LeftQuoteV:
		literal = append(literal, RightQuoteV)
	}
	return &Token{
		Type:    TypeString,
		Literal: literal,
		Range:   rg,
	}
}

// NewVarQuoteToken -
func NewVarQuoteToken(buf []rune, rg TokenRange) *Token {
	return &Token{
		Type:    TypeVarQuote,
		Literal: util.Copy(buf),
		Range:   rg,
	}
}

// NewCommentToken -
func NewCommentToken(buf []rune, note []rune, rg TokenRange) *Token {
	prefix := fmt.Sprintf("注%s：", string(note))
	literal := append([]rune(prefix), util.Copy(buf)...)
	return &Token{
		Type:    TypeComment,
		Literal: literal,
		Range:   rg,
	}
}

// NewNumberToken -
func NewNumberToken(buf []rune, rg TokenRange) *Token {
	return &Token{
		Type:    TypeNumber,
		Literal: util.Copy(buf),
		Range:   rg,
	}
}

// NewMarkToken -
func NewMarkToken(buf []rune, t TokenType, startR TokenRange, num int) *Token {
	rg := startR
	rg.EndLine = startR.StartLine
	rg.EndIdx = startR.StartIdx + num
	return &Token{
		Type:    t,
		Literal: util.Copy(buf),
		Range:   rg,
	}
}

// NewKeywordToken -
func NewKeywordToken(t TokenType) *Token {
	var l = []rune{}
	if item, ok := KeywordTypeMap[t]; ok {
		l = item
	}
	return &Token{
		Type:    t,
		Literal: l,
	}
}

// NewIdentifierToken -
func NewIdentifierToken(buf []rune, rg TokenRange) *Token {
	return &Token{
		Type:    TypeIdentifier,
		Literal: util.Copy(buf),
		Range:   rg,
	}
}
