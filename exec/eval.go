package exec

import (
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/reg0007/Zn/error"
	"github.com/reg0007/Zn/syntax"
)

// eval.go evaluates program from generated AST tree with specific scopes
// common signature of eval functions:
//
// evalXXXXStmt(ctx *Context, scope Scope, node Node) *error.Error
//
// or
//
// evalXXXXExpr(ctx *Context, scope Scope, node Node) (ZnValue, *error.Error)
//
// NOTICE:
// `evalXXXXStmt` will change the value of its corresponding scope; However, `evalXXXXExpr` will export
// a ZnValue object and mostly won't change scopes (but search a variable from scope is frequently used)

// duplicateValue -
func duplicateValue(in ZnValue) ZnValue {
	switch v := in.(type) {
	case *ZnBool:
		return &ZnBool{
			ZnObject: v.ZnObject,
			Value:    v.Value,
		}
	case *ZnString:
		return &ZnString{
			ZnObject: v.ZnObject,
			Value:    v.Value,
		}
	case *ZnDecimal:
		x := new(big.Int)
		return &ZnDecimal{
			ZnObject: v.ZnObject,
			co:       x.Set(v.co),
			exp:      v.exp,
		}
	case *ZnNull:
		return in // no need to copy since all "NULL" values are same
	case *ZnArray:
		newArr := []ZnValue{}
		for _, val := range v.Value {
			newArr = append(newArr, duplicateValue(val))
		}
		return &ZnArray{ZnObject: v.ZnObject, Value: newArr}
	case *ZnHashMap:
		newHashMap := map[string]ZnValue{}
		newKeyOrder := []string{}
		for key, val := range v.Value {
			newHashMap[key] = duplicateValue(val)
		}
		for _, keyItem := range v.KeyOrder {
			newKeyOrder = append(newKeyOrder, keyItem)
		}
		return &ZnHashMap{
			ZnObject: v.ZnObject,
			Value:    newHashMap,
			KeyOrder: newKeyOrder,
		}
	case *ZnFunction: // function itself is immutable, so return directly
		return in
	case *ZnObject:
		newPropList := map[string]ZnValue{}

		// copy prop
		for key, prop := range v.PropList {
			newPropList[key] = duplicateValue(prop)
		}
		return &ZnObject{
			ClassRef: v.ClassRef,
			PropList: newPropList,
		}
	}
	return in
}

type compareVerb uint8

// Define compareVerbs, for details of each verb, check the following comments
// on compareValues() function.
const (
	CmpEq compareVerb = 1
	CmpLt compareVerb = 2
	CmpGt compareVerb = 3
)

