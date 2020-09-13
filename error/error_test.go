package error

import (
	"strings"
	"testing"

	"reflect"
)

func TestError_DisplayMasks(t *testing.T) {
	expError := Error{
		code: 0x2045,
		text: "这是一个测试用的异常信息",
		cursor: Cursor{
			File:    "draft/example.zn",
			LineNum: 3,
			ColNum:  4,
			Text:    "233如果梦想成真，：",
		},
	}

	testcases := []struct {
		name   string
		mask   uint16
		expect []string
	}{
		{
			name: "show all 0x00",
			mask: 0x0000,
			expect: []string{
				"在「draft/example.zn」中，位于第 3 行发现错误：",
				"    233如果梦想成真，：",
				"         ^",
				"‹2045› 语法错误：这是一个测试用的异常信息",
			},
		},
		{
			name: "hide file name 0x01",
			mask: dpHideFileName,
			expect: []string{
				"在第 3 行发现错误：",
				"    233如果梦想成真，：",
				"         ^",
				"‹2045› 语法错误：这是一个测试用的异常信息",
			},
		},
		{
			name: "hide line num 0x04",
			mask: dpHideLineNum,
			expect: []string{
				"在「draft/example.zn」中发现错误：",
				"    233如果梦想成真，：",
				"         ^",
				"‹2045› 语法错误：这是一个测试用的异常信息",
			},
		},
		{
			name: "hide line num & filename 0x05",
			mask: dpHideFileName | dpHideLineNum,
			expect: []string{
				"发现错误：",
				"    233如果梦想成真，：",
				"         ^",
				"‹2045› 语法错误：这是一个测试用的异常信息",
			},
		},
		{
			name: "hide line text 0x08",
			mask: dpHideLineText,
			expect: []string{
				"在「draft/example.zn」中，位于第 3 行发现错误：",
				"‹2045› 语法错误：这是一个测试用的异常信息",
			},
		},
		{
			name: "hide line cursor only 0x02",
			mask: dpHideLineCursor,
			expect: []string{
				"在「draft/example.zn」中，位于第 3 行发现错误：",
				"    233如果梦想成真，：",
				"    ",
				"‹2045› 语法错误：这是一个测试用的异常信息",
			},
		},
		{
			name: "hide line cursor and text 0x0A",
			mask: dpHideLineCursor | dpHideLineText,
			expect: []string{
				"在「draft/example.zn」中，位于第 3 行发现错误：",
				"‹2045› 语法错误：这是一个测试用的异常信息",
			},
		},
		{
			name: "hide err class only 0x10",
			mask: dpHideErrClass,
			expect: []string{
				"在「draft/example.zn」中，位于第 3 行发现错误：",
				"    233如果梦想成真，：",
				"         ^",
				"这是一个测试用的异常信息",
			},
		},
		{
			name: "hide all 0x1F",
			mask: dpHideErrClass | dpHideFileName | dpHideLineCursor | dpHideLineNum | dpHideLineText,
			expect: []string{
				"发现错误：",
				"这是一个测试用的异常信息",
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			expError.displayMask = tt.mask
			got := expError.Display()

			expectStr := strings.Join(tt.expect, "\n")
			if expectStr != got {
				t.Errorf("display result different:\n  expect ->\n%s\n  got->\n%s\n", expectStr, got)
			}
		})
	}
}

func TestError_CalcCursorOffset(t *testing.T) {
	text := "汉字TA汉字		245μg测试Ѣ2为什么"

	testcases := []struct {
		name   string
		cursor int
		expect int
	}{
		{
			name:   "negative number",
			cursor: -1,
			expect: -1,
		},
		{
			name:   "zero index",
			cursor: 0,
			expect: 0,
		},
		{
			name:   "first CJK char",
			cursor: 1,
			expect: 2,
		},
		{
			name:   "after tabs",
			cursor: 8,
			expect: 12,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			got := calcCursorOffset(text, tt.cursor)

			if got != tt.expect {
				t.Errorf("expect cursor = %d, got = %d", tt.expect, got)
			}
		})
	}
}

func TestError_GetInfo(t *testing.T) {
	testcases := []struct {
		name   string
		info   string
		expect map[string]string
	}{
		{
			name: "single item",
			info: "item=(100,200)",
			expect: map[string]string{
				"item": "100,200",
			},
		},
		{
			name: "multiple items",
			info: "left=(LEFT_WING)  right=(RIGHT_WING)",
			expect: map[string]string{
				"left":  "LEFT_WING",
				"right": "RIGHT_WING",
			},
		},
		{
			name: "ignore invalid syntax",
			info: "invalid_syntax hello=(\"World)\")",
			expect: map[string]string{
				"hello": "\"World)\"",
			},
		},
		{
			name: "with numbers and underscores",
			info: "wher123_49=((pig)(pot)) 2pig=(3pig)",
			expect: map[string]string{
				"wher123_49": "(pig)(pot)",
				"2pig":       "3pig",
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			e := Error{
				info: tt.info,
			}

			got := e.GetInfo()
			if !reflect.DeepEqual(tt.expect, got) {
				t.Errorf("expect info ->\n     %v\n got ->\n     %v", tt.expect, got)
			}
		})
	}
}
