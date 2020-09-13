package util

// Copy - deep copy source []rune -> dst to produce a new slice.
func Copy(source []rune) []rune {
	newRune := make([]rune, len(source))

	for idx, item := range source {
		newRune[idx] = item
	}

	return newRune
}
