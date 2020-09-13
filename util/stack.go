package util

// RuneStack - a LIFO stack with the following properties:
// 1. the stack size is limited, thus it's impossible
// to push an element into a full stack
type RuneStack struct {
	stack   []rune
	count   int
	maxSize int
}

// NewRuneStack - new stack
func NewRuneStack(maxSize int) *RuneStack {
	return &RuneStack{
		stack:   make([]rune, maxSize),
		count:   0,
		maxSize: maxSize,
	}
}

// Push - push item
// @returns bool - if data has been pushed successfully
func (rs *RuneStack) Push(item rune) bool {
	if rs.count == rs.maxSize {
		return false
	}
	rs.stack[rs.count] = item
	rs.count++
	return true
}

// Pop - pop item
func (rs *RuneStack) Pop() (rune, bool) {
	if rs.count == 0 {
		return 0, false
	}
	rs.count--
	return rs.stack[rs.count], true
}

// Current - current value
func (rs *RuneStack) Current() (rune, bool) {
	if rs.count == 0 {
		return 0, false
	}

	return rs.stack[rs.count-1], true
}

// IsEmpty - is empty
func (rs *RuneStack) IsEmpty() bool {
	return rs.count == 0
}

// IsFull - is full
func (rs *RuneStack) IsFull() bool {
	return rs.count == rs.maxSize
}

// GetMaxSize -
func (rs *RuneStack) GetMaxSize() int {
	return rs.maxSize
}