// compareValues - some ZnValues are comparable from specific types of right value
// otherwise it will throw error.
//
// There are three types of compare verbs (actions): Eq, Lt and Gt.
//
// Eq - compare if two values are "equal". Usually there are two rules:
// 1. types of left and right value are same. A number MUST BE equals to a number, that means
// (string) “2” won't be equals to (number) 2;
// 2. each items SHOULD BE identical, even for composited types (i.e. array, hashmap)
//
// Lt - for two decimals ONLY. If leftValue < rightValue.
//
// Gt - for two decimals ONLY. If leftValue > rightValue.
//
func compareValues(left ZnValue, right ZnValue, verb compareVerb) (bool, *error.Error) {
	switch vl := left.(type) {
	case *ZnNull:
		if _, ok := right.(*ZnNull); ok {
			return true, nil
		}
		return false, nil
	case *ZnDecimal:
		// compare right value - decimal only
		if vr, ok := right.(*ZnDecimal); ok {
			r1, r2 := rescalePair(vl, vr)
			cmpResult := false
			switch verb {
			case CmpEq:
				cmpResult = (r1.co.Cmp(r2.co) == 0)
			case CmpLt:
				cmpResult = (r1.co.Cmp(r2.co) < 0)
			case CmpGt:
				cmpResult = (r1.co.Cmp(r2.co) > 0)
			default:
				return false, error.UnExpectedCase("比较原语", strconv.Itoa(int(verb)))
			}
			return cmpResult, nil
		}
		// if vert == CmbEq and rightValue is not decimal type
		// then return `false` directly
		if verb == CmpEq {
			return false, nil
		}
		return false, error.InvalidCompareRType("decimal")
	case *ZnString:
		// Only CmpEq is valid for comparison
		if verb != CmpEq {
			return false, error.InvalidCompareLType("decimal", "string", "bool", "array", "hashmap")
		}
		// compare right value - string only
		if vr, ok := right.(*ZnString); ok {
			cmpResult := (strings.Compare(vl.Value, vr.Value) == 0)
			return cmpResult, nil
		}
		return false, nil
	case *ZnBool:
		if verb != CmpEq {
			return false, error.InvalidCompareLType("decimal", "string", "bool", "array", "hashmap")
		}
		// compare right value - bool only
		if vr, ok := right.(*ZnBool); ok {
			cmpResult := vl.Value == vr.Value
			return cmpResult, nil
		}
		return false, nil
	case *ZnArray:
		if verb != CmpEq {
			return false, error.InvalidCompareLType("decimal", "string", "bool", "array", "hashmap")
		}

		if vr, ok := right.(*ZnArray); ok {
			if len(vl.Value) != len(vr.Value) {
				return false, nil
			}
			// cmp each item
			for idx := range vl.Value {
				cmpVal, err := compareValues(vl.Value[idx], vr.Value[idx], CmpEq)
				if err != nil {
					return false, err
				}
				return cmpVal, nil
			}
			return true, nil
		}
		return false, nil
	case *ZnHashMap:
		if verb != CmpEq {
			return false, error.InvalidCompareLType("decimal", "string", "bool", "array", "hashmap")
		}

		if vr, ok := right.(*ZnHashMap); ok {
			if len(vl.Value) != len(vr.Value) {
				return false, nil
			}
			// cmp each item
			for idx := range vl.Value {
				// ensure the key exists on vr
				vrr, ok := vr.Value[idx]
				if !ok {
					return false, nil
				}
				cmpVal, err := compareValues(vl.Value[idx], vrr, CmpEq)
				if err != nil {
					return false, err
				}
				return cmpVal, nil
			}
			return true, nil
		}
		return false, nil
	}
	return false, error.InvalidCompareLType("decimal", "string", "bool", "array", "hashmap")
}

//// eval program
func evalProgram(ctx *Context, scope *RootScope, program *syntax.Program) *error.Error {
	return evalStmtBlock(ctx, scope, program.Content)
}

//// eval statements

// EvalStatement - eval statement
func evalStatement(ctx *Context, scope Scope, stmt syntax.Statement) *error.Error {
	// when evalStatement, last value should be set as ZnNull{}
	resetLastValue := true
	defer func() {
		if resetLastValue {
			scope.GetRoot().SetLastValue(NewZnNull())
		}
	}()
	scope.GetRoot().SetCurrentLine(stmt.GetCurrentLine())
	switch v := stmt.(type) {
	case *syntax.VarDeclareStmt:
		return evalVarDeclareStmt(ctx, scope, v)
	case *syntax.WhileLoopStmt:
		return evalWhileLoopStmt(ctx, scope, v)
	case *syntax.BranchStmt:
		return evalBranchStmt(ctx, scope, v)
	case *syntax.EmptyStmt:
		return nil
	case *syntax.FunctionDeclareStmt:
		fn := NewZnFunction(v)
		return bindValue(ctx, scope, v.FuncName.GetLiteral(), fn, false)
	case *syntax.ClassDeclareStmt:
		sp, ok := scope.(*RootScope)
		if !ok {
			return error.NewErrorSLOT("只能在RootScope使用类定义")
		}
		return bindClassRef(ctx, sp, v)
	case *syntax.IterateStmt:
		return evalIterateStmt(ctx, scope, v)
	case *syntax.FunctionReturnStmt:
		val, err := evalExpression(ctx, scope, v.ReturnExpr)
		if err != nil {
			return err
		}
		// send RETURN break
		return error.ReturnBreakError(val)
	case syntax.Expression:
		resetLastValue = false
		val, err := evalExpression(ctx, scope, v)
		if err != nil {
			return err
		}
		// set last value (of rootScope or funcScope)
		sp := scope
		for sp != nil {
			switch v := sp.(type) {
			case *RootScope:
				v.SetLastValue(val)
				return nil
			case *FuncScope:
				v.SetReturnValue(val)
				return nil
			}
			sp = sp.GetParent()
		}
		return nil
	default:
		return error.UnExpectedCase("语句类型", reflect.TypeOf(v).Name())
	}
}

