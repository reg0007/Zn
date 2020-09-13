package cmds

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/flopp/go-findfont"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/reg0007/Zn/lex"
	"golang.org/x/image/font"

	"github.com/spf13/cobra"
)

const (
	marginTop    = 15.0
	marginBottom = 15.0
	marginLeft   = 12.0
	marginRight  = 12.0
	lineHeight   = 21
	fontSize     = 15

	finalWidth = 838 // GitHub <pre> bar width
)

var (
	optOutFile  string
	optFontFile string
)

type colorTextMap struct {
	text  string
	color string
}

// GenCodeImageCmd -
var GenCodeImageCmd = &cobra.Command{
	Use:   "gen-code-image [file]",
	Short: "对特定Zn语言文件生成相应的带语法高亮的图片",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inFile := args[0]
		// load font
		fontFace, err := loadFontFace(optFontFile, fontSize)
		if err != nil {
			fmt.Println(err)
			return
		}

		textMap, err := parseCodeFromFile(inFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		textHeight := measureCodeHeight(textMap)
		finalHeight := textHeight + marginTop + marginBottom + 4 // add some blanks at bottom for better apperance

		// new background
		dc := newBackground(finalWidth, finalHeight, "#f6f8fa")
		dc.SetFontFace(fontFace)

		drawColorText(dc, textMap)

		dc.SavePNG(optOutFile)
	},
}

func loadFontFace(fileName string, fontSize float64) (font.Face, error) {
	// load font
	fontPath, err := findfont.Find(fileName)
	if err != nil {
		return nil, err
	}

	// load the font with the freetype library
	fontData, err := ioutil.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}
	fontT, err := truetype.Parse(fontData)
	if err != nil {
		return nil, err
	}

	fontFace := truetype.NewFace(fontT, &truetype.Options{
		Size:    fontSize,
		Hinting: font.HintingFull,
	})
	return fontFace, nil
}

func newBackground(width int, height int, bgColor string) *gg.Context {
	dc := gg.NewContext(width, height)
	dc.SetHexColor(bgColor)
	dc.DrawRectangle(0, 0, float64(width), float64(height))
	dc.Fill()

	return dc
}

func measureCodeHeight(tMap [][]colorTextMap) int {
	return lineHeight * len(tMap)
}

func drawColorText(dc *gg.Context, tMap [][]colorTextMap) {
	spTop := marginTop   // start point Top (Y-axis)
	spLeft := marginLeft // start point Left (X-axis)

	for _, tMapItems := range tMap {
		spLeft = marginLeft
		for _, tMapItem := range tMapItems {
			w, _ := dc.MeasureString(tMapItem.text)
			dc.SetHexColor(tMapItem.color)
			dc.DrawString(tMapItem.text, spLeft, spTop+lineHeight*0.5)
			spLeft = spLeft + w
		}

		spTop = spTop + lineHeight
	}
	dc.Fill()
}

