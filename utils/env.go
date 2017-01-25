package utils

import (
	"log"
	"os"
	"strconv"
	"strings"
)

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
