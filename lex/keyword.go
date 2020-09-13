package lex

//// keyword character (ideoglyphs) definition
// keywords are all ideoglyphs that its length varies from its definitions.
// so here we define all possible chars that may be an element of one keyword.
const (
	// GlyphBU - 不 - 不等于，不小于，不大于，不为
	GlyphBU rune = 0x4E0D
	// GlyphQIE - 且 - 且
	GlyphQIE rune = 0x4E14
	// GlyphWEI - 为 - 是为，成为，恒为，何为，为，不为
	GlyphWEI rune = 0x4E3A
	// GlyphYIy - 义 - 定义
	GlyphYIy rune = 0x4E49
	// GlyphZHI - 之 - 此之，之
	GlyphZHI rune = 0x4E4B
	// GlyphYU - 于 - 等于，小于，大于，不等于，不小于，不大于
	GlyphYU rune = 0x4E8E
	// GlyphLING - 令 - 令
	GlyphLING rune = 0x4EE4
	// GlyphYIi - 以 - 以
	GlyphYIi rune = 0x4EE5
	// GlyphHE - 何 - 如何，何为
	GlyphHE rune = 0x4F55
	// GlyphQI - 其 - 其
	GlyphQI rune = 0x5176
	// GlyphZAI - 再 - 再如
	GlyphZAI rune = 0x518D
	// GlyphZE - 则 - 否则
	GlyphZE rune = 0x5219
	// GlyphLI - 历 - 遍历
	GlyphLI rune = 0x5386
	// GlyphFOU - 否 - 否则
	GlyphFOU rune = 0x5426
	// GlyphHUI - 回 - 返回
	GlyphHUI rune = 0x56DE
	// GlyphDA - 大 - 大于，不大于
	GlyphDA rune = 0x5927
	// GlyphRU - 如 - 如果，如何，再如
	GlyphRU rune = 0x5982
	// GlyphDING - 定 - 定义
	GlyphDING rune = 0x5B9A
	// GlyphXIAO - 小 - 小于，不小于
	GlyphXIAO rune = 0x5C0F
	// GlyphYI - 已 - 已知
	GlyphYI rune = 0x5DF2
	// GlyphDANG - 当 - 每当
	GlyphDANG rune = 0x5F53
	// GlyphHENG - 恒 - 恒为
	GlyphHENG rune = 0x6052
	// GlyphCHENG - 成 - 成为
	GlyphCHENG rune = 0x6210
	// GlyphHUO - 或 - 或
	GlyphHUO rune = 0x6216
	// GlyphSHI - 是 - 是为
	GlyphSHI rune = 0x662F
	// GlyphGUO - 果 - 如果
	GlyphGUO rune = 0x679C
	// GlyphCI - 此 - 此之
	GlyphCI rune = 0x6B64
	// GlyphMEI - 每 - 每当
	GlyphMEI rune = 0x6BCF
	// GlyphZHU - 注 -
	GlyphZHU rune = 0x6CE8
	// GlyphZHIy - 知 - 已知
	GlyphZHIy rune = 0x77E5
	// GlyphDENG - 等 - 等于，不等于
	GlyphDENG rune = 0x7B49
	// GlyphFAN - 返 - 返回
	GlyphFAN rune = 0x8FD4
	// GlyphBIAN - 遍 - 遍历
	GlyphBIAN rune = 0x904D
)

// KeywordLeads - all glyphs that would be possible of the first character of one keyword.
var KeywordLeads = []rune{
	GlyphBU, GlyphQIE, GlyphWEI,
	GlyphZHI, GlyphLING, GlyphYIi,
	GlyphHE, GlyphQI, GlyphZAI,
	GlyphFOU, GlyphDA, GlyphRU,
	GlyphDING, GlyphXIAO, GlyphYI,
	GlyphHENG, GlyphCHENG, GlyphHUO,
	GlyphSHI, GlyphCI, GlyphMEI,
	GlyphDENG, GlyphFAN, GlyphBIAN,
}

