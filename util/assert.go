package util

import (
	"fmt"
	"reflect"
)

func Assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}

func AssertEnum(msg string, variants ...any) {
	notNil := 0
	for _, v := range variants {
		if v == nil {
			continue
		}

		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Ptr && !val.IsNil() {
			notNil++
		}
	}

	if notNil != 1 {
		panic(msg)
	}
}

func Must[T any](val T, err error) T {
	if err != nil {
		panic(fmt.Errorf("expected no error: %w", err))
	}

	return val
}

func Has[T any](val T, ok bool) T {
	if !ok {
		panic(fmt.Sprintf("expected value of type %T to exist", val))
	}

	return val
}
