package main

import "strings"

func padStart(str string, targetLen int, padChar rune) string {
	if len(str) >= targetLen {
		return str
	}

	padding := strings.Repeat(string(padChar), targetLen-len(str))
	return padding + str
}
