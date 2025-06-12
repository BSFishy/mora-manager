package util

import (
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
