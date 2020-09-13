package lex

import (
	"fmt"
	"reflect"
	"testing"
)

type setIndentInput struct {
	count    int
	idetType IndentType
}

func TestLineStack_SetIndent(t *testing.T) {
	cases := []struct {
		name             string
		args             setIndentInput
		expectError      bool
		expectCursor     scanCursor
		expectIndentType IndentType
	}{
		{
			name: "empty indent",
			args: setIndentInput{
				count:    0,
				idetType: IdetUnknown,
			},
			expectError:      false,
			expectCursor:     scanCursor{0, 0, scanIndent},
			expectIndentType: IdetUnknown,
		},
		{
			name: "4 spaces",
			args: setIndentInput{
				count:    4,
				idetType: IdetSpace,
			},
			expectError:      false,
			expectCursor:     scanCursor{4, 1, scanIndent},
			expectIndentType: IdetSpace,
		},
		{
			name: "2 tabs",
			args: setIndentInput{
				count:    2,
				idetType: IdetTab,
			},
			expectError:      false,
			expectCursor:     scanCursor{2, 2, scanIndent},
			expectIndentType: IdetTab,
		},
		{
			name: "[ERROR] 3 spaces",
			args: setIndentInput{
				count:    3,
				idetType: IdetSpace,
			},
			expectError: true,
		},
	}

	for _, tt := range cases {
		ls := NewLineStack(nil)
		t.Run(tt.name, func(t *testing.T) {
			err := ls.SetIndent(tt.args.count, tt.args.idetType)

			if tt.expectError == false {
				if err != nil {
					t.Errorf("SetIndent() failed! expected no error, but got error")
					t.Error(err)
				} else if !reflect.DeepEqual(tt.expectCursor, ls.scanCursor) {
					t.Errorf("SetIndent() return scanState failed! expect: %v, got: %v", tt.expectCursor, ls.scanCursor)
				} else if !reflect.DeepEqual(tt.expectIndentType, ls.IndentType) {
					t.Errorf("SetIndent() return indentType failed! expect: %v, got: %v", tt.expectIndentType, ls.IndentType)
				}
			} else {
				if err == nil {
					t.Errorf("SetIndent() failed! expected error, but got no error")
				}
			}
		})
	}
}

// valid ops: append, setIndent, pushLine, newLine
type lineStackOp struct {
	opstr string
	r1    []rune
	num1  int
	u1    IndentType
}

// valid change ops: indentType, lines, scanCursor, lineBuffer
type lineStackChangeOp struct {
	opstr  string
	r1     []rune
	lines  []LineInfo
	cursor scanCursor
	u1     IndentType
}

