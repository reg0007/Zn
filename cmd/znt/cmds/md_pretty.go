package cmds

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/reg0007/Zn/lex"
	"github.com/spf13/cobra"
)

const (
	inline       = 1
	block        = 2
	inlineParsed = 3
	blockParsed  = 4

	// GitHub style (light) color scheme
	csKeyword  = "#d73a49"
	csToken    = "#6f42c1"
	csNumber   = "#005cc5"
	csString   = "#032f62"
	csVariable = "#e36209"
	csComment  = "#6a737d"

	fontFamily = "Sarasa Mono SC, Microsoft YaHei, monospace"
)

// MdPrettyCmd -
var MdPrettyCmd = &cobra.Command{
	Use:   "md-pretty [file]",
	Short: "语法高亮 Markdown 文件中的 Zn 语言代码",
	Long: "将 Markdown 文件中带有 ```zn 开头的块状语句以及 `zn: <data>` 标识进行语法高亮，\n" +
		"使得文档变得更加容易阅读。\n" +
		"其原理非常简单：将代码放到Lexer中进行解析，解析成功后即对代码块进行正则替换，使之带上一个个HTML标签。\n\n" +
		"例：\n" +
		"```zn\n" +
		"令变量名为125\n" +
		"```\n\n" +
		"经过替换后变成：\n" +
		"<pre style='display:none' class='zn-ref-ADs23R'>\n" +
		"令变量名为125\n" +
		"</pre>\n" +
		"<pre class='zn-source-ADs23R'>\n" +
		"  <span style='color:red'>令</span>\n" +
		"  <span style='color:blue'>变量名</span>\n" +
		"  <span style='color:red'>为</span>\n" +
		"  <span style='color:green'>125</span>\n" +
		"</pre>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		fileData, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Println("[1] Error:", err.Error())
			return
		}

		newData := matchAndReplaceFile(string(fileData))
		if err := ioutil.WriteFile(filename, []byte(newData), 0644); err != nil {
			fmt.Println("[2] Error:", err.Error())
			return
		}

		fmt.Printf("替换 %s 成功\n", filename)
	},
}

func matchAndReplaceFile(data string) string {
	reInline := regexp.MustCompile("`zn: (.*?)`")
	reBlock := regexp.MustCompile("(?s)```zn(.*?)```")
	reRefCode := regexp.MustCompile(`<code.*?class='zn-ref-(\w+)'.*?>(.*?)</code>`)
	reRefPre := regexp.MustCompile(`(?s)<pre.*?class='zn-ref-(\w+)'.*?>(.*?)</pre>`)
	reSourceCode := regexp.MustCompile(`<code.*?class='zn-source-(\w+)'.*?>(.*?)</code>`)
	reSourcePre := regexp.MustCompile(`(?s)<pre.*?class='zn-source-(\w+)'.*?>(.*?)</pre>`)

	sourceMap := map[string]string{}

	// construct source map
	ref1 := reRefCode.FindAllStringSubmatch(data, -1)
	for _, match := range ref1 {
		sourceMap[match[1]] = match[2]
	}
	ref2 := reRefPre.FindAllStringSubmatch(data, -1)
	for _, match := range ref2 {
		sourceMap[match[1]] = match[2]
	}
	// replace parsed code from source map
	data = reSourceCode.ReplaceAllStringFunc(data, func(source string) string {
		m := reSourceCode.FindStringSubmatchIndex(source)
		ref := source[m[2]:m[3]]
		src, ok := sourceMap[ref]

		if !ok {
			return source
		}
		parsedCode := parseZnCode(inlineParsed, src)
		return strings.Join([]string{source[:m[4]], parsedCode, source[m[5]:]}, "")
	})

	data = reSourcePre.ReplaceAllStringFunc(data, func(source string) string {
		m := reSourcePre.FindStringSubmatchIndex(source)
		ref := source[m[2]:m[3]]
		src, ok := sourceMap[ref]

		if !ok {
			return source
		}
		parsedCode := parseZnCode(inlineParsed, src)
		return strings.Join([]string{source[:m[4]], parsedCode, source[m[5]:]}, "")
	})

	// parsed new ones
	data = reBlock.ReplaceAllStringFunc(data, func(source string) string {
		code := reBlock.FindStringSubmatch(source)[1]
		// trim CRLFs
		code = strings.Trim(code, "\r\n")
		return parseZnCode(block, code)
	})

	data = reInline.ReplaceAllStringFunc(data, func(source string) string {
		code := reInline.FindStringSubmatch(source)[1]
		return parseZnCode(inline, code)
	})

	return data
}

