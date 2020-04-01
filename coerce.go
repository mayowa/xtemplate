package xtemplate

import (
	"reflect"
)

// IsSlice ...
func IsSlice(val interface{}) bool {
	t := reflect.TypeOf(val)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Kind() == reflect.Slice
}

// IsMap ...
func IsMap(val interface{}) bool {
	t := reflect.TypeOf(val)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Kind() == reflect.Map
}