// Keyword token types
const (
	TypeDeclareW      TokenType = 40 // 令
	TypeLogicYesW     TokenType = 41 // 为
	TypeAssignConstW  TokenType = 42 // 恒为
	TypeCondOtherW    TokenType = 43 // 再如
	TypeCondW         TokenType = 44 // 如果
	TypeFuncW         TokenType = 45 // 如何
	TypeGetterW       TokenType = 46 // 何为
	TypeParamAssignW  TokenType = 47 // 已知
	TypeReturnW       TokenType = 48 // 返回
	TypeLogicNotW     TokenType = 49 // 不为
	TypeLogicNotEqW   TokenType = 51 // 不等于
	TypeLogicLteW     TokenType = 52 // 不大于
	TypeLogicGteW     TokenType = 53 // 不小于
	TypeLogicLtW      TokenType = 54 // 小于
	TypeLogicGtW      TokenType = 55 // 大于
	TypeVarOneW       TokenType = 56 // 以
	TypeCondElseW     TokenType = 59 // 否则
	TypeWhileLoopW    TokenType = 60 // 每当
	TypeObjNewW       TokenType = 61 // 成为
	TypeObjDefineW    TokenType = 63 // 定义
	TypeObjThisW      TokenType = 65 // 其
	TypeLogicOrW      TokenType = 69 // 或
	TypeLogicAndW     TokenType = 70 // 且
	TypeObjDotW       TokenType = 71 // 之
	TypeObjConstructW TokenType = 73 // 是为
	TypeLogicEqualW   TokenType = 74 // 等于
	TypeStaticSelfW   TokenType = 75 // 此之
	TypeIteratorW     TokenType = 76 // 遍历
)

// KeywordTypeMap -
var KeywordTypeMap = map[TokenType][]rune{
	TypeDeclareW:      {GlyphLING},
	TypeLogicYesW:     {GlyphWEI},
	TypeAssignConstW:  {GlyphHENG, GlyphWEI},
	TypeCondOtherW:    {GlyphZAI, GlyphRU},
	TypeCondW:         {GlyphRU, GlyphGUO},
	TypeFuncW:         {GlyphRU, GlyphHE},
	TypeGetterW:       {GlyphHE, GlyphWEI},
	TypeParamAssignW:  {GlyphYI, GlyphZHIy},
	TypeReturnW:       {GlyphFAN, GlyphHUI},
	TypeLogicNotW:     {GlyphBU, GlyphWEI},
	TypeLogicNotEqW:   {GlyphBU, GlyphDENG, GlyphYU},
	TypeLogicLteW:     {GlyphBU, GlyphDA, GlyphYU},
	TypeLogicGteW:     {GlyphBU, GlyphXIAO, GlyphYU},
	TypeLogicLtW:      {GlyphXIAO, GlyphYU},
	TypeLogicGtW:      {GlyphDA, GlyphYU},
	TypeVarOneW:       {GlyphYIi},
	TypeCondElseW:     {GlyphFOU, GlyphZE},
	TypeWhileLoopW:    {GlyphMEI, GlyphDANG},
	TypeObjNewW:       {GlyphCHENG, GlyphWEI},
	TypeObjDefineW:    {GlyphDING, GlyphYIy},
	TypeObjThisW:      {GlyphQI},
	TypeLogicOrW:      {GlyphHUO},
	TypeLogicAndW:     {GlyphQIE},
	TypeObjDotW:       {GlyphZHI},
	TypeObjConstructW: {GlyphSHI, GlyphWEI},
	TypeLogicEqualW:   {GlyphDENG, GlyphYU},
	TypeStaticSelfW:   {GlyphCI, GlyphZHI},
	TypeIteratorW:     {GlyphBIAN, GlyphLI},
}

