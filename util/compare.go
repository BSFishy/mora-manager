package util

import (
	"fmt"
	"reflect"
)

func MatchDefinition(actual, expected any) (bool, error) {
	return matchValue(reflect.ValueOf(actual), reflect.ValueOf(expected))
}

func matchValue(actual, expected reflect.Value) (bool, error) {
	if !expected.IsValid() || isZero(expected) {
		// Treat zero values in expected as "don't care"
		return true, nil
	}
	if !actual.IsValid() {
		return false, fmt.Errorf("missing expected value: %v", expected.Interface())
	}

	if expected.Kind() == reflect.Ptr {
		if actual.Kind() != reflect.Ptr {
			return false, fmt.Errorf("expected ptr, got %s", actual.Kind())
		}
		if expected.IsNil() {
			return true, nil // don't care
		}
		return matchValue(actual.Elem(), expected.Elem())
	}

	switch expected.Kind() {
	case reflect.Map:
		if actual.Kind() != reflect.Map {
			return false, fmt.Errorf("expected map, got %s", actual.Kind())
		}
		for _, key := range expected.MapKeys() {
			expVal := expected.MapIndex(key)
			if isZero(expVal) {
				continue // don't care
			}
			actVal := actual.MapIndex(key)
			ok, err := matchValue(actVal, expVal)
			if !ok || err != nil {
				return false, fmt.Errorf("key %v mismatch: %w", key.Interface(), err)
			}
		}
		return true, nil

	case reflect.Slice, reflect.Array:
		if actual.Kind() != expected.Kind() {
			return false, fmt.Errorf("expected %s, got %s", expected.Kind(), actual.Kind())
		}
		if expected.Len() == 0 {
			return true, nil // don't care
		}
		if actual.Len() != expected.Len() {
			return false, fmt.Errorf("slice length mismatch: expected %d, got %d", expected.Len(), actual.Len())
		}
		for i := 0; i < expected.Len(); i++ {
			ok, err := matchValue(actual.Index(i), expected.Index(i))
			if !ok || err != nil {
				return false, fmt.Errorf("index %d mismatch: %w", i, err)
			}
		}
		return true, nil

	case reflect.Struct:
		if actual.Kind() != reflect.Struct {
			return false, fmt.Errorf("expected struct, got %s", actual.Kind())
		}
		for i := 0; i < expected.NumField(); i++ {
			field := expected.Type().Field(i)
			if field.PkgPath != "" {
				continue // unexported
			}
			expField := expected.Field(i)
			if isZero(expField) {
				continue // skip zero-valued fields
			}
			actField := actual.FieldByName(field.Name)
			ok, err := matchValue(actField, expField)
			if !ok || err != nil {
				return false, fmt.Errorf("field %s mismatch: %w", field.Name, err)
			}
		}
		return true, nil

	default:
		eq := reflect.DeepEqual(actual.Interface(), expected.Interface())
		if !eq {
			return false, fmt.Errorf("value mismatch: expected %#v, got %#v", expected.Interface(), actual.Interface())
		}
		return true, nil
	}
}

func isZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	if !v.CanInterface() {
		// Unexported or unsafe — treat as "don't care"
		return true
	}

	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()

	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZero(v.Index(i)) {
				return false
			}
		}
		return true

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if field.PkgPath != "" { // unexported
				continue
			}
			if !isZero(v.Field(i)) {
				return false
			}
		}
		return true

	default:
		zero := reflect.Zero(v.Type())
		if !zero.CanInterface() || !v.CanInterface() {
			return true // treat non-interfaceable as zero
		}
		return reflect.DeepEqual(v.Interface(), zero.Interface())
	}
}
