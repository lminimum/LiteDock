// Package util provides shared utility functions for the assistant module.
package util

import "strings"

// StripShellChars removes shell metacharacters (|, ;, $, `, \, ", ') from input to prevent command injection.
func StripShellChars(input string) string {
	s := input
	s = strings.ReplaceAll(s, "|", "")
	s = strings.ReplaceAll(s, ";", "")
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "\\", "")
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, "'", "")
	return s
}
