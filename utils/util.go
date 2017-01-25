package utils

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
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

func GetEnvOrDefault(key, defaultValue string) string {
	if os.Getenv(key) != "" {
		return os.Getenv(key)
	}
	return defaultValue
}

func GetEnvOrDefaultArray(key string, defaultValues []string) []string {
	if os.Getenv(key) != "" {
		return strings.Split(os.Getenv(key), ",")
	}
	return defaultValues
}

func GetEnvOrDefaultInt(key string, defaultValue int) int {
	if os.Getenv(key) != "" {
		vals, err := strconv.Atoi(os.Getenv(key))
		if err != nil {
			log.Fatalf("Unable to parse %s [%s] as integer\n", key, os.Getenv(key))
		}
		return vals
	}
	return defaultValue
}

type RetryFunc func() (interface{}, error)

func Retry(f RetryFunc, timeout time.Duration, interval time.Duration) (interface{}, error) {
	finish := time.After(timeout)
	for {
		result, err := f()
		if err == nil {
			return result, nil
		}
		select {
		case <-finish:
			return nil, err
		case <-time.After(interval):
		}
	}
}

func SendError(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content Type,", "text/plain; charset=UTF-8")
	w.WriteHeader(code)
	fmt.Fprint(w, error)
}
