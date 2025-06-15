package util

import (
	"fmt"
	"os"
)

func Getenv(name string) string {
	value, ok := os.LookupEnv(name)
	if ok {
		return value
	}

	file, ok := os.LookupEnv(fmt.Sprintf("%s_FILE", name))
	if !ok {
		return ""
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return ""
	}

	return string(data)
}

func GetenvDefault(name, def string) string {
	value := Getenv(name)
	if value == "" {
		value = def
	}

	return value
}
