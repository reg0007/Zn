package lex

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// StringifyToken is the process that transforms an abstract token type to a readable string sequence.
// usually used for debugging.
func StringifyToken(tk *Token) string {
	return stringifyToken(tk)
}

// StringifyAllTokens - similar with StringifyToken, but this time it's for all tokens
func StringifyAllTokens(tks []*Token) string {
	var tokenStrs = []string{}
	for _, tk := range tks {
		tokenStrs = append(tokenStrs, stringifyToken(tk))
	}
	return strings.Join(tokenStrs, " ")
}

// ParseTokenStr - from token str to tokens
func ParseTokenStr(str string) []Token {
	tks := make([]Token, 0)
	r := regexp.MustCompile(`\$(\d+)\[(.+?)\]`)
	matches := r.FindAllStringSubmatch(str, -1)

	lineCursor := 1
	colCursor := 0
	for _, match := range matches {
		n, _ := strconv.Atoi(match[1])
		l := []rune(match[2])
		tks = append(tks, Token{
			Type:    TokenType(n),
			Literal: l,
			Range: TokenRange{
				StartLine: lineCursor,
				StartIdx:  colCursor,
				EndLine:   lineCursor,
				EndIdx:    colCursor + len(l) + 1,
			},
		})

		colCursor = colCursor + len(l) + 1
	}

	return tks
}

// StringifyLines - stringify current parsed lines into readable string info
// format::=
//   {lineInfo1} {lineInfo2} {lineInfo3} ...
//
// lineInfo ::=
//   SP<2>[text1] or
//   T<4>[text2] or
//   E<0>
func StringifyLines(ls *LineStack) string {
	ss := []string{}
	var indentChar string
	// get indent type
	switch ls.IndentType {
	case IdetUnknown:
		indentChar = "U"
	case IdetSpace:
		indentChar = "SP"
	case IdetTab:
		indentChar = "T"
	}

	for i, line := range ls.lines {
		if line.startIdx == line.endIdx {
			ss = append(ss, fmt.Sprintf("E<%d>", line.indents))
		} else {
			var text = ls.GetLineText(i+1, false)
			ss = append(ss,
				fmt.Sprintf("%s<%d>[%s]", indentChar, line.indents, string(text)))
		}
	}
	return strings.Join(ss, " ")
}

func stringifyToken(tk *Token) string {
	return fmt.Sprintf("$%d[%s]", tk.Type, string(tk.Literal))
}
