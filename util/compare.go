package util

// Compare - compare two integers
// if a > b ==> 1
// if a == b ==> 0
// if a < b ==> -1
func Compare(a, b int) int {
	if a > b {
		return 1
	}
	if a == b {
		return 0
	}
	return -1
}
