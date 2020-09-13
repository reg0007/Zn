package syntax

import (
	"github.com/reg0007/Zn/error"
	"github.com/reg0007/Zn/lex"
)

// Parser - parse all nodes
type Parser struct {
	*lex.Lexer
	tokens       [3]*lex.Token
	lineTermFlag bool
}

const (
	modeInline uint16 = 0x01
)

// NewParser -
func NewParser(l *lex.Lexer) *Parser {
	return &Parser{
		Lexer:        l,
		lineTermFlag: false,
	}
}

// Parse - parse all tokens into an AST (stored as ProgramNode)
func (p *Parser) Parse() (block *BlockStmt, err *error.Error) {
	defer func() {
		var ok bool
		if r := recover(); r != nil {
			err, ok = r.(*error.Error)
			// for other kinds of error (e.g. runtime error), panic it directly
			if !ok {
				panic(r)
			}
		}
		handleDeferError(p, err)
	}()

	// advance tokens TWICE
	p.next()
	p.next()

	peekIndent := p.getPeekIndent()
	// parse global block
	block = ParseBlockStmt(p, peekIndent)

	// ensure there's no remaining token after parsing global block
	if p.peek().Type != lex.TypeEOF {
		err = error.UnexpectedEOF()
	}
	return
}

func (p *Parser) next() *lex.Token {
	var tk *lex.Token
	var err *error.Error

	tk, err = p.NextToken()
	if err != nil {
		panic(err)
	}

	// move advanced token buffer
	p.tokens[0] = p.tokens[1]
	p.tokens[1] = p.tokens[2]
	p.tokens[2] = tk

	return p.tokens[0]
}

func (p *Parser) current() *lex.Token {
	return p.tokens[0]
}

func (p *Parser) peek() *lex.Token {
	return p.tokens[1]
}

func (p *Parser) peek2() *lex.Token {
	return p.tokens[2]
}

// meetStmtLineBreak - if there's a statement line-break at the end of token.
//
// StmtLineBreak is not an existing token, i.e. it wouldn't insert into token stream.
// It's a virtual mark that separates parsing statements, thus you can image it like a
// "virtually" inserted semicolon, like the following code:
//
// 如果此IDE之 ';' <-- here is the statement line-break
//     名为「VSCODE」
//
// There's stmt line-break at the end of first line, thus the process of parsing IF-statement
// should terminate due to lacking matched tokens, similiar to the behaviour that an semicolon is inserted.
//
// Theoratically, any type of token that located at the end of line
//
//   i.e.  this token is the last one of current line,
//   or    $token.Range.EndLine < ($token+1).Range.StartLine,
//
// should meet StmtLineBreak, which means meetStmtLineBreak() = true. Still, there're some
// exceptions listed below:
//
//   1.    token type is one of the following 5 punctuations:  ，  {  【  ：  ？
//   or
//   2.    the next token type if one of the following 3 marks:  】  }  EOF
//
// Example 1#, 2#, 3# illustrates those exceptions that even if there're two or more lines, it's still
// a valid and complete statement:
//
// Example 1#
//
// （显示并连缀：
//     「结果为」，
//      人口-中位数）
//
// Example 2#
//
// 令天干地支表为【
//     「子」 == 「甲」，
//     「丑」 == 「乙」，
//     「寅」 == 「丙」，
//     「卯」 == 「丁」
// 】
//
// Example 3#
//
// ·时·等于12 且{
//     ·分·等于0 或 ·分·等于30
// }等于真
func (p *Parser) meetStmtLineBreak() bool {
	current := p.current()
	peek := p.peek()

	exceptCurrentTokenTypes := []lex.TokenType{
		lex.TypeCommaSep,
		lex.TypeStmtQuoteL,
		lex.TypeArrayQuoteL,
		lex.TypeFuncCall,
		lex.TypeFuncDeclare,
	}

	exceptFollowingTokenTypes := []lex.TokenType{
		lex.TypeArrayQuoteR,
		lex.TypeStmtQuoteR,
	}

	if peek == nil || current == nil {
		return false
	}
	if current.Type == lex.TypeEOF || peek.Type == lex.TypeEOF {
		return false
	}

	// current token is at line end
	if peek.Range.StartLine > current.Range.EndLine {
		// exception rule 1
		for _, currTk := range exceptCurrentTokenTypes {
			if currTk == current.Type {
				return false
			}
		}
		// exception rule 2
		for _, followingTk := range exceptFollowingTokenTypes {
			if followingTk == peek.Type {
				return false
			}
		}
		return true
	}
	return false
}

func (p *Parser) resetLineTermFlag() {
	p.lineTermFlag = false
}

func (p *Parser) setLineTermFlag() {
	p.lineTermFlag = true
}

// consume one token with denoted validTypes
// if not, return syntaxError
func (p *Parser) consume(validTypes ...lex.TokenType) {
	if p.meetStmtLineBreak() && p.lineTermFlag {
		panic(error.InvalidSyntaxCurr())
	}

	tkType := p.peek().Type
	for _, item := range validTypes {
		if item == tkType {
			p.setLineTermFlag()
			p.next()
			return
		}
	}
	err := error.InvalidSyntax()
	panic(err)
}

// trying to consume one token. if the token is valid in the given range of tokenTypes,
// will return its tokenType; if not, then nothing will happen.
//
// returns (matched, tokenType)
func (p *Parser) tryConsume(validTypes ...lex.TokenType) (bool, *lex.Token) {
	if p.meetStmtLineBreak() && p.lineTermFlag {
		return false, nil
	}
	tk := p.peek()
	for _, vt := range validTypes {
		if vt == tk.Type {
			p.setLineTermFlag()
			p.next()
			return true, tk
		}
	}

	return false, nil
}

// expectBlockIndent - detect if the Indent(peek) == Indent(current) + 1
// returns (validBlockIndent, newIndent)
func (p *Parser) expectBlockIndent() (bool, int) {
	var peekLine = p.peek().Range.StartLine
	var currLine = p.current().Range.StartLine

	var peekIndent = p.GetLineIndent(peekLine)
	var currIndent = p.GetLineIndent(currLine)

	if peekIndent == currIndent+1 {
		return true, peekIndent
	}
	return false, 0
}

// getPeekIndent -
func (p *Parser) getPeekIndent() int {
	var peekLine = p.peek().Range.StartLine

	return p.GetLineIndent(peekLine)
}

//// helper functions

// similar to lexer's version, but with given line & col
func moveAndSetCursor(p *Parser, tk *lex.Token, err *error.Error) {
	line := tk.Range.StartLine
	cursor := error.Cursor{
		File:    p.Lexer.InputStream.GetFile(),
		ColNum:  p.GetLineColumn(line, tk.Range.StartIdx),
		LineNum: tk.Range.StartLine,
		Text:    p.GetLineText(line, true),
	}

	err.SetCursor(cursor)
}

func handleDeferError(p *Parser, err *error.Error) {
	var tk *lex.Token

	if err != nil && err.GetErrorClass() == error.SyntaxErrorClass {
		if cursorType, ok := err.GetInfo()["cursor"]; ok {
			if cursorType == "peek" {
				tk = p.peek()
			} else if cursorType == "current" {
				tk = p.current()
			}
			if tk != nil {
				moveAndSetCursor(p, tk, err)
			}
		}
	}
}
