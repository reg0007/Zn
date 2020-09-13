package cmds

import (
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type kwItem struct {
	name    string
	literal []rune
}

var keywordFileTemplate = `package lex

//// keyword character (ideoglyphs) definition
// keywords are all ideoglyphs that its length varies from its definitions.
// so here we define all possible chars that may be an element of one keyword.
const (
	%s
)

// KeywordLeads - all glyphs that would be possible of the first character of one keyword.
var KeywordLeads = []rune{
	%s
}

// Keyword token types
const (
	%s
)

// KeywordTypeMap -
var KeywordTypeMap = map[TokenType][]rune{
	%s
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
		%s
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
}`

var optKWOutputFile string

// GenKeywordCmd - generate keyword token definition from config file
var GenKeywordCmd = &cobra.Command{
	Use:   "gen-keyword [file]",
	Short: "根据关键词配置以生成对应（keyword）定义及解析逻辑 - lex/keyword.go",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dat, err := ioutil.ReadFile(args[0])
		if err != nil {
			panic(err)
		}
		charMap, keywordMap := splitInputFile(dat)

		leadsMap := exportKeywordLeadMap(charMap, keywordMap)
		containMap := exportCharContainMap(charMap, keywordMap)

		charsList := getCharsList(charMap)

		genCode := fmt.Sprintf(keywordFileTemplate,
			genCharConsts(charsList, charMap, containMap),
			genKeywordLeadsConsts(leadsMap, charMap),
			genKeywordTypeConsts(keywordMap),
			genKeywordTypeMap(keywordMap, charMap),
			genKeywordParsingLogic(leadsMap, keywordMap, charMap),
		)
		prettifyAndWriteCode(genCode, optKWOutputFile)
	},
}

func prettifyAndWriteCode(raw string, fileName string) {
	// parse file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "keyword.go", raw, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}

	cfg := printer.Config{
		Mode:     printer.UseSpaces | printer.TabIndent,
		Tabwidth: 4,
	}
	cfg.Fprint(file, fset, node)
}

// parse and insert to charMap & keywordMap
func splitInputFile(dat []byte) (map[rune]string, map[int]kwItem) {
	phase := 1
	lines := strings.Split(string(dat), "\n")
	charMap := map[rune]string{}
	keywordMap := map[int]kwItem{}
	for _, line := range lines {
		if strings.HasPrefix(line, "========") {
			phase = 2
			continue
		}
		// regard it as comment, ignore it
		if len(line) == 0 || strings.HasPrefix(line, "#") || strings.HasPrefix(line, " ") {
			continue
		}

		items := strings.Fields(line)
		// if not, ignore this line and parse next line
		if phase == 1 {
			if len(items) == 2 {
				r := []rune(items[1])
				charMap[r[0]] = fmt.Sprintf("Glyph%s", items[0])
			}
		} else if phase == 2 {
			if len(items) == 3 {
				t, e := strconv.Atoi(items[1])
				if e != nil {
					panic(e)
				}

				keywordMap[t] = kwItem{
					name:    fmt.Sprintf("Type%s", items[0]),
					literal: []rune(items[2]),
				}
			}
		}
	}
	return charMap, keywordMap
}

// exportKeywordLeadMap -
// get all possible keywordTypes that are leads with one specific character
func exportKeywordLeadMap(charMap map[rune]string, kwMap map[int]kwItem) map[rune][]int {
	keywordLeadMap := map[rune][]int{}
	// only include keywords that char is contained
	for kwType, kw := range kwMap {
		lead := kw.literal[0]
		// if lead rune exists in charMap
		if _, ok := charMap[lead]; ok {
			if _, ok2 := keywordLeadMap[lead]; !ok2 {
				keywordLeadMap[lead] = []int{}
			}
			keywordLeadMap[lead] = append(keywordLeadMap[lead], kwType)
		}
	}
	return keywordLeadMap
}

// exportCharContainMap -
// get get all keywords (strings) that contains one character
func exportCharContainMap(charMap map[rune]string, kwMap map[int]kwItem) map[rune][]string {
	containMap := map[rune][]string{}

	// only include keywords that char is contained
	for _, kw := range kwMap {
		for _, ch := range kw.literal {
			if _, ok := charMap[ch]; ok {
				if _, ok2 := containMap[ch]; !ok2 {
					containMap[ch] = []string{}
				}
				// add one item
				containMap[ch] = append(containMap[ch], string(kw.literal))
			}
		}
	}

	// sort containsMap
	for c := range containMap {
		sort.Slice(containMap[c], func(i, j int) bool {
			return strings.Compare(containMap[c][i], containMap[c][j]) > 0
		})
	}
	return containMap
}

///// generators
func genCharConsts(chars []rune, charMap map[rune]string, containsMap map[rune][]string) string {
	// generate code items
	codeList := []string{}

	for _, ch := range chars {
		commentLine := fmt.Sprintf("// %s - %s - %s", charMap[ch], string(ch), strings.Join(containsMap[ch], "，"))
		varLine := fmt.Sprintf("%s rune = 0x%X", charMap[ch], ch)

		codeList = append(codeList, commentLine, varLine)
	}
	return strings.Join(codeList, "\n\t")
}

func genKeywordLeadsConsts(leadsMap map[rune][]int, charMap map[rune]string) string {
	//// dump leads
	leads := make([]rune, len(leadsMap))
	i := 0
	for k := range leadsMap {
		leads[i] = k
		i++
	}
	sort.Slice(leads, func(i, j int) bool {
		return leads[i] < leads[j]
	})

	codeList := []string{}

	tmpStrs := []string{}
	for idx, ch := range leads {
		if idx > 0 && idx%3 == 0 {
			tmpStrs = append(tmpStrs, "") // add an empty string to compose tail comma
			codeList = append(codeList, strings.Join(tmpStrs, ","))
			tmpStrs = []string{}
		}
		tmpStrs = append(tmpStrs, charMap[ch])
	}
	// compose final ones
	if len(tmpStrs) > 0 {
		tmpStrs = append(tmpStrs, "") // add an empty string to compose tail comma
		codeList = append(codeList, strings.Join(tmpStrs, ","))
	}

	return strings.Join(codeList, "\n\t")
}