// parseKeyword -
// @return bool matchKeyword
// @return *Token token
//
// when matchKeyword = true, a keyword token will be generated
// matchKeyword = false, regard it as normal identifer
// and return directly.
func (l *Lexer) parseKeyword(ch rune, moveForward bool) (bool, *Token) {
	var tk *Token
	var wordLen = 1

	rg := newTokenRange(l)
	// manual matching one or consecutive keywords
	switch ch {
	case GlyphBU:
		if l.peek() == GlyphWEI {
			wordLen = 2
			tk = NewKeywordToken(TypeLogicNotW)
		} else if l.peek() == GlyphXIAO && l.peek2() == GlyphYU {
			wordLen = 3
			tk = NewKeywordToken(TypeLogicGteW)
		} else if l.peek() == GlyphDA && l.peek2() == GlyphYU {
			wordLen = 3
			tk = NewKeywordToken(TypeLogicLteW)
		} else if l.peek() == GlyphDENG && l.peek2() == GlyphYU {
			wordLen = 3
			tk = NewKeywordToken(TypeLogicNotEqW)
		} else {
			return false, nil
		}
	case GlyphQIE:
		tk = NewKeywordToken(TypeLogicAndW)
	case GlyphWEI:
		tk = NewKeywordToken(TypeLogicYesW)
	case GlyphZHI:
		tk = NewKeywordToken(TypeObjDotW)
	case GlyphLING:
		tk = NewKeywordToken(TypeDeclareW)
	case GlyphYIi:
		tk = NewKeywordToken(TypeVarOneW)
	case GlyphHE:
		if l.peek() == GlyphWEI {
			wordLen = 2
			tk = NewKeywordToken(TypeGetterW)
		} else {
			return false, nil
		}
	case GlyphQI:
		tk = NewKeywordToken(TypeObjThisW)
	case GlyphZAI:
		if l.peek() == GlyphRU {
			wordLen = 2
			tk = NewKeywordToken(TypeCondOtherW)
		} else {
			return false, nil
		}
	case GlyphFOU:
		if l.peek() == GlyphZE {
			wordLen = 2
			tk = NewKeywordToken(TypeCondElseW)
		} else {
			return false, nil
		}
	case GlyphDA:
		if l.peek() == GlyphYU {
			wordLen = 2
			tk = NewKeywordToken(TypeLogicGtW)
		} else {
			return false, nil
		}
	case GlyphRU:
		if l.peek() == GlyphHE {
			wordLen = 2
			tk = NewKeywordToken(TypeFuncW)
		} else if l.peek() == GlyphGUO {
			wordLen = 2
			tk = NewKeywordToken(TypeCondW)
		} else {
			return false, nil
		}
	case GlyphDING:
		if l.peek() == GlyphYIy {
			wordLen = 2
			tk = NewKeywordToken(TypeObjDefineW)
		} else {
			return false, nil
		}
	case GlyphXIAO:
		if l.peek() == GlyphYU {
			wordLen = 2
			tk = NewKeywordToken(TypeLogicLtW)
		} else {
			return false, nil
		}
	case GlyphYI:
		if l.peek() == GlyphZHIy {
			wordLen = 2
			tk = NewKeywordToken(TypeParamAssignW)
		} else {
			return false, nil
		}
	case GlyphHENG:
		if l.peek() == GlyphWEI {
			wordLen = 2
			tk = NewKeywordToken(TypeAssignConstW)
		} else {
			return false, nil
		}
	case GlyphCHENG:
		if l.peek() == GlyphWEI {
			wordLen = 2
			tk = NewKeywordToken(TypeObjNewW)
		} else {
			return false, nil
		}
	case GlyphHUO:
		tk = NewKeywordToken(TypeLogicOrW)
	case GlyphSHI:
		if l.peek() == GlyphWEI {
			wordLen = 2
			tk = NewKeywordToken(TypeObjConstructW)
		} else {
			return false, nil
		}
	case GlyphCI:
		if l.peek() == GlyphZHI {
			wordLen = 2
			tk = NewKeywordToken(TypeStaticSelfW)
		} else {
			return false, nil
		}
	case GlyphMEI:
		if l.peek() == GlyphDANG {
			wordLen = 2
			tk = NewKeywordToken(TypeWhileLoopW)
		} else {
			return false, nil
		}
	case GlyphDENG:
		if l.peek() == GlyphYU {
			wordLen = 2
			tk = NewKeywordToken(TypeLogicEqualW)
		} else {
			return false, nil
		}
	case GlyphFAN:
		if l.peek() == GlyphHUI {
			wordLen = 2
			tk = NewKeywordToken(TypeReturnW)
		} else {
			return false, nil
		}
	case GlyphBIAN:
		if l.peek() == GlyphLI {
			wordLen = 2
			tk = NewKeywordToken(TypeIteratorW)
		} else {
			return false, nil
		}
	}

	if tk != nil {
		if moveForward {
			switch wordLen {
			case 1:
				l.pushBuffer(ch)
			case 2:
				l.pushBuffer(ch, l.next())
			case 3:
				l.pushBuffer(ch, l.next(), l.next())
			}
		}
		rg.EndLine = rg.StartLine
		rg.EndIdx = rg.StartIdx + wordLen
		tk.Range = rg
		return true, tk
	}
	return false, nil
}