// evalVarDeclareStmt - consists of three branches:
// 1. A，B 为 C
// 2. A，B 成为 X：P1，P2，...
// 3. A，B 恒为 C
func evalVarDeclareStmt(ctx *Context, scope Scope, node *syntax.VarDeclareStmt) *error.Error {
	for _, vpair := range node.AssignPair {
		switch vpair.Type {
		case syntax.VDTypeAssign, syntax.VDTypeAssignConst: // 为，恒为
			obj, err := evalExpression(ctx, scope, vpair.AssignExpr)
			if err != nil {
				return err
			}
			// if assign a constant variable or not
			isConst := false
			if vpair.Type == syntax.VDTypeAssignConst {
				isConst = true
			}

			for _, v := range vpair.Variables {
				vtag := v.GetLiteral()
				finalObj := duplicateValue(obj)

				if bindValue(ctx, scope, vtag, finalObj, isConst); err != nil {
					return err
				}
			}
		case syntax.VDTypeObjNew: // 成为
			if err := evalNewObjectPart(ctx, scope, vpair); err != nil {
				return err
			}
		}
	}
	return nil
}

// eval A,B 成为 C：P1，P2，P3，...
// ensure VDAssignPair.Type MUST BE syntax.VDTypeObjNew
func evalNewObjectPart(ctx *Context, scope Scope, node syntax.VDAssignPair) *error.Error {
	vtag := node.ObjClass.GetLiteral()
	// get class definition
	classRef, err := getClassRef(ctx, scope.GetRoot(), vtag)
	if err != nil {
		return err
	}

	cParams := []ZnValue{}
	for _, objParam := range node.ObjParams {
		expr, err := evalExpression(ctx, scope, objParam)
		if err != nil {
			return err
		}
		cParams = append(cParams, expr)
	}

	// assign new object to variables
	for _, v := range node.Variables {
		vtag := v.GetLiteral()
		// compose a new object instance
		fScope := NewFuncScope(scope, nil)
		finalObj, err := classRef.Construct(ctx, fScope, cParams)
		if err != nil {
			return err
		}

		if bindValue(ctx, scope, vtag, finalObj, false); err != nil {
			return err
		}
	}
	return nil
}

// evalWhileLoopStmt -
func evalWhileLoopStmt(ctx *Context, scope Scope, node *syntax.WhileLoopStmt) *error.Error {
	loopScope := NewWhileScope(scope)
	for {
		// #1. first execute expr
		trueExpr, err := evalExpression(ctx, loopScope, node.TrueExpr)
		if err != nil {
			return err
		}
		// #2. assert trueExpr to be ZnBool
		vTrueExpr, ok := trueExpr.(*ZnBool)
		if !ok {
			return error.InvalidExprType("bool")
		}
		// break the loop if expr yields not true
		if vTrueExpr.Value == false {
			return nil
		}
		// #3. stmt block
		if err := evalStmtBlock(ctx, loopScope, node.LoopBlock); err != nil {
			if err.GetCode() == error.ContinueBreakSignal {
				// continue next turn
				continue
			}
			if err.GetCode() == error.BreakBreakSignal {
				// break directly
				return nil
			}
			return err
		}
	}
}

// EvalStmtBlock -
func evalStmtBlock(ctx *Context, scope Scope, block *syntax.BlockStmt) *error.Error {
	enableHoist := false
	rootScope, ok := scope.(*RootScope)
	if ok {
		enableHoist = true
	}

	if enableHoist {
		// ROUND I: declare function stmt FIRST
		for _, stmtI := range block.Children {
			switch v := stmtI.(type) {
			case *syntax.FunctionDeclareStmt:
				fn := NewZnFunction(v)
				if err := bindValue(ctx, scope, v.FuncName.GetLiteral(), fn, false); err != nil {
					return err
				}
			case *syntax.ClassDeclareStmt:
				if err := bindClassRef(ctx, rootScope, v); err != nil {
					return err
				}
			}
		}
		// ROUND II: exec statement except functionDecl stmt
		for _, stmtII := range block.Children {
			switch stmtII.(type) {
			case *syntax.FunctionDeclareStmt, *syntax.ClassDeclareStmt:
				continue
			default:
				if err := evalStatement(ctx, scope, stmtII); err != nil {
					return err
				}
			}
		}
	} else {
		for _, stmt := range block.Children {
			if err := evalStatement(ctx, scope, stmt); err != nil {
				return err
			}
		}
	}
	return nil
}

