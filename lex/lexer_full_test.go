package lex

import (
	"testing"
)

type tokensCase struct {
	name        string
	input       string
	expectError bool
	tokens      string
	lines       string
}

// stringify token grammer:
// $type[literal] $type2[literal]
//
// example:
// $0[这是一个长长的单行注释] $1[引用一个文本]
func TestNextToken_MixedText(t *testing.T) {
	cases := []tokensCase{
		{
			name:        "empty data",
			input:       "",
			expectError: false,
			tokens:      "",
			lines:       "E<0>",
		},
		{
			name:        "1 number, 1 identifier",
			input:       `12.5rpm`,
			expectError: false,
			tokens:      `$4[12.5] $5[rpm]`,
			lines:       "U<0>[12.5rpm]",
		},
		{
			name:        "1 identifier with 1 inline comment",
			input:       `标识符名注：这是一个标识符啊 `,
			expectError: false,
			tokens:      `$5[标识符名] $10[注：这是一个标识符啊 ]`,
			lines:       "U<0>[标识符名注：这是一个标识符啊 ]",
		},
		{
			name:        "1 identifier (mixed number) with 1 inline comment",
			input:       `标识符名12注：这是一个标识符啊 `,
			expectError: false,
			tokens:      `$5[标识符名12] $10[注：这是一个标识符啊 ]`,
			lines:       "U<0>[标识符名12注：这是一个标识符啊 ]",
		},
		{
			name:        "1 identifier, 注 is not comment",
			input:       `起居注23不为其和`,
			expectError: false,
			tokens:      `$5[起居注23] $49[不为] $65[其] $5[和]`,
			lines:       "U<0>[起居注23不为其和]",
		},
		{
			name:        "identifer in keyword",
			input:       `令变量不为空`,
			expectError: false,
			tokens:      `$40[令] $5[变量] $49[不为] $5[空]`,
			lines:       "U<0>[令变量不为空]",
		},
		{
			name:        "1 identifier sep 1 number",
			input:       `变量1为12.45E+3`,
			expectError: false,
			tokens:      `$5[变量1] $41[为] $4[12.45E+3]`,
			lines:       "U<0>[变量1为12.45E+3]",
		},
		{
			name:        "comment 2 lines, one string",
			input:       "注：“可是都 \n  不为空”“是为”《淮南子》",
			expectError: false,
			tokens:      "$10[注：“可是都 \n  不为空”] $2[“是为”] $2[《淮南子》]",
			lines:       "U<0>[注：“可是都 ] U<0>[  不为空”“是为”《淮南子》]",
		},
		{
			name:        "nest multiple strings",
			input:       "·显然在其中·“不为空”‘为\n\n空’「「「随意“嵌套”」233」456」",
			expectError: false,
			tokens:      "$3[显然在其中] $2[“不为空”] $2[‘为\n\n空’] $2[「「「随意“嵌套”」233」456」]",
			lines:       "U<0>[·显然在其中·“不为空”‘为] E<0> U<0>[空’「「「随意“嵌套”」233」456」]",
		},
		{
			name:        "incomplete var quote at end",
			input:       "如何·显然在其中",
			expectError: false,
			tokens:      "$45[如何] $3[显然在其中]",
			lines:       "U<0>[如何·显然在其中]",
		},
		{
			name:        "consecutive keywords",
			input:       "以其为",
			expectError: false,
			tokens:      "$56[以] $65[其] $41[为]",
			lines:       "U<0>[以其为]",
		},
		{
			name:        "consecutive keywords #2",
			input:       "不以其为",
			expectError: false,
			tokens:      "$5[不] $56[以] $65[其] $41[为]",
			lines:       "U<0>[不以其为]",
		},
		{
			name:        "multi line string with var quote inside",
			input:       "“搞\n个\n    大新闻”《·焦点在哪里·》\n\t注：“又是一年\n    春来到”",
			expectError: false,
			tokens:      "$2[“搞\n个\n    大新闻”] $2[《·焦点在哪里·》] $10[注：“又是一年\n    春来到”]",
			lines:       "T<0>[“搞] T<0>[个] T<0>[    大新闻”《·焦点在哪里·》] T<1>[注：“又是一年] T<0>[    春来到”]",
		},
		{
			name:        "markers with spaces",
			input:       "\n    （  ） ， A/B  #  25",
			expectError: false,
			tokens:      "$22[（] $23[）] $11[，] $5[A/B] $18[#] $4[25]",
			lines:       "E<0> SP<1>[（  ） ， A/B  #  25]",
		},
		{
			name:        "keyword after line",
			input:       "令甲，乙为（【12，34，【“测试到底”，10】】）\n令丙为“23”",
			expectError: false,
			tokens: "$40[令] $5[甲] $11[，] $5[乙] $41[为] $22[（] $20[【] $4[12]" +
				" $11[，] $4[34] $11[，] $20[【] $2[“测试到底”] $11[，] $4[10] $21[】]" +
				" $21[】] $23[）] $40[令] $5[丙] $41[为] $2[“23”]",
			lines: "U<0>[令甲，乙为（【12，34，【“测试到底”，10】】）] U<0>[令丙为“23”]",
		},
		{
			name:        "multiple levels of indents",
			input:       "如何求证？\r\n    &小心为妙（2.2_23）；\n        _A**",
			expectError: false,
			tokens:      "$45[如何] $5[求证] $14[？] $15[&] $5[小心] $41[为] $5[妙] $22[（] $4[2.223] $23[）] $12[；] $5[_A**]",
			lines:       "SP<0>[如何求证？] SP<1>[&小心为妙（2.2_23）；] SP<2>[_A**]",
		},
		{
			name:        "array index",
			input:       "【23，45】#34 #{5}",
			expectError: false,
			tokens:      "$20[【] $4[23] $11[，] $4[45] $21[】] $18[#] $4[34] $27[#{] $4[5] $26[}]",
			lines:       "U<0>[【23，45】#34 #{5}]",
		},
	}

	assertTokens(cases, t)
}

func assertTokens(cases []tokensCase, t *testing.T) {
	for _, tt := range cases {

		lex := NewLexer(NewTextStream(tt.input))
		t.Run(tt.name, func(t *testing.T) {
			var tErr error
			var tokens = make([]*Token, 0)
			// iterate to get tokens
			for {
				tk, err := lex.NextToken()
				if err != nil {
					tErr = err
					break
				}
				if tk.Type == TypeEOF {
					break
				}
				tokens = append(tokens, tk)
			}
			// assert data
			if tt.expectError == false {
				if tErr != nil {
					t.Errorf("parse Tokens failed! expected no error, but got error")
					t.Error(tErr)
					return
				}

				// conform all tokens to string
				var actualStr = StringifyAllTokens(tokens)
				if actualStr != tt.tokens {
					t.Errorf("tokens not same!\nexpect->\n%s\ngot->\n%s", tt.tokens, actualStr)
				}

				lineInfo := StringifyLines(lex.LineStack)
				// compare line info
				if lineInfo != tt.lines {
					t.Errorf("line info not same!\nexpect->\n%s\ngot->\n%s", lineInfo, tt.lines)
				}

			} else {
				if tErr == nil {
					t.Errorf("NextToken() failed! expected error, but got no error")
				}
			}
		})
	}
}