func genKeywordTypeConsts(keywordMap map[int]kwItem) string {
	types := make([]int, len(keywordMap))
	i := 0
	for k := range keywordMap {
		types[i] = k
		i++
	}
	sort.Slice(types, func(i, j int) bool {
		return types[i] < types[j]
	})

	codeList := []string{}

	for _, t := range types {
		codeList = append(codeList, fmt.Sprintf(
			"%s TokenType = %d // %s", keywordMap[t].name, t, string(keywordMap[t].literal),
		))
	}
	return strings.Join(codeList, "\n\t")
}

func genKeywordTypeMap(keywordMap map[int]kwItem, charMap map[rune]string) string {
	types := make([]int, len(keywordMap))
	i := 0
	for k := range keywordMap {
		types[i] = k
		i++
	}
	sort.Slice(types, func(i, j int) bool {
		return types[i] < types[j]
	})

	codeList := []string{}

	for _, t := range types {
		chars := []string{}
		for _, k := range keywordMap[t].literal {
			chars = append(chars, charMap[k])
		}
		codeList = append(codeList, fmt.Sprintf(
			"%s: {%s},", keywordMap[t].name, strings.Join(chars, ", "),
		))
	}
	return strings.Join(codeList, "\n\t")
}

// mostly support 3 characters
func genKeywordParsingLogic(leadsMap map[rune][]int, keywordMap map[int]kwItem, charMap map[rune]string) string {
	//// dump leads
	leads := make([]rune, len(leadsMap))
	i := 0
	for k := range leadsMap {
		leads[i] = k
		i++
	}
	sort.Slice(leads, func(i, j int) bool {
		return leads[i] < leads[j]
	})

	codeList := []string{}
	//// analyse logic by each lead char
	for _, leadCh := range leads {
		// append case statement
		// example:
		//   case GlyphWEI:
		codeList = append(codeList, fmt.Sprintf("case %s:", charMap[leadCh]))
		// nestMap stores all useful info for all types leads with this character
		nestMap := struct {
			oneChar    []string   // [TypeI]
			twoChars   [][]string // [ [CharII, TypeII], ... ]
			threeChars [][]string // [ [CharII, CharIII, TypeIII], ... ]
		}{}
		/** if-block general template:

		if l.peek() == GlyphX {
			wordLen = 2
			tk = NewKeywordToken(TypeX)
		} else if l.peek() === GlyphA && l.peek2() == GlyphB {
			wordLen = 3
			tk = NewKeywordToken(TypeX)
		} else {
			return false, nil
			// or  tk = NewKeywordToken(TypeK)
		}
		*/
		for _, t := range leadsMap[leadCh] {
			kw := keywordMap[t]
			switch len(kw.literal) {
			case 1:
				nestMap.oneChar = []string{kw.name}
			case 2:
				nestMap.twoChars = append(nestMap.twoChars, []string{
					charMap[kw.literal[1]],
					kw.name,
				})
			case 3:
				nestMap.threeChars = append(nestMap.threeChars, []string{
					charMap[kw.literal[1]],
					charMap[kw.literal[2]],
					kw.name,
				})
			}
		}

		// generate blocks
		// CASE I: only one char valid
		if len(nestMap.threeChars) == 0 && len(nestMap.twoChars) == 0 && len(nestMap.oneChar) == 1 {
			codeList = append(codeList, fmt.Sprintf("tk = NewKeywordToken(%s)", nestMap.oneChar[0]))
		} else {
			firstIfBlock := true
			// append two chars keyword
			for _, tch := range nestMap.twoChars {
				startIf := "if"
				if firstIfBlock {
					firstIfBlock = false
				} else {
					startIf = "} else if"
				}
				codeList = append(codeList,
					fmt.Sprintf("%s l.peek() == %s {", startIf, tch[0]),
					"wordLen = 2",
					fmt.Sprintf("tk = NewKeywordToken(%s)", tch[1]),
				)
			}
			// append three chars keyword
			for _, rch := range nestMap.threeChars {
				startIff := "if"
				if firstIfBlock {
					firstIfBlock = false
				} else {
					startIff = "} else if"
				}
				codeList = append(codeList,
					fmt.Sprintf("%s l.peek() == %s && l.peek2() == %s {", startIff, rch[0], rch[1]),
					"wordLen = 3",
					fmt.Sprintf("tk = NewKeywordToken(%s)", rch[2]),
				)
			}
			// generate else block
			if len(nestMap.oneChar) == 1 {
				codeList = append(codeList,
					"} else {",
					fmt.Sprintf("tk = NewKeywordToken(%s)", nestMap.oneChar[0]),
					"}",
				)
			} else {
				codeList = append(codeList,
					"} else {",
					"return false, nil",
					"}",
				)
			}
		}
	}

	return strings.Join(codeList, "\n")
}

func getCharsList(charMap map[rune]string) []rune {
	chars := make([]rune, len(charMap))
	i := 0
	for k := range charMap {
		chars[i] = k
		i++
	}
	sort.Slice(chars, func(i, j int) bool {
		return chars[i] < chars[j]
	})

	return chars
}

func init() {
	GenKeywordCmd.Flags().StringVarP(&optKWOutputFile, "outFile", "o", "lex/keyword.go", "导出文件位置")
}
