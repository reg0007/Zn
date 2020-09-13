package util

// Contains - if one character (rune) inside a list
func Contains(ch rune, list []rune) bool {
	for _, item := range list {
		if item == ch {
			return true
		}
	}

	return false
}

// ContainsInt - if one character (int) inside a list
func ContainsInt(input int, list []int) bool {
	for _, item := range list {
		if item == input {
			return true
		}
	}

	return false
}
