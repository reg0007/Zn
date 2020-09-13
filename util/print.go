package util

import "fmt"

//// print helpers

// PrintChar - print char with its hex value and string
func PrintChar(ch rune) {
	str := fmt.Sprintf("<0x%x %s>", ch, string([]rune{ch}))

	fmt.Println(str)
}