func evalBranchStmt(ctx *Context, scope Scope, node *syntax.BranchStmt) *error.Error {
	// #1. condition header
	ifExpr, err := evalExpression(ctx, scope, node.IfTrueExpr)
	if err != nil {
		return err
	}
	vIfExpr, ok := ifExpr.(*ZnBool)
	if !ok {
		return error.InvalidExprType("bool")
	}
	// exec if-branch
	if vIfExpr.Value == true {
		return evalStmtBlock(ctx, scope, node.IfTrueBlock)
	}
	// exec else-if branches
	for idx, otherExpr := range node.OtherExprs {
		otherExprI, err := evalExpression(ctx, scope, otherExpr)
		if err != nil {
			return err
		}
		vOtherExprI, ok := otherExprI.(*ZnBool)
		if !ok {
			return error.InvalidExprType("bool")
		}
		// exec else-if branch
		if vOtherExprI.Value == true {
			return evalStmtBlock(ctx, scope, node.OtherBlocks[idx])
		}
	}
	// exec else branch if possible
	if node.HasElse == true {
		return evalStmtBlock(ctx, scope, node.IfFalseBlock)
	}
	return nil
}

func evalIterateStmt(ctx *Context, scope Scope, node *syntax.IterateStmt) *error.Error {
	// pre-defined key, value variable name
	var keySlot, valueSlot string
	var nameLen = len(node.IndexNames)

	iterScope := NewIterateScope(scope)
	// 以A，B遍历C： D
	// execute expr: C
	targetExpr, err := evalExpression(ctx, scope, node.IterateExpr)
	if err != nil {
		return err
	}

	// execIterationBlock, including set "currentKey" and "currentValue" to scope,
	// and preDefined indication variables
	execIterationBlockFn := func(key ZnValue, val ZnValue) *error.Error {
		// set values of 此之值 and 此之
		iterScope.setCurrentKV(key, val)

		// set pre-defined value
		if nameLen == 1 {
			if err := setValue(ctx, iterScope, valueSlot, val); err != nil {
				return err
			}
		} else if nameLen == 2 {
			if err := setValue(ctx, iterScope, keySlot, key); err != nil {
				return err
			}
			if err := setValue(ctx, iterScope, valueSlot, val); err != nil {
				return err
			}
		}
		return evalStmtBlock(ctx, iterScope, node.IterateBlock)
	}

	// define indication variables as "currentKey" and "currentValue" under new iterScope
	// of course since there's no any iteration is executed yet, the initial values are all "Null"
	if nameLen == 1 {
		valueSlot = node.IndexNames[0].Literal
		if err := bindValue(ctx, iterScope, valueSlot, NewZnNull(), false); err != nil {
			return err
		}
	} else if nameLen == 2 {
		keySlot = node.IndexNames[0].Literal
		valueSlot = node.IndexNames[1].Literal
		if err := bindValue(ctx, iterScope, keySlot, NewZnNull(), false); err != nil {
			return err
		}
		if err := bindValue(ctx, iterScope, valueSlot, NewZnNull(), false); err != nil {
			return err
		}
	} else if nameLen > 2 {
		return error.MostParamsError(2)
	}

	// execute iterations
	switch tv := targetExpr.(type) {
	case *ZnArray:
		for idx, val := range tv.Value {
			idxVar := NewZnDecimalFromInt(idx, 0)
			if err := execIterationBlockFn(idxVar, val); err != nil {
				if err.GetCode() == error.ContinueBreakSignal {
					// continue next turn
					continue
				}
				if err.GetCode() == error.BreakBreakSignal {
					// break directly
					return nil
				}
				return err
			}
		}
	case *ZnHashMap:
		for _, key := range tv.KeyOrder {
			val := tv.Value[key]
			keyVar := NewZnString(key)
			// handle interrupts
			if err := execIterationBlockFn(keyVar, val); err != nil {
				if err.GetCode() == error.ContinueBreakSignal {
					// continue next turn
					continue
				}
				if err.GetCode() == error.BreakBreakSignal {
					// break directly
					return nil
				}
				return err
			}
		}
	default:
		return error.InvalidExprType("array", "hashmap")
	}
	return nil
}

//// execute expressions

