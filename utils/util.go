package utils

import (
	"fmt"
	"net/http"
	"regexp"
)

func EnvironmentEnabled(name string, enabled []string) bool {
	for _, p := range enabled {
		pattern, err := regexp.Compile(p)
		if err == nil {
			if pattern.MatchString(name) {
				return true
			}
		}
	}
	return false
}

func SendError(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content Type,", "text/plain; charset=UTF-8")
	w.WriteHeader(code)
	fmt.Fprint(w, error)
}
