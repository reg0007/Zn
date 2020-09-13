package exec

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/reg0007/Zn/lex"
)

func TestExecuteCode_OK(t *testing.T) {
	cases := []struct {
		name        string
		text        string
		resultValue ZnValue
	}{
		{
			"normal oneline expression",
			"令A为10；A为10241024",
			newDecimal("10241024"),
		},
		{
			"function call",
			"如何测试？\n    （X+Y：2，3）\n\n（测试）",
			newDecimal("5"),
		},
		{
			"with return",
			`如何测试？
	已知阈值
	如果阈值大于10：
		返回「大于」
	返回「小于」
	「等于」  注：这是一个干扰项
	
（测试：6）`,
			NewZnString("小于"),
		},
	}
	ctx := NewContext()

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			in := lex.NewTextStream(tt.text)
			scope := NewRootScope()
			res := ctx.ExecuteCode(in, scope)
			if res.HasError == true {
				t.Errorf("expect no error, has got error: %v", res.Error)
				return
			}

			if !reflect.DeepEqual(tt.resultValue, res.Value) {
				t.Errorf("expect value: %v, got: %v", tt.resultValue, res.Value)
				return
			}
		})
	}
}

// display full error info
func TestExecuteCode_FAIL(t *testing.T) {
	text := `令变量名-甲为10
令变量名-乙为20
（X+Y：变量名-未定，变量名-甲）`

	in := lex.NewTextStream(text)
	ctx := NewContext()
	scope := NewRootScope()
	result := ctx.ExecuteCode(in, scope)

	if result.HasError == false {
		t.Errorf("should got error, return no error")
		return
	}

	displayText := `在「$repl」中，位于第 3 行发现错误：
    （X+Y：变量名-未定，变量名-甲）
    
‹2501› 标识错误：标识「变量名-未定」未有定义`

	if result.Error.Display() != displayText {
		t.Errorf("should return \n%s\n, got \n%s\n", displayText, result.Error.Display())
		fmt.Println([]rune(result.Error.Display()))
		fmt.Println([]rune(displayText))
		return
	}
}

// create decimal (and ignore errors)
func newDecimal(value string) *ZnDecimal {
	dat, _ := NewZnDecimal(value)
	return dat
}
