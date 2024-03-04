package translate

import "strings"

// PackageName Get package name from Go SDK path
func PackageName(list []string) string {
	if len(list) == 0 {
		return ""
	}
	return list[len(list)-1]
}

// MakeIndentationEqual Check max lenght of the string in the list and then add spaces at the end of very name to make equal indentation
func MakeIndentationEqual(list []string) []string {
	maxLength := 0

	for _, str := range list {
		if len(str) > maxLength {
			maxLength = len(str)
		}
	}

	for idx, str := range list {
		list[idx] = str + strings.Repeat(" ", maxLength-len(str))
	}

	return list
}
