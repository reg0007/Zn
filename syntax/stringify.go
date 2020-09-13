package syntax

import (
	"fmt"
	"strings"
)

// StringifyAST - stringify an abstract ProgramNode into a readable string
func StringifyAST(node Node) string {
	switch v := node.(type) {
	case *Program:
		return fmt.Sprintf("$PG(%s)", StringifyAST(v.Content))
	// expressions
	case *ArrayExpr:
		var exprs = []string{}
		for _, expr := range v.Items {
			exprs = append(exprs, StringifyAST(expr))
		}
		return fmt.Sprintf("$ARR(%s)", strings.Join(exprs, " "))
	case *HashMapExpr:
		var exprs = []string{}
		for _, expr := range v.KVPair {
			exprs = append(exprs, fmt.Sprintf("key[]=(%s) value[]=(%s)", StringifyAST(expr.Key), StringifyAST(expr.Value)))
		}
		return fmt.Sprintf("$HM(%s)", strings.Join(exprs, " "))
	case *Number:
		return fmt.Sprintf("$NUM(%s)", v.Literal)
	case *String:
		return fmt.Sprintf("$STR(%s)", v.Literal)
	case *ID:
		return fmt.Sprintf("$ID(%s)", v.Literal)
	case *LogicExpr:
		var typeStrMap = map[LogicTypeE]string{
			LogicEQ:  "$EQ",
			LogicNEQ: "$NEQ",
			LogicAND: "$AND",
			LogicOR:  "$OR",
			LogicGT:  "$GT",
			LogicGTE: "$GTE",
			LogicLT:  "$LT",
			LogicLTE: "$LTE",
		}

		lstr := StringifyAST(v.LeftExpr)
		rstr := StringifyAST(v.RightExpr)
		return fmt.Sprintf("%s(L=(%s) R=(%s))", typeStrMap[v.Type], lstr, rstr)
	case *MemberExpr:
		var str = ""
		var sType = ""
		switch v.MemberType {
		case MemberID:
			str = StringifyAST(v.MemberID)
			sType = "mID"
		case MemberIndex:
			str = StringifyAST(v.MemberIndex)
			sType = "mIndex"
		case MemberMethod:
			str = StringifyAST(v.MemberMethod)
			sType = "mMethod"
		}
		rootTypeStr := "rootScope"
		if v.RootType == RootTypeProp || v.RootType == RootTypeScope {
			if v.RootType == RootTypeProp {
				rootTypeStr = "rootProp"
			}
			return fmt.Sprintf("$MB(%s type=(%s) object=(%s))", rootTypeStr, sType, str)
		}
		rootStr := StringifyAST(v.Root)
		return fmt.Sprintf("$MB(root=(%s) type=(%s) object=(%s))", rootStr, sType, str)
	case *EmptyStmt:
		return "$"
	case *VarDeclareStmt:
		var items = []string{}
		// parse vars
		for _, vpair := range v.AssignPair {
			var vars = []string{}
			var expr string
			for _, vd := range vpair.Variables {
				vars = append(vars, StringifyAST(vd))
			}

			switch vpair.Type {
			case VDTypeAssign, VDTypeAssignConst:
				expr = StringifyAST(vpair.AssignExpr)
				if vpair.Type == VDTypeAssignConst {
					items = append(items, fmt.Sprintf("$VP(const vars[]=(%s) expr[]=(%s))", strings.Join(vars, " "), expr))
				} else {
					items = append(items, fmt.Sprintf("$VP(vars[]=(%s) expr[]=(%s))", strings.Join(vars, " "), expr))
				}
			case VDTypeObjNew:
				paramsStr := []string{}
				for _, vp := range vpair.ObjParams {
					paramsStr = append(paramsStr, StringifyAST(vp))
				}
				items = append(items, fmt.Sprintf(
					"$VP(object vars[]=(%s) class=(%s) params[]=(%s))",
					strings.Join(vars, " "),
					StringifyAST(vpair.ObjClass),
					strings.Join(paramsStr, " "),
				))
			}
		}
		// parse exprs
		return fmt.Sprintf("$VD(%s)", strings.Join(items, " "))
	// var assign expressions
	case *VarAssignExpr:
		var target, assign string
		target = StringifyAST(v.TargetVar)
		assign = StringifyAST(v.AssignExpr)

		return fmt.Sprintf("$VA(target=(%s) assign=(%s))", target, assign)
	case *FuncCallExpr:
		var params = []string{}
		var name = StringifyAST(v.FuncName)

		for _, p := range v.Params {
			params = append(params, StringifyAST(p))
		}

		return fmt.Sprintf("$FN(name=(%s) params=(%s))", name, strings.Join(params, " "))
	case *BranchStmt:
		var conds = []string{}
		// add if-branch
		ifExpr := fmt.Sprintf("ifExpr=(%s)", StringifyAST(v.IfTrueExpr))
		ifBlock := fmt.Sprintf("ifBlock=(%s)", StringifyAST(v.IfTrueBlock))
		conds = append(conds, ifExpr, ifBlock)

		// add else-branch
		if v.HasElse {
			elseBlock := fmt.Sprintf("elseBlock=(%s)", StringifyAST(v.IfFalseBlock))
			conds = append(conds, elseBlock)
		}

		// add other branchs
		for idx, expr := range v.OtherExprs {
			otherExpr := fmt.Sprintf("otherExpr[]=(%s)", StringifyAST(expr))
			otherBlock := fmt.Sprintf("otherBlock[]=(%s)", StringifyAST(v.OtherBlocks[idx]))

			conds = append(conds, otherExpr, otherBlock)
		}

		return fmt.Sprintf("$IF(%s)", strings.Join(conds, " "))
	case *WhileLoopStmt:
		return fmt.Sprintf("$WL(expr=(%s) block=(%s))", StringifyAST(v.TrueExpr), StringifyAST(v.LoopBlock))
	case *FunctionReturnStmt:
		return fmt.Sprintf("$RT(%s)", StringifyAST(v.ReturnExpr))
	case *FunctionDeclareStmt:
		paramsStr := []string{}
		for _, p := range v.ParamList {
			paramsStr = append(paramsStr, StringifyAST(p))
		}

		return fmt.Sprintf("$FN(name=(%s) params=(%s) blockTokens=(%s))",
			StringifyAST(v.FuncName),
			strings.Join(paramsStr, " "),
			StringifyAST(v.ExecBlock))
	case *GetterDeclareStmt:
		return fmt.Sprintf("$GT(name=(%s) blockTokens=(%s))",
			StringifyAST(v.GetterName),
			StringifyAST(v.ExecBlock))
	case *BlockStmt:
		var statements = []string{}
		for _, stmt := range v.Children {
			statements = append(statements, StringifyAST(stmt))
		}
		return fmt.Sprintf("$BK(%s)", strings.Join(statements, " "))
	case *IterateStmt:
		paramsStr := []string{}
		for _, p := range v.IndexNames {
			paramsStr = append(paramsStr, StringifyAST(p))
		}
		return fmt.Sprintf("$IT(target=(%s) idxList=(%s) block=(%s))",
			StringifyAST(v.IterateExpr),
			strings.Join(paramsStr, " "),
			StringifyAST(v.IterateBlock))
	case *ClassDeclareStmt:
		constructorStr := []string{}
		propertyStr := []string{}
		methodStr := []string{}
		getterStr := []string{}

		for _, c := range v.ConstructorIDList {
			constructorStr = append(constructorStr, StringifyAST(c))
		}
		for _, p := range v.PropertyList {
			propertyStr = append(propertyStr, StringifyAST(p))
		}
		for _, m := range v.MethodList {
			methodStr = append(methodStr, StringifyAST(m))
		}
		for _, g := range v.GetterList {
			getterStr = append(getterStr, StringifyAST(g))
		}

		return fmt.Sprintf(
			"$CLS(name=(%s) properties=(%s) constructor=(%s) methods=(%s) getters=(%s))",
			StringifyAST(v.ClassName),
			strings.Join(propertyStr, " "),
			strings.Join(constructorStr, " "),
			strings.Join(methodStr, " "),
			strings.Join(getterStr, " "),
		)
	case *PropertyDeclareStmt:
		return fmt.Sprintf(
			"$PD(id=(%s) expr=(%s))",
			StringifyAST(v.PropertyID),
			StringifyAST(v.InitValue),
		)
	default:
		return ""
	}
}
