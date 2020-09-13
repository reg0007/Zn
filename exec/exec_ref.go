package exec

import (
	"github.com/reg0007/Zn/error"
	"github.com/reg0007/Zn/syntax"
)

type funcExecutor func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error)

type paramHandler func(ctx *Context, scope *FuncScope, params []ZnValue) *error.Error

// ClosureRef - aka. Closure Exection Reference
// This structure wraps the execution logic inside the closure
// statically
type ClosureRef struct {
	Name         string
	ParamHandler paramHandler // bind & validate params before actual execution
	Executor     funcExecutor // actual execution logic
}

// NewClosureRef -
func NewClosureRef(name string, paramTags []*syntax.ID, stmtBlock *syntax.BlockStmt) *ClosureRef {

	var executor = func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
		// iterate block round I - function hoisting
		for _, stmtI := range stmtBlock.Children {
			if v, ok := stmtI.(*syntax.FunctionDeclareStmt); ok {
				fn := NewZnFunction(v)
				if err := bindValue(ctx, scope, v.FuncName.GetLiteral(), fn, false); err != nil {
					return nil, err
				}
			}
		}
		// iterate block round II
		for _, stmtII := range stmtBlock.Children {
			if _, ok := stmtII.(*syntax.FunctionDeclareStmt); !ok {
				if err := evalStatement(ctx, scope, stmtII); err != nil {
					// if recv breaks
					if err.GetCode() == error.ReturnBreakSignal {
						if extra, ok := err.GetExtra().(ZnValue); ok {
							return extra, nil
						}
					}
					return nil, err
				}
			}
		}
		return scope.GetReturnValue(), nil
	}

	var paramHandler = func(ctx *Context, scope *FuncScope, params []ZnValue) *error.Error {
		// check param length
		if len(params) != len(paramTags) {
			return error.MismatchParamLengthError(len(paramTags), len(params))
		}

		// bind params (as variable) to function scope
		for idx, param := range params {
			paramTag := paramTags[idx].GetLiteral()
			if err := bindValue(ctx, scope, paramTag, param, false); err != nil {
				return err
			}
		}
		return nil
	}

	return &ClosureRef{
		Name:         name,
		ParamHandler: paramHandler,
		Executor:     executor,
	}
}

// NewNativeClosureRef - define native function
func NewNativeClosureRef(name string, executor funcExecutor) *ClosureRef {
	return &ClosureRef{
		Name:         name,
		ParamHandler: nil,
		Executor:     executor,
	}
}

// Exec - exec function
func (cr *ClosureRef) Exec(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
	// handle params
	if cr.ParamHandler != nil {
		if err := cr.ParamHandler(ctx, scope, params); err != nil {
			return nil, err
		}
	}
	// do execution
	return cr.Executor(ctx, scope, params)
}

// ClassRef -
type ClassRef struct {
	Name        string
	Constructor funcExecutor           // a function to initialize all properties
	GetterList  map[string]*ClosureRef // stores defined getters inside the class
	MethodList  map[string]*ClosureRef // stores defined methods inside the class
}

// NewClassRef -
func NewClassRef(name string, classNode *syntax.ClassDeclareStmt) *ClassRef {
	ref := &ClassRef{
		Name:        name,
		Constructor: nil,
		GetterList:  map[string]*ClosureRef{},
		MethodList:  map[string]*ClosureRef{},
	}

	// define default constrcutor
	var constructor = func(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
		obj := NewZnObject(ref)
		// init prop list
		for _, propPair := range classNode.PropertyList {
			propID := propPair.PropertyID.GetLiteral()
			expr, err := evalExpression(ctx, scope, propPair.InitValue)
			if err != nil {
				return nil, err
			}
			obj.PropList[propID] = expr
		}
		// constructor: set some properties' value
		if len(params) != len(classNode.ConstructorIDList) {
			return nil, error.MismatchParamLengthError(len(params), len(classNode.ConstructorIDList))
		}
		for idx, objParam := range params {
			propID := classNode.ConstructorIDList[idx].GetLiteral()
			obj.PropList[propID] = objParam
		}

		return obj, nil
	}
	// set constructor
	ref.Constructor = constructor

	// add getters
	for _, gNode := range classNode.GetterList {
		getterTag := gNode.GetterName.GetLiteral()
		ref.GetterList[getterTag] = NewClosureRef(getterTag, []*syntax.ID{}, gNode.ExecBlock)
	}

	// add methods
	for _, mNode := range classNode.MethodList {
		mTag := mNode.FuncName.GetLiteral()
		ref.MethodList[mTag] = NewClosureRef(mTag, mNode.ParamList, mNode.ExecBlock)
	}

	return ref
}

// Construct - yield new instance of this class
func (cr *ClassRef) Construct(ctx *Context, scope *FuncScope, params []ZnValue) (ZnValue, *error.Error) {
	return cr.Constructor(ctx, scope, params)
}
