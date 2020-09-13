package lex

import (
	"fmt"
	"reflect"
	"testing"
)

type nextTokenCase struct {
	name        string
	input       string
	expectError bool
	token       Token
	errCursor   int
}

// mainly for testing parseCommentHead()
func TestNextToken_CommentsONLY(t *testing.T) {
	cases := []nextTokenCase{
		{
			name:        "singleLine comment",
			input:       "注：这是一个长 长 的单行注释comment",
			expectError: false,
			token: Token{
				Type:    TypeComment,
				Literal: []rune("注：这是一个长 长 的单行注释comment"),
			},
		},
		{
			name:        "singleLine empty comment",
			input:       "注：",
			expectError: false,
			token: Token{
				Type:    TypeComment,
				Literal: []rune("注："),
			},
		},
		{
			name:        "singleLine empty comment (single quote)",
			input:       "注： “",
			expectError: false,
			token: Token{
				Type:    TypeComment,
				Literal: []rune("注： “"),
			},
		},
		{
			name:        "singleLine empty comment (with number)",
			input:       "注 1024 2048 ：",
			expectError: false,
			token: Token{
				Type:    TypeComment,
				Literal: []rune("注 1024 2048 ："),
			},
		},
		{
			name:        "singleLine comment with newline",
			input:       "注：注：注：nach nach\r\n注：又是一个注",
			expectError: false,
			token: Token{
				Type:    TypeComment,
				Literal: []rune("注：注：注：nach nach"),
			},
		},
		//// multi-line comment
		{
			name:        "mutlLine comment with no new line",
			input:       "注：“假设这是一个注” 后面假设又是一些数",
			expectError: false,
			token: Token{
				Type:    TypeComment,
				Literal: []rune("注：“假设这是一个注”"),
			},
		},
		{
			name:        "mutlLine comment with no other string",
			input:       "注：“假设这又是一个注”",
			expectError: false,
			token: Token{
				Type:    TypeComment,
				Literal: []rune("注：“假设这又是一个注”"),
			},
		},
		{
			name:        "mutlLine comment (with number)",
			input:       "注 1234 5678 ：“假设这又是一个注”",
			expectError: false,
			token: Token{
				Type:    TypeComment,
				Literal: []rune("注 1234 5678 ：“假设这又是一个注”"),
			},
		},
		{
			name:        "mutlLine comment with empty string",
			input:       "注：“”",
			expectError: false,
			token: Token{
				Type:    TypeComment,
				Literal: []rune("注：“”"),
			},
		},
		{
			name:        "mutlLine comment with multiple lines",
			input:       "注：“一一\r\n    二二\n三三\n四四”",
			expectError: false,
			token: Token{
				Type:    TypeComment,
				Literal: []rune("注：“一一\r\n    二二\n三三\n四四”"),
			},
		},
		{
			name:        "mutlLine comment with quote stack",
			input:       "注：“一一「2233」《某本书》注：“”二二\n     ”",
			expectError: false,
			token: Token{
				Type:    TypeComment,
				Literal: []rune("注：“一一「2233」《某本书》注：“”二二\n     ”"),
			},
		},
		{
			name:        "mutlLine comment with straight quote",
			input:       "注：「PK」",
			expectError: false,
			token: Token{
				Type:    TypeComment,
				Literal: []rune("注：「PK」"),
			},
		},
		{
			name:        "mutlLine comment unfinished quote",
			input:       "注：「PKG“”",
			expectError: false,
			token: Token{
				Type:    TypeComment,
				Literal: []rune("注：「PKG“”"),
			},
		},
	}
	assertNextToken(cases, t)
}

func TestNextToken_StringONLY(t *testing.T) {
	cases := []nextTokenCase{
		{
			name:        "normal quote string",
			input:       "“LSK” 多出来的",
			expectError: false,
			token: Token{
				Type:    TypeString,
				Literal: []rune("“LSK”"),
			},
		},
		{
			name:        "normal quote string (with whitespaces)",
			input:       "“这 是 一 个 字 符 串”",
			expectError: false,
			token: Token{
				Type:    TypeString,
				Literal: []rune("“这 是 一 个 字 符 串”"),
			},
		},
		{
			name:        "normal quote string (with multiple quotes)",
			input:       "“「233」 ‘456’ 《〈who〉》『『is』』”",
			expectError: false,
			token: Token{
				Type:    TypeString,
				Literal: []rune("“「233」 ‘456’ 《〈who〉》『『is』』”"),
			},
		},
		{
			name:        "multiple-line string",
			input:       "『233\n    456\r\n7  』",
			expectError: false,
			token: Token{
				Type:    TypeString,
				Literal: []rune("『233\n    456\r\n7  』"),
			},
		},
	}

	assertNextToken(cases, t)
}