func evalExpression(ctx *Context, scope Scope, expr syntax.Expression) (ZnValue, *error.Error) {
	scope.GetRoot().SetCurrentLine(expr.GetCurrentLine())
	switch e := expr.(type) {
	case *syntax.VarAssignExpr:
		return evalVarAssignExpr(ctx, scope, e)
	case *syntax.LogicExpr:
		if e.Type == syntax.LogicAND || e.Type == syntax.LogicOR {
			return evalLogicCombiner(ctx, scope, e)
		}
		return evalLogicComparator(ctx, scope, e)
	case *syntax.MemberExpr:
		iv, err := getMemberExprIV(ctx, scope, e)
		if err != nil {
			return nil, err
		}
		return iv.Reduce(ctx, scope, nil, false)
	case *syntax.Number, *syntax.String, *syntax.ID, *syntax.ArrayExpr, *syntax.HashMapExpr:
		return evalPrimeExpr(ctx, scope, e)
	case *syntax.FuncCallExpr:
		return evalFunctionCall(ctx, scope, e)
	default:
		return nil, error.InvalidExprType()
	}
}

// （显示：A，B，C）
func evalFunctionCall(ctx *Context, scope Scope, expr *syntax.FuncCallExpr) (ZnValue, *error.Error) {
	vtag := expr.FuncName.GetLiteral()
	var zf *ClosureRef

	// if current scope is FuncScope, find ID from funcScope's "targetThis" method list
	if sp, ok := scope.(*FuncScope); ok {
		targetThis := sp.GetTargetThis()
		if targetThis != nil {
			if val, err := targetThis.GetMethod(vtag); err == nil {
				zf = val
			}
		}
	}

	// if function value not found from object scope, look up from local scope
	if zf == nil {
		// find function definction
		val, err := getValue(ctx, scope, vtag)
		if err != nil {
			return nil, err
		}
		// assert value
		zval, ok := val.(*ZnFunction)
		if !ok {
			return nil, error.InvalidFuncVariable(vtag)
		}
		zf = zval.ClosureRef
	}

	// exec params
	params, err := exprsToValues(ctx, scope, expr.Params)
	if err != nil {
		return nil, err
	}

	fScope := NewFuncScope(scope, nil)
	// exec function call via its ClosureRef
	return zf.Exec(ctx, fScope, params)
}

// evaluate logic combination expressions
// such as A 且 B
// or A 或 B
func evalLogicCombiner(ctx *Context, scope Scope, expr *syntax.LogicExpr) (*ZnBool, *error.Error) {
	logicType := expr.Type
	// #1. eval left
	left, err := evalExpression(ctx, scope, expr.LeftExpr)
	if err != nil {
		return nil, err
	}
	// #2. assert left expr type to be ZnBool
	vleft, ok := left.(*ZnBool)
	if !ok {
		return nil, error.InvalidExprType("bool")
	}
	// #3. check if the result could be retrieved earlier
	//
	// 1) for Y = A and B, if A = false, then Y must be false
	// 2) for Y = A or  B, if A = true, then Y must be true
	//
	// for those cases, we can yield result directly
	if logicType == syntax.LogicAND && vleft.Value == false {
		return NewZnBool(false), nil
	}
	if logicType == syntax.LogicOR && vleft.Value == true {
		return NewZnBool(true), nil
	}
	// #4. eval right
	right, err := evalExpression(ctx, scope, expr.RightExpr)
	if err != nil {
		return nil, err
	}
	vright, ok := right.(*ZnBool)
	if !ok {
		return nil, error.InvalidExprType("bool")
	}
	// then evalute data
	switch logicType {
	case syntax.LogicAND:
		return NewZnBool(vleft.Value && vright.Value), nil
	default: // logicOR
		return NewZnBool(vleft.Value || vright.Value), nil
	}
}

