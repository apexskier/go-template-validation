package main

import "strings"

// CountDigits returns the number of digits in a decimal number
func CountDigits(i int) (count int) {
	for i != 0 {
		i /= 10
		count = count + 1
	}
	return count
}

// SplitLines splits a string into lines
// Supports windows or unix line segments
func SplitLines(str string) []string {
	return strings.Split(strings.Replace(str, "\r\n", "\n", -1), "\n")
}