func TestNextToken_VarQuoteONLY(t *testing.T) {
	cases := []nextTokenCase{
		{
			name:        "normal variable quote",
			input:       "·正常之变量·",
			expectError: false,
			token: Token{
				Type:    TypeVarQuote,
				Literal: []rune("正常之变量"),
			},
		},
		{
			name:        "normal variable quote (with spaces)",
			input:       "· 正常 之 变量  ·",
			expectError: false,
			token: Token{
				Type:    TypeVarQuote,
				Literal: []rune("正常之变量"),
			},
		},
		{
			name:        "normal variable quote (with slashs)",
			input:       "· 知/其/不- 可/而*为+ _abcd_之1235 AJ·",
			expectError: false,
			token: Token{
				Type:    TypeVarQuote,
				Literal: []rune("知/其/不-可/而*为+_abcd_之1235AJ"),
			},
		},
		{
			name:        "normal variable quote - english variable",
			input:       "·_korea_char102·",
			expectError: false,
			token: Token{
				Type:    TypeVarQuote,
				Literal: []rune("_korea_char102"),
			},
		},
		{
			name:        "invalid quote - number at first",
			input:       "·123ABC·",
			expectError: true,
			errCursor:   1,
		},
		{
			name:        "invalid quote - invalid punctuation",
			input:       "·正（大）光明·",
			expectError: true,
			errCursor:   2,
		},
		{
			name:        "invalid quote - char buffer overflow",
			input:       "·这是一个很长变量这是一个很长变量这是一个很长变量这是一个很长变量这是一个很长变量·",
			expectError: true,
			errCursor:   33,
		},
		{
			name:        "invalid quote - CR, LFs are not allowed inside quotes",
			input:       "·变量\r\n又是变量名·",
			expectError: true,
			errCursor:   3,
		},
	}
	assertNextToken(cases, t)
}

func TestNextToken_NumberONLY(t *testing.T) {
	// NOTE 1:
	// nums such as 2..3 will be regarded as `2.`(2.0) and `.3`(0.3) combination
	cases := []nextTokenCase{
		{
			name:        "normal number (all digits)",
			input:       "123456七",
			expectError: false,
			token: Token{
				Type:    TypeNumber,
				Literal: []rune("123456"),
			},
		},
		{
			name:        "normal number (start to end)",
			input:       "1234567",
			expectError: false,
			token: Token{
				Type:    TypeNumber,
				Literal: []rune("1234567"),
			},
		},
		{
			name:        "normal number (with dot and minus)",
			input:       "-123.456km",
			expectError: false,
			token: Token{
				Type:    TypeNumber,
				Literal: []rune("-123.456"),
			},
		},
		{
			name:        "normal number (with plus at beginning)",
			input:       "+00000.456km",
			expectError: false,
			token: Token{
				Type:    TypeNumber,
				Literal: []rune("+00000.456"),
			},
		},
		{
			name:        "normal number (with plus)",
			input:       "+000003 Rs",
			expectError: false,
			token: Token{
				Type:    TypeNumber,
				Literal: []rune("+000003"),
			},
		},
		{
			name:        "normal number (with E+)",
			input:       "+000003E+05 Rs",
			expectError: false,
			token: Token{
				Type:    TypeNumber,
				Literal: []rune("+000003E+05"),
			},
		},
		{
			name:        "normal number (with e-)",
			input:       "+000003e-25 Rs",
			expectError: false,
			token: Token{
				Type:    TypeNumber,
				Literal: []rune("+000003e-25"),
			},
		},
		{
			name:        "normal number (decimal with e+)",
			input:       "-003.0452e+25 Rs",
			expectError: false,
			token: Token{
				Type:    TypeNumber,
				Literal: []rune("-003.0452e+25"),
			},
		},
		{
			name:        "normal number (ignore underscore)",
			input:       "-12_300_500_800_900 RSU",
			expectError: false,
			token: Token{
				Type:    TypeNumber,
				Literal: []rune("-12300500800900"),
			},
		},
		{
			name:        "*10^ as E",
			input:       "23.5*10^8",
			expectError: false,
			token: Token{
				Type:    TypeNumber,
				Literal: []rune("23.5*10^8"),
			},
		},
		// test fail cases
		{
			name:        "operater only",
			input:       "---",
			expectError: true,
			errCursor:   1,
		},
		{
			name:        "operater only #2",
			input:       "-++",
			expectError: true,
			errCursor:   1,
		},
		{
			name:        "E first",
			input:       "-E+3",
			expectError: true,
			errCursor:   1,
		},
		{
			name:        "E without following PM mark",
			input:       "2395.234E34",
			expectError: true,
			errCursor:   9,
		},
		{
			name:        "number with other weird char",
			input:       "23.r",
			expectError: true,
			errCursor:   3,
		},
		{
			name:        "numbers *9^",
			input:       "1111*9^23",
			expectError: true,
			errCursor:   5,
		},
		{
			name:        "incomplete *10^",
			input:       "1234*10^",
			expectError: true,
			errCursor:   8,
		},
	}

	assertNextToken(cases, t)
}

