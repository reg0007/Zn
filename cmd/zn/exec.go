package zn

import (
	"fmt"
	"io"
	"os"

	"github.com/peterh/liner"
	"github.com/reg0007/Zn/exec"
	"github.com/reg0007/Zn/lex"
)

const version = "rev04"

// EnterREPL - enter REPL to handle data
func EnterREPL() {
	linerR := liner.NewLiner()
	linerR.SetCtrlCAborts(true)
	ctx := exec.NewContext()
	scope := exec.NewRootScope()
	// REPL loop
	for {
		text, err := linerR.Prompt("Zn> ")
		if err != nil {
			if err == liner.ErrPromptAborted {
				os.Exit(0)
			} else if err.Error() == "EOF" {
				os.Exit(0)
			} else {
				fmt.Printf("未知错误：%s\n", err.Error())
				os.Exit(0)
			}
		}
		// append history
		linerR.AppendHistory(text)
		// add special command
		if text == ".print" {
			printSymbols(ctx)
			continue
		} else if text == ".exit" {
			break
		}

		// execute program
		in := lex.NewTextStream(text)
		result := ctx.ExecuteCode(in, scope)
		if !result.HasError {
			if result.Value != nil {
				prettyDisplayValue(result.Value, os.Stdout)
			}
		} else {
			fmt.Println(result.Error.Display())
		}
	}
}

// ExecProgram - exec program from file directly
func ExecProgram(file string) {
	ctx := exec.NewContext()
	scope := exec.NewRootScope()
	in, errF := lex.NewFileStream(file)
	if errF != nil {
		fmt.Println(errF.Display())
		return
	}

	result := ctx.ExecuteCode(in, scope)
	// when exec program, unlike REPL, it's not necessary to print last executed value
	if result.HasError {
		fmt.Println(result.Error.Display())
	}
}

// ShowVersion - show version
func ShowVersion() {
	fmt.Printf("Zn语言版本：%s\n", version)
}

//// display helpers
func prettyDisplayValue(val exec.ZnValue, w io.Writer) {
	var displayData = ""

	switch v := val.(type) {
	case *exec.ZnDecimal:
		// FG color: Cyan (lightblue)
		displayData = fmt.Sprintf("\x1b[38;5;147m%s\x1b[0m\n", v.String())
	case *exec.ZnString:
		// FG color: Green
		displayData = fmt.Sprintf("\x1b[38;5;184m%s\x1b[0m\n", v.String())
	case *exec.ZnBool:
		// FG color: White
		displayData = fmt.Sprintf("\x1b[38;5;231m%s\x1b[0m\n", v.String())
	case *exec.ZnNull, *exec.ZnFunction:
		displayData = fmt.Sprintf("‹\x1b[38;5;80m%s\x1b[0m›\n", v.String())
	default:
		displayData = fmt.Sprintf("%s\n", v.String())
	}

	w.Write([]byte(displayData))
}

// printSymbols -
func printSymbols(ctx *exec.Context) {
	/** TODO
	strs := []string{}
	for k, symArr := range ctx.GetSymbols() {
		if symArr != nil {
			for _, symItem := range symArr {
				symStr := "ε"
				if symItem.Value != nil {
					symStr = symItem.Value.String()
				}
				strs = append(strs, fmt.Sprintf("‹%s, %d› => %s", k, symItem.NestLevel, symStr))
			}
		}
	}

	data := strings.Join(strs, "\n")
	fmt.Println(data)
	*/
}