func TestLineStack_LineStackSnapshot(t *testing.T) {
	// simulating parsing the following text:
	//
	// 如果它是真的：
	//     搞个大新闻
	//
	// 否则：
	//         不搞大新闻
	/** command list:

	ls.AppendLineBuffer([]rune("如果它是真的：\r\n    "))
	ls.SetIndent(0, IdetUnknown)
	ls.PushLine(6)

	// line 2
	ls.NewLine(9)
	ls.AppendLineBuffer([]rune("搞个大新闻\n\n否则：\n\r        不搞大新闻"))
	ls.SetIndent(4, IdetSpace)
	ls.PushLine(8)

	// line 3
	ls.NewLine(5)
	ls.SetIndent(0, IdetUnknown)
	ls.PushLine(0)

	// line 4
	ls.NewLine(1)
	ls.SetIndent(0, IdetUnknown)
	ls.PushLine(2)

	// line 5
	ls.NewLine(5)
	ls.SetIndent(8, IdetSpace)
	ls.PushLine(12)
	*/
	ls := NewLineStack(nil)

	flows := []struct {
		op       lineStackOp
		snapshot []lineStackChangeOp
	}{
		// line 1
		{
			op: lineStackOp{
				opstr: "append",
				r1:    []rune("如果它是真的：\r\n    "),
			},
			snapshot: []lineStackChangeOp{
				{
					opstr: "lineBuffer",
					r1:    []rune("如果它是真的：\r\n    "),
				},
			},
		},
		{
			op: lineStackOp{
				opstr: "setIndent",
				num1:  0,
				u1:    IdetUnknown,
			},
			snapshot: []lineStackChangeOp{
				{
					opstr:  "scanCursor",
					cursor: scanCursor{0, 0, scanIndent},
				},
			},
		},
		{
			op: lineStackOp{
				opstr: "pushLine",
				num1:  7,
			},
			snapshot: []lineStackChangeOp{
				{
					opstr: "lines",
					lines: []LineInfo{
						{0, 0, 7},
					},
				},
				{
					opstr:  "scanCursor",
					cursor: scanCursor{0, 0, scanEnd},
				},
			},
		},
		// line 2
		{
			op: lineStackOp{
				opstr: "newLine",
				num1:  9,
			},
			snapshot: []lineStackChangeOp{
				{
					opstr: "lineBuffer",
					r1:    []rune("如果它是真的：\r\n    "),
				},
				{
					opstr:  "cursor",
					cursor: scanCursor{9, 0, scanIndent},
				},
			},
		},
		{
			op: lineStackOp{
				opstr: "append",
				r1:    []rune("搞个大新闻\n\n否则：\n\r        不搞大新闻"),
			},
			snapshot: []lineStackChangeOp{
				{
					opstr: "lineBuffer",
					r1:    []rune("如果它是真的：\r\n    搞个大新闻\n\n否则：\n\r        不搞大新闻"),
				},
			},
		},
		// setIndent with known indentType will change the global indentType
		{
			op: lineStackOp{
				opstr: "setIndent",
				num1:  4,
				u1:    IdetSpace,
			},
			snapshot: []lineStackChangeOp{
				{
					opstr: "indentType",
					u1:    IdetSpace,
				},
			},
		},
		{
			op: lineStackOp{
				opstr: "pushLine",
				num1:  18,
			},
		},
		// line 3
		{
			op: lineStackOp{
				opstr: "newLine",
				num1:  19,
			},
			snapshot: []lineStackChangeOp{
				{
					opstr: "lines",
					lines: []LineInfo{
						{0, 0, 7},
						{1, 13, 18},
					},
				},
			},
		},
		{
			op: lineStackOp{
				opstr: "setIndent",
				num1:  0,
				u1:    IdetUnknown,
			},
		},
		{
			op: lineStackOp{
				opstr: "pushLine",
				num1:  19,
			},
		},
		// line 4
		{
			op: lineStackOp{
				opstr: "newLine",
				num1:  20,
			},
			snapshot: []lineStackChangeOp{
				{
					opstr: "lines",
					lines: []LineInfo{
						{0, 0, 7},
						{1, 13, 18},
						{0, 19, 19},
					},
				},
				{
					opstr: "lineBuffer",
					r1:    []rune("如果它是真的：\r\n    搞个大新闻\n\n否则：\n\r        不搞大新闻"),
				},
			},
		},
		{
			op: lineStackOp{
				opstr: "setIndent",
				num1:  0,
				u1:    IdetUnknown,
			},
		},
		{
			op: lineStackOp{
				opstr: "pushLine",
				num1:  23,
			},
		},
		// line 5
		{
			op: lineStackOp{
				opstr: "newLine",
				num1:  24,
			},
		},
		{
			op: lineStackOp{
				opstr: "setIndent",
				num1:  8,
				u1:    IdetSpace,
			},
		},
		{
			op: lineStackOp{
				opstr: "pushLine",
				num1:  37,
			},
			snapshot: []lineStackChangeOp{
				{
					opstr: "lines",
					lines: []LineInfo{
						{0, 0, 7},
						{1, 13, 18},
						{0, 19, 19},
						{0, 20, 23},
						{2, 32, 37},
					},
				},
			},
		},
	}
	// line 1

	for idx, flow := range flows {
		tag := fmt.Sprintf("step #%d: %s", idx+1, flow.op.opstr)
		t.Run(tag, func(t *testing.T) {
			assertLineBufferOps(t, ls, flow.op, flow.snapshot)
		})
	}
}

// test AppendLineBuffer, getChar, getLineBufferSize
func TestLineStack_LineBuffer(t *testing.T) {
	ls := NewLineStack(nil)

	ls.AppendLineBuffer([]rune("123456789"))
	ls.AppendLineBuffer([]rune("ABCD    EFGH"))

	got := ls.getLineBufferSize()
	if got != 21 {
		t.Errorf("buffer length should be %d, got %d", 21, got)
	}

	c1 := ls.getChar(4)
	if c1 != '5' {
		t.Errorf("c1 should be %v, got %v", '5', c1)
	}

	c2 := ls.getChar(100)
	if c2 != EOF {
		t.Errorf("c2 should be %v, got %v", EOF, c1)
	}
}

func assertLineBufferOps(t *testing.T, ls *LineStack, input lineStackOp, snaps []lineStackChangeOp) {
	// do operation
	switch input.opstr {
	case "append":
		ls.AppendLineBuffer(input.r1)
	case "setIndent":
		ls.SetIndent(input.num1, input.u1)
	case "pushLine":
		ls.PushLine(input.num1)
	case "newLine":
		ls.NewLine(input.num1)
	}

	// assert lineStackData
	for _, snap := range snaps {
		switch snap.opstr {
		case "indentType":
			if !reflect.DeepEqual(ls.IndentType, snap.u1) {
				t.Errorf("snap indentType failed! expect: %v, got: %v", snap.u1, ls.IndentType)
			}
		case "lines":
			if !reflect.DeepEqual(ls.lines, snap.lines) {
				t.Errorf("snap lines failed! expect: %v, got: %v", snap.lines, ls.lines)
			}
		case "scanCursor":
			if !reflect.DeepEqual(ls.scanCursor, snap.cursor) {
				t.Errorf("snap scanCursor failed! expect: %v, got: %v", snap.cursor, ls.scanCursor)
			}
		case "lineBuffer":
			if !reflect.DeepEqual(ls.lineBuffer, snap.r1) {
				t.Errorf("snap lineBuffer failed! expect: %v, got: %v", snap.r1, ls.lineBuffer)
			}
		}
	}
}
