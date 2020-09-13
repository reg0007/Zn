package exec

import (
	"github.com/reg0007/Zn/debug"
	"github.com/reg0007/Zn/error"
	"github.com/reg0007/Zn/lex"
	"github.com/reg0007/Zn/syntax"
)

// Context - GLOBAL execution context, usually create only once in one program.
type Context struct {
	globals map[string]ZnValue
	arith   *Arith
	// a seperate map to store inner debug data
	// usage: call （__probe：「tagName」，variable）
	// it will record all logs (including variable value, curernt scope, etc.)
	// the value is deep-copied so don't worry - the value logged won't be changed
	_probe *debug.Probe
}

const defaultPrecision = 8

// Result - context execution result structure
// NOTICE: when HasError = true, Value = nil, while execution yields error
//         when HasError = false, Error = nil, Value = <result Value>
//
// Currently only one value is supported as return argument.
type Result struct {
	HasError bool
	Value    ZnValue
	Error    *error.Error
}

// NewContext - create new Zn Context for furthur execution
func NewContext() *Context {
	return &Context{
		globals: predefinedValues,
		arith:   NewArith(defaultPrecision),
		_probe:  debug.NewProbe(),
	}
}

// ExecuteCode - execute program from input Zn code (whether from file or REPL)
func (ctx *Context) ExecuteCode(in *lex.InputStream, scope *RootScope) Result {
	l := lex.NewLexer(in)
	p := syntax.NewParser(l)
	// start
	block, err := p.Parse()
	if err != nil {
		return Result{true, nil, err}
	}
	// init scope
	scope.Init(l)

	// construct root (program) node
	program := syntax.NewProgramNode(block)

	// eval program
	if err := evalProgram(ctx, scope, program); err != nil {
		wrapError(ctx, scope, err)
		return Result{true, nil, err}
	}
	return Result{false, scope.GetLastValue(), nil}
}

// wrapError if lineInfo is missing (mostly for non-syntax errors)
// If lineInfo missing, then we will add current execution line and hide some part to
// display errors properly.
func wrapError(ctx *Context, scope *RootScope, err *error.Error) {
	cursor := err.GetCursor()

	if cursor.LineNum == 0 {
		newCursor := error.Cursor{
			File:    scope.file,
			LineNum: scope.currentLine,
			Text:    scope.lineStack.GetLineText(scope.currentLine, false),
		}
		err.SetCursor(newCursor)
	}
}
