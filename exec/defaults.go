package exec

import (
	"fmt"
	"strings"

	"github.com/reg0007/Zn/error"
)

var predefinedValues map[string]ZnValue

// （显示） 方法的执行逻辑
var displayExecutor = func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
	// display format string
	var items = []string{}

	for _, param := range params {
		if v, ok := param.(*ZnString); ok {
			items = append(items, v.Value)
		} else {
			items = append(items, param.String())
		}
	}
	fmt.Printf("%s\n", strings.Join(items, " "))
	return NewZnNull(), nil
}

// （递增）方法的执行逻辑
var addValueExecutor = func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
	var decimals = []*ZnDecimal{}
	if len(params) == 0 {
		return nil, error.LeastParamsError(1)
	}
	// validate types
	for _, param := range params {
		vparam, ok := param.(*ZnDecimal)
		if !ok {
			return nil, error.InvalidParamType("decimal")
		}
		decimals = append(decimals, vparam)
	}

	if len(decimals) == 1 {
		return decimals[0], nil
	}

	sum := ctx.arith.Add(decimals[0], decimals[1:]...)
	return sum, nil
}

// （递减）方法的执行逻辑
var subValueExecutor = func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
	var decimals = []*ZnDecimal{}
	if len(params) == 0 {
		return nil, error.LeastParamsError(1)
	}
	// validate types
	for _, param := range params {
		vparam, ok := param.(*ZnDecimal)
		if !ok {
			return nil, error.InvalidParamType("decimal")
		}
		decimals = append(decimals, vparam)
	}

	if len(decimals) == 1 {
		return decimals[0], nil
	}

	sum := ctx.arith.Sub(decimals[0], decimals[1:]...)
	return sum, nil
}

var mulValueExecutor = func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
	var decimals = []*ZnDecimal{}
	if len(params) == 0 {
		return nil, error.LeastParamsError(1)
	}
	// validate types
	for _, param := range params {
		vparam, ok := param.(*ZnDecimal)
		if !ok {
			return nil, error.InvalidParamType("decimal")
		}
		decimals = append(decimals, vparam)
	}

	if len(decimals) == 1 {
		return decimals[0], nil
	}

	sum := ctx.arith.Mul(decimals[0], decimals[1:]...)
	return sum, nil
}

var divValueExecutor = func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
	var decimals = []*ZnDecimal{}
	if len(params) == 0 {
		return nil, error.LeastParamsError(1)
	}
	// validate types
	for _, param := range params {
		vparam, ok := param.(*ZnDecimal)
		if !ok {
			return nil, error.InvalidParamType("decimal")
		}
		decimals = append(decimals, vparam)
	}
	if len(decimals) == 1 {
		return decimals[0], nil
	}

	return ctx.arith.Div(decimals[0], decimals[1:]...)
}

var probeExecutor = func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
	if len(params) != 2 {
		return nil, error.ExactParamsError(2)
	}

	vtag, ok := params[0].(*ZnString)
	if !ok {
		return nil, error.InvalidParamType("string")
	}
	// add probe data to log
	ctx._probe.AddLog(vtag.Value, params[1])
	return params[1], nil
}

var defaultDecimalClassRef = &ClassRef{
	Name: "数值",
	Constructor: func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
		return NewZnNull(), nil
	},
	// decimal to string
	GetterList: map[string]*ClosureRef{
		"文本": {
			Name: "文本",
			Executor: func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
				this, ok := scope.GetTargetThis().(*ZnDecimal)
				if !ok {
					return nil, error.NewErrorSLOT("invalid object type")
				}
				return NewZnString(this.String()), nil
			},
		},
	},
}

var defaultArrayClassRef = &ClassRef{
	Name: "数组",
	Constructor: func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
		return NewZnNull(), nil
	},
	GetterList: map[string]*ClosureRef{
		"和": {
			// 【1，2，3】之和
			Name: "和",
			Executor: func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
				this, ok := scope.GetTargetThis().(*ZnArray)
				if !ok {
					return nil, error.NewErrorSLOT("invalid object type")
				}
				return addValueExecutor(ctx, scope, this.Value)
			},
		},
		"差": {
			Name: "差",
			Executor: func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
				this, ok := scope.GetTargetThis().(*ZnArray)
				if !ok {
					return nil, error.NewErrorSLOT("invalid object type")
				}
				return subValueExecutor(ctx, scope, this.Value)
			},
		},
		"积": {
			Name: "积",
			Executor: func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
				this, ok := scope.GetTargetThis().(*ZnArray)
				if !ok {
					return nil, error.NewErrorSLOT("invalid object type")
				}
				return mulValueExecutor(ctx, scope, this.Value)
			},
		},
		"商": {
			Name: "商",
			Executor: func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
				this, ok := scope.GetTargetThis().(*ZnArray)
				if !ok {
					return nil, error.NewErrorSLOT("invalid object type")
				}
				return divValueExecutor(ctx, scope, this.Value)
			},
		},
		// get first item of array
		"首": {
			Name: "首",
			Executor: func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
				this, ok := scope.GetTargetThis().(*ZnArray)
				if !ok {
					return nil, error.NewErrorSLOT("invalid object type")
				}
				if len(this.Value) == 0 {
					return NewZnNull(), nil
				}
				return this.Value[0], nil
			},
		},
		// get last item of array
		"尾": {
			Name: "尾",
			Executor: func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
				this, ok := scope.GetTargetThis().(*ZnArray)
				if !ok {
					return nil, error.NewErrorSLOT("invalid object type")
				}
				if len(this.Value) == 0 {
					return NewZnNull(), nil
				}
				return this.Value[len(this.Value)-1], nil
			},
		},
		"数目": {
			Name: "数目",
			Executor: func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
				this, ok := scope.GetTargetThis().(*ZnArray)
				if !ok {
					return nil, error.NewErrorSLOT("invalid object type")
				}
				return NewZnDecimalFromInt(len(this.Value), 0), nil
			},
		},
	},
}

// init function
func init() {
	//// predefined values - those variables (symbols) are defined before
	//// any execution procedure.
	//// NOTICE: those variables are all constants!
	predefinedValues = map[string]ZnValue{
		"真":       NewZnBool(true),
		"假":       NewZnBool(false),
		"空":       NewZnNull(),
		"显示":      NewZnNativeFunction("显示", displayExecutor),
		"X+Y":     NewZnNativeFunction("X+Y", addValueExecutor),
		"求和":      NewZnNativeFunction("X+Y", addValueExecutor),
		"X-Y":     NewZnNativeFunction("X-Y", subValueExecutor),
		"求差":      NewZnNativeFunction("X-Y", subValueExecutor),
		"X*Y":     NewZnNativeFunction("X*Y", mulValueExecutor),
		"求积":      NewZnNativeFunction("X*Y", mulValueExecutor),
		"X/Y":     NewZnNativeFunction("X/Y", divValueExecutor),
		"求商":      NewZnNativeFunction("X/Y", divValueExecutor),
		"__probe": NewZnNativeFunction("__probe", probeExecutor),
	}
}