func parseCodeFromFile(file string) ([][]colorTextMap, error) {

	ts, err := lex.NewFileStream(file)
	if err != nil {
		return nil, fmt.Errorf("解析代码错误：具体报错为 %s", err.Display())
	}
	l := lex.NewLexer(ts)
	// generate [][]colorTextMap:
	//  []{    <-- all lines
	//     []colorTextMap{   <-- data in one line
	//
	//     }
	//  }
	tMap := [][]colorTextMap{}
	tMapItems := []colorTextMap{}
	lastLine := 0
	lastIndex := 0
	var lastTok *lex.Token
	for {
		tok, err := l.NextToken()
		indentType := l.LineStack.IndentType
		if err != nil {
			return nil, fmt.Errorf("解析代码错误：具体报错为 %s", err.Display())
		}

		if tok.Type == lex.TypeEOF {
			// before break, commit last line
			tMap = append(tMap, tMapItems)
			break
		}
		lineNum := tok.Range.StartLine
		if lineNum > lastLine {
			// commit old ones
			if lastLine != 0 {
				tMap = append(tMap, tMapItems)
				tMapItems = []colorTextMap{}
			}
			// add additional CRLFs
			for i := 0; i < lineNum-lastLine-1; i++ {
				tMap = append(tMap, []colorTextMap{})
			}
			// add indents
			indentStr := ""
			if indentType == lex.IdetTab {
				indentStr = strings.Repeat("\t", l.LineStack.GetLineIndent(lineNum))
			} else if indentType == lex.IdetSpace {
				indentStr = strings.Repeat(" ", l.LineStack.GetLineIndent(lineNum)*4)
			}
			if indentStr != "" {
				tMapItems = append(tMapItems, colorTextMap{
					text:  indentStr,
					color: matchColorScheme(lex.TypeSpace, lastTok),
				})
			}
		}
		// set spaces between tokens
		if lineNum == lastLine {
			colDiff := tok.Range.StartIdx - lastIndex
			if colDiff > 0 {
				nbsps := strings.Repeat(" ", colDiff)
				tMapItems = append(tMapItems, colorTextMap{
					text:  nbsps,
					color: matchColorScheme(lex.TypeSpace, lastTok),
				})
			}
		}
		// add literal (support multi-line token)
		literals := splitMultiLineString(tok.Literal)
		for idx, lt := range literals {
			tMapItems = append(tMapItems, colorTextMap{
				text:  lt,
				color: matchColorScheme(tok.Type, lastTok),
			})
			if idx < len(literals)-1 {
				tMap = append(tMap, tMapItems)
				tMapItems = []colorTextMap{}
			}
		}

		lastLine = tok.Range.EndLine
		lastIndex = tok.Range.EndIdx
		lastTok = tok
	}

	return tMap, nil
}

func matchColorScheme(tkType lex.TokenType, lastTok *lex.Token) string {
	const (
		// GitHub style (light) color scheme
		csKeyword  = "#d73a49"
		csToken    = "#6f42c1"
		csNumber   = "#005cc5"
		csString   = "#032f62"
		csVariable = "#e36209"
		csComment  = "#6a737d"
		csNormal   = "#24292e"
		csMember   = "#005cc5"
	)

	colorScheme := csNormal
	switch tkType {
	case lex.TypeString:
		colorScheme = csString
	case lex.TypeNumber:
		colorScheme = csNumber
	case lex.TypeMapData, lex.TypeFuncCall, lex.TypeFuncDeclare:
		colorScheme = csToken
	case lex.TypeComment:
		colorScheme = csComment
	case lex.TypeIdentifier:
		if lastTok != nil {
			// if lastToken is （ or 之 or 如何, that means the identifier is a member or a function name
			if lastTok.Type == lex.TypeObjDotW || lastTok.Type == lex.TypeFuncQuoteL || lastTok.Type == lex.TypeFuncW {
				colorScheme = csMember
			}
		}
	}
	if tkType >= lex.TypeDeclareW && tkType <= lex.TypeIteratorW {
		colorScheme = csKeyword
	}
	return colorScheme
}

// split string by '\n\r' | '\r\n' | '\n' | '\r'
func splitMultiLineString(literal []rune) []string {
	final := []string{}
	start := 0
	idx := 0
	lLen := len(literal) // literal length
	for {
		if idx >= lLen {
			break
		}
		r := literal[idx]
		if r == '\n' {
			final = append(final, string(literal[start:idx]))

			start = idx + 1
			if idx < lLen-1 && literal[idx+1] == '\r' {
				idx++
				start = idx + 2
			}
		} else if r == '\r' {
			final = append(final, string(literal[start:idx]))

			start = idx + 1
			if idx < lLen-1 && literal[idx+1] == '\n' {
				idx++
				start = idx + 2
			}
		}
		idx++
	}

	final = append(final, string(literal[start:]))
	return final
}

func init() {
	GenCodeImageCmd.Flags().StringVarP(&optOutFile, "outFile", "o", "out.png", "输出图片文件")
	GenCodeImageCmd.Flags().StringVarP(&optFontFile, "fontFile", "f", "sarasa-mono-sc-regular.ttf", "字体文件")
}