func parseZnCode(blockType int, code string) string {
	// remove zn:
	if blockType == inlineParsed || blockType == blockParsed {
		code = strings.Replace(code, "zn: ", "", 1)
	}

	ts := lex.NewTextStream(code)

	tokens := []*lex.Token{}
	l := lex.NewLexer(ts)

	// get all tokens
	for {
		tok, err := l.NextToken()
		if err != nil {
			fmt.Printf("解析错误： %s\n", err.Error())
			return code
		}

		if tok.Type == lex.TypeEOF {
			break
		}

		tokens = append(tokens, tok)
	}

	switch blockType {
	case inline:
		tag := generateTag(8)
		return fmt.Sprintf(
			"<code class='zn-ref-%s' style='display: none'>zn: %s</code><code class='zn-source-%s' style='font-family: %s'>%s</code>",
			tag,
			code,
			tag,
			fontFamily,
			composePrettyString(tokens, *l.LineStack),
		)
	case block:
		tag := generateTag(8)
		return fmt.Sprintf(
			"<pre class='zn-ref-%s' style='display: none'>zn: %s</pre>\n<pre class='zn-source-%s' style='font-family: %s'>%s</pre>",
			tag,
			code,
			tag,
			fontFamily,
			composePrettyString(tokens, *l.LineStack),
		)
	case inlineParsed, blockParsed:
		return composePrettyString(tokens, *l.LineStack)
	default:
		return code
	}
}

func composePrettyString(tokens []*lex.Token, lineStack lex.LineStack) string {
	lines := []string{}
	lineItem := []string{}

	indentType := lineStack.IndentType
	lastLine := 0
	lastIndex := 0

	for _, tk := range tokens {
		lineNum := tk.Range.StartLine
		if lineNum > lastLine {
			// commit old ones
			if lastLine != 0 {
				lines = append(lines, strings.Join(lineItem, ""))
				lineItem = []string{}
			}
			// add additional CRLFs
			for i := 0; i < lineNum-lastLine-1; i++ {
				lines = append(lines, "") // commit new line
			}

			// add indents
			if indentType == lex.IdetTab {
				lineItem = append(lineItem, strings.Repeat("\t", lineStack.GetLineIndent(lineNum)))
			} else if indentType == lex.IdetSpace {
				nbsps := strings.Repeat("&nbsp;", lineStack.GetLineIndent(lineNum)*4)
				lineItem = append(lineItem, fmt.Sprintf("<span>%s</span>", nbsps))
			}
		}
		// add additional spaces
		if lineNum == lastLine {
			colDiff := tk.Range.StartIdx - lastIndex
			if colDiff > 0 {
				nbsps := strings.Repeat("&nbsp;", colDiff)
				lineItem = append(lineItem, fmt.Sprintf("<span>%s</span>", nbsps))
			}
		}
		colorScheme := ""
		switch tk.Type {
		case lex.TypeString:
			colorScheme = csString
		case lex.TypeNumber:
			colorScheme = csNumber
		case lex.TypeMapData, lex.TypeFuncCall, lex.TypeFuncDeclare:
			colorScheme = csToken
		case lex.TypeComment:
			colorScheme = csComment
		}
		if tk.Type >= lex.TypeDeclareW && tk.Type <= lex.TypeStaticSelfW {
			colorScheme = csKeyword
		}

		if colorScheme == "" {
			lineItem = append(lineItem, string(tk.Literal))
		} else {
			lineItem = append(lineItem, fmt.Sprintf("<span style='color: %s'>%s</span>", colorScheme, string(tk.Literal)))
		}
		lastLine = tk.Range.EndLine
		lastIndex = tk.Range.EndIdx
	}

	if len(lineItem) > 0 {
		lines = append(lines, strings.Join(lineItem, ""))
	}
	return strings.Join(lines, "\n")
}

func generateTag(num int) string {
	dict := []rune("023456789ABCDEFGHIJKMNOPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
	final := []rune{}

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < num; i++ {
		n := rand.Intn(len(dict))
		final = append(final, dict[n])
	}
	return string(final)
}
