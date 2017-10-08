package seedr

import (
	"fmt"
	"unicode"
)

func panicf(s string, v ...interface{}) {
	panic(fmt.Sprintf(s, v...))
}

func toString(v interface{}) string {
	s, _ := v.(string)
	return s
}

func validString(s string, msg string) string {
	if s == "" {
		panic(msg)
	}
	return s
}

// toSnake convert the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
func toSnake(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}
	return string(out)
}

type stringSice []string

func (s stringSice) contains(str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