func TestNextToken_MarkerONLY(t *testing.T) {
	// 01. generate TRUE cases
	var markerMap = map[string]TokenType{
		"，":  TypeCommaSep,
		"：":  TypeFuncCall,
		"；":  TypeStmtSep,
		"？":  TypeFuncDeclare,
		"&":  TypeObjRef,
		"！":  TypeMustT,
		"@":  TypeAnnoT,
		"#":  TypeMapHash,
		"……": TypeMoreParam,
		"【":  TypeArrayQuoteL,
		"】":  TypeArrayQuoteR,
		"（":  TypeFuncQuoteL,
		"）":  TypeFuncQuoteR,
		"{":  TypeStmtQuoteL,
		"}":  TypeStmtQuoteR,
		"==": TypeMapData,
		"⟺":  TypeMapData,
	}

	var cases = make([]nextTokenCase, 0)
	for k, v := range markerMap {
		cases = append(cases, nextTokenCase{
			name:        fmt.Sprintf("generate token %s", k),
			input:       fmt.Sprintf("%s EE", k),
			expectError: false,
			token: Token{
				Type:    v,
				Literal: []rune(k),
			},
		})
	}

	assertNextToken(cases, t)
}

func TestNextToken_IdentifierONLY_SUCCESS(t *testing.T) {
	cases := []nextTokenCase{
		{
			name:        "normal identifier",
			input:       "反",
			expectError: false,
			token: Token{
				Type:    TypeIdentifier,
				Literal: []rune("反"),
			},
		},
		{
			name:        "normal identifier #2",
			input:       "正定县",
			expectError: false,
			token: Token{
				Type:    TypeIdentifier,
				Literal: []rune("正定县"),
			},
		},
		{
			name:        "normal identifier #3 with spaces",
			input:       "正  定  县",
			expectError: false,
			token: Token{
				Type:    TypeIdentifier,
				Literal: []rune("正定县"),
			},
		},
		{
			name:        "normal identifier with number followed",
			input:       "正定县2345",
			expectError: false,
			token: Token{
				Type:    TypeIdentifier,
				Literal: []rune("正定县2345"),
			},
		},
		{
			name:        "normal identifier with + - * /",
			input:       "正定/+_县/2345",
			expectError: false,
			token: Token{
				Type:    TypeIdentifier,
				Literal: []rune("正定/+_县/2345"),
			},
		},
		{
			name:        "normal identifier (quote as terminator)",
			input:       "正定县「」",
			expectError: false,
			token: Token{
				Type:    TypeIdentifier,
				Literal: []rune("正定县"),
			},
		},
		{
			name:        "normal identifier (var quote as terminator)",
			input:       "正定县·如果·",
			expectError: false,
			token: Token{
				Type:    TypeIdentifier,
				Literal: []rune("正定县"),
			},
		},
		{
			name:        "normal identifier (marker) as terminator)",
			input:       "正定县（河北）",
			expectError: false,
			token: Token{
				Type:    TypeIdentifier,
				Literal: []rune("正定县"),
			},
		},
		{
			name:        "normal identifier (keyword as terminator)",
			input:       "正定县成为大县",
			expectError: false,
			token: Token{
				Type:    TypeIdentifier,
				Literal: []rune("正定县"),
			},
		},
		{
			name:        "normal identifier (following keyword lead but not keyword formed)",
			input:       "正定县如大县",
			expectError: false,
			token: Token{
				Type:    TypeIdentifier,
				Literal: []rune("正定县如大县"),
			},
		},
		{
			name:        "normal identifier (like keyword)",
			input:       "如不果返回",
			expectError: false,
			token: Token{
				Type:    TypeIdentifier,
				Literal: []rune("如不果"),
			},
		},
	}
	assertNextToken(cases, t)
}

func assertNextToken(cases []nextTokenCase, t *testing.T) {
	for _, tt := range cases {
		lex := NewLexer(NewBufferStream([]byte(tt.input)))
		t.Run(tt.name, func(t *testing.T) {
			tk, err := lex.NextToken()
			// validate error
			if tt.expectError == false {
				if err != nil {
					t.Errorf("NextToken() failed! expected no error, but got error")
					t.Error(err)
				} else if !reflect.DeepEqual(tk.Type, tt.token.Type) || !reflect.DeepEqual(tk.Literal, tt.token.Literal) {
					t.Errorf("NextToken() return Token failed! expect: %v, got: %v", tt.token, *tk)
				}
			} else {
				if err == nil {
					t.Errorf("NextToken() failed! expected error, but got no error")
				} else if err.GetCursor().ColNum != tt.errCursor {
					t.Errorf("Err cursor location not match! expect: %d, got: %d", tt.errCursor, err.GetCursor().ColNum)
				}
			}
		})
	}
}