// evaluate logic comparator
// ensure both expressions are comparable (i.e. subtype of ZnComparable)
func evalLogicComparator(ctx *Context, scope Scope, expr *syntax.LogicExpr) (*ZnBool, *error.Error) {
	logicType := expr.Type
	// #1. eval left
	left, err := evalExpression(ctx, scope, expr.LeftExpr)
	if err != nil {
		return nil, err
	}
	// #2. eval right
	right, err := evalExpression(ctx, scope, expr.RightExpr)
	if err != nil {
		return nil, err
	}

	var cmpRes bool
	var cmpErr *error.Error
	// #3. do comparison
	switch logicType {
	case syntax.LogicEQ:
		cmpRes, cmpErr = compareValues(left, right, CmpEq)
	case syntax.LogicNEQ:
		cmpRes, cmpErr = compareValues(left, right, CmpEq)
		cmpRes = !cmpRes // reverse result
	case syntax.LogicGT:
		cmpRes, cmpErr = compareValues(left, right, CmpGt)
	case syntax.LogicGTE:
		var cmp1, cmp2 bool
		cmp1, cmpErr = compareValues(left, right, CmpGt)
		if cmpErr != nil {
			return nil, cmpErr
		}
		cmp2, cmpErr = compareValues(left, right, CmpEq)
		cmpRes = cmp1 || cmp2
	case syntax.LogicLT:
		cmpRes, cmpErr = compareValues(left, right, CmpLt)
	case syntax.LogicLTE:
		var cmp1, cmp2 bool
		cmp1, cmpErr = compareValues(left, right, CmpLt)
		if cmpErr != nil {
			return nil, cmpErr
		}
		cmp2, cmpErr = compareValues(left, right, CmpEq)
		cmpRes = cmp1 || cmp2
	default:
		return nil, error.UnExpectedCase("比较类型", strconv.Itoa(int(logicType)))
	}

	return NewZnBool(cmpRes), cmpErr
}

// eval prime expr
func evalPrimeExpr(ctx *Context, scope Scope, expr syntax.Expression) (ZnValue, *error.Error) {
	switch e := expr.(type) {
	case *syntax.Number:
		return NewZnDecimal(e.GetLiteral())
	case *syntax.String:
		return NewZnString(e.GetLiteral()), nil
	case *syntax.ID:
		vtag := e.GetLiteral()
		return getValue(ctx, scope, vtag)
	case *syntax.ArrayExpr:
		znObjs := []ZnValue{}
		for _, item := range e.Items {
			expr, err := evalExpression(ctx, scope, item)
			if err != nil {
				return nil, err
			}
			znObjs = append(znObjs, expr)
		}

		return NewZnArray(znObjs), nil
	case *syntax.HashMapExpr:
		znPairs := []KVPair{}
		for _, item := range e.KVPair {
			expr, err := evalExpression(ctx, scope, item.Key)
			if err != nil {
				return nil, err
			}
			exprKey, ok := expr.(*ZnString)
			if !ok {
				return nil, error.InvalidExprType("string")
			}
			exprVal, err := evalExpression(ctx, scope, item.Value)
			if err != nil {
				return nil, err
			}
			znPairs = append(znPairs, KVPair{
				Key:   exprKey.Value,
				Value: exprVal,
			})
		}
		return NewZnHashMap(znPairs), nil
	default:
		return nil, error.UnExpectedCase("表达式类型", reflect.TypeOf(e).Name())
	}
}

// eval var assign
func evalVarAssignExpr(ctx *Context, scope Scope, expr *syntax.VarAssignExpr) (ZnValue, *error.Error) {
	// Right Side
	val, err := evalExpression(ctx, scope, expr.AssignExpr)
	if err != nil {
		return nil, err
	}

	// Left Side
	switch v := expr.TargetVar.(type) {
	case *syntax.ID:
		// set ID
		vtag := v.GetLiteral()
		err2 := setValue(ctx, scope, vtag, val)
		return val, err2
	case *syntax.MemberExpr:
		iv, err := getMemberExprIV(ctx, scope, v)
		if err != nil {
			return nil, err
		}
		return iv.Reduce(ctx, scope, val, true)
	default:
		return nil, error.UnExpectedCase("被赋值", reflect.TypeOf(v).Name())
	}
}

