package libhttp

import (
	"strings"
)

//SingleJoiningSlash joins urls.
func SingleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

//ToSlice converts the replaces to a string slice.
func ToSlice(in [][]string) []string {
	var out []string
	for _, each := range in {
		for _, elem := range each {
			out = append(out, elem)
		}
	}
	return out
}

//Contains looks for string in a []string slice.
func Contains(s []string, what string) bool {
	for _, a := range s {
		if a == what {
			return true
		}
	}
	return false
}
