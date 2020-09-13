package util

import "testing"

func TestRuneStack(t *testing.T) {
	// run a cycle of a rune stack
	maxSize := 4
	stack := NewRuneStack(maxSize)

	// 01. first push all items
	items := []rune{2, 4, 6, 8, 10, 12}
	expectedResult := []bool{true, true, true, true, false, false}
	isEmptyResult := []bool{false, false, false, false, false, false}
	isFullResult := []bool{false, false, false, true, true, true}

	for idx, item := range items {
		er := stack.Push(item)
		em := stack.IsEmpty()
		ef := stack.IsFull()

		if er != expectedResult[idx] || em != isEmptyResult[idx] || ef != isFullResult[idx] {
			t.Errorf("push() item failed! index:%d, expected(res, isEmpty, isFull):(%v,%v,%v); actual(%v, %v,%v)",
				idx, expectedResult[idx], isEmptyResult[idx], isFullResult[idx],
				er, em, ef,
			)
		}
	}

	// 02. pop items
	expectedItems := []rune{8, 6, 4, 2}
	for idx, item := range expectedItems {
		data, _ := stack.Pop()

		if data != expectedItems[idx] {
			t.Errorf("pop() item failed! index:%d, expect:%v, actual:%v", idx, expectedItems[idx], item)
		}
	}
}