func getMemberExprIV(ctx *Context, scope Scope, expr *syntax.MemberExpr) (ZnIV, *error.Error) {
	if expr.RootType == syntax.RootTypeScope { // 此之 XX
		switch expr.MemberType {
		case syntax.MemberID:
			tag := expr.MemberID.Literal
			return &ZnScopeMemberIV{tag}, nil
		case syntax.MemberMethod:
			m := expr.MemberMethod
			funcName := m.FuncName.Literal
			paramVals, err := exprsToValues(ctx, scope, m.Params)
			if err != nil {
				return nil, err
			}
			return &ZnScopeMethodIV{funcName, paramVals}, nil
		}
		return nil, error.UnExpectedCase("子项类型", strconv.Itoa(int(expr.MemberType)))
	}

	if expr.RootType == syntax.RootTypeProp { // 其 XX
		if expr.MemberType == syntax.MemberID {
			tag := expr.MemberID.Literal
			return &ZnPropIV{tag}, nil
		}
		return nil, error.UnExpectedCase("子项类型", strconv.Itoa(int(expr.MemberType)))
	}

	// RootType = RootTypeExpr
	valRoot, err := evalExpression(ctx, scope, expr.Root)
	if err != nil {
		return nil, err
	}
	switch expr.MemberType {
	case syntax.MemberID: // A 之 B
		tag := expr.MemberID.Literal
		return &ZnMemberIV{valRoot, tag}, nil
	case syntax.MemberMethod:
		m := expr.MemberMethod
		funcName := m.FuncName.Literal
		paramVals, err := exprsToValues(ctx, scope, m.Params)
		if err != nil {
			return nil, err
		}

		return &ZnMethodIV{valRoot, funcName, paramVals}, nil
	case syntax.MemberIndex:
		idx, err := evalExpression(ctx, scope, expr.MemberIndex)
		if err != nil {
			return nil, err
		}
		switch v := valRoot.(type) {
		case *ZnArray:
			vr, ok := idx.(*ZnDecimal)
			if !ok {
				return nil, error.InvalidExprType("integer")
			}
			return &ZnArrayIV{v, vr}, nil
		case *ZnHashMap:
			var s *ZnString
			switch x := idx.(type) {
			// regard decimal value directly as string
			case *ZnDecimal:
				// transform decimal value to string
				// x.exp < 0 express that its a decimal value with point mark, not an integer
				if x.exp < 0 {
					return nil, error.InvalidExprType("integer", "string")
				}
				s = NewZnString(x.String())
			case *ZnString:
				s = x
			default:
				return nil, error.InvalidExprType("integer", "string")
			}
			return &ZnHashMapIV{v, s}, nil
		default:
			return nil, error.InvalidExprType("array", "hashmap")
		}
	}
	return nil, error.UnExpectedCase("子项类型", reflect.TypeOf(expr.MemberType).Name())
}

//// scope value setters/getters
func getValue(ctx *Context, scope Scope, name string) (ZnValue, *error.Error) {
	// find on globals first
	if symVal, inGlobals := ctx.globals[name]; inGlobals {
		return symVal, nil
	}
	// ...then in symbols
	sp := scope
	for sp != nil {
		sym, ok := sp.GetSymbol(name)
		if ok {
			return sym.Value, nil
		}
		// if not found, search its parent
		sp = sp.GetParent()
	}
	return nil, error.NameNotDefined(name)
}

func setValue(ctx *Context, scope Scope, name string, value ZnValue) *error.Error {
	if _, inGlobals := ctx.globals[name]; inGlobals {
		return error.NameRedeclared(name)
	}
	// ...then in symbols
	sp := scope
	for sp != nil {
		sym, ok := sp.GetSymbol(name)
		if ok {
			if sym.IsConstant {
				return error.AssignToConstant()
			}
			sp.SetSymbol(name, value, false)
			return nil
		}
		// if not found, search its parent
		sp = sp.GetParent()
	}
	return error.NameNotDefined(name)
}

func getClassRef(ctx *Context, scope *RootScope, name string) (*ClassRef, *error.Error) {
	ref, ok := scope.classRefMap[name]
	if ok {
		return ref, nil
	}
	return nil, error.NameNotDefined(name)
}

func bindClassRef(ctx *Context, scope *RootScope, classStmt *syntax.ClassDeclareStmt) *error.Error {
	name := classStmt.ClassName.GetLiteral()
	_, ok := scope.classRefMap[name]
	if ok {
		return error.NameRedeclared(name)
	}
	scope.classRefMap[name] = NewClassRef(name, classStmt)
	return nil
}

func bindValue(ctx *Context, scope Scope, name string, value ZnValue, isConstatnt bool) *error.Error {
	if _, inGlobals := ctx.globals[name]; inGlobals {
		return error.NameRedeclared(name)
	}
	// bind directly
	if _, ok := scope.GetSymbol(name); ok {
		return error.NameRedeclared(name)
	}
	scope.SetSymbol(name, value, isConstatnt)
	return nil
}

//// helpers

// exprsToValues - []syntax.Expression -> []eval.ZnValue
func exprsToValues(ctx *Context, scope Scope, exprs []syntax.Expression) ([]ZnValue, *error.Error) {
	params := []ZnValue{}
	for _, paramExpr := range exprs {
		pval, err := evalExpression(ctx, scope, paramExpr)
		if err != nil {
			return nil, err
		}
		params = append(params, pval)
	}
	return params, nil
}
