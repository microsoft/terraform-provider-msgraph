package utils

import "strings"

func LastSegment(input string) string {
	index := strings.LastIndex(input, "/")
	return input[index+1:]
}
