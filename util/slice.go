package util

import (
	"reflect"
)

// SlicePredicateCallback checks whether element el passes filter condition(s)
// in the callback.
//
// i is el index in the slice.
type SlicePredicateCallback func(el interface{}, i int) bool

// SliceFilter return new slice interface with all elements in slice interface s
// that pass filter condition(s) in cb.
func SliceFilter(slice interface{}, cb SlicePredicateCallback) interface{} {
	sType := reflect.TypeOf(slice)

	if sType.Kind() != reflect.Slice {
		return nil
	}

	if cb == nil {
		return slice
	}

	var (
		sVal     = reflect.ValueOf(slice)
		newSlice = reflect.MakeSlice(sType, 0, sVal.Cap())
	)

	for i := 0; i < sVal.Len(); i++ {
		el := sVal.Index(i)

		pass := cb(el.Interface(), i)
		if pass {
			newSlice = reflect.Append(newSlice, el)
		}
	}

	return newSlice.Interface()
}

// SliceContain checks whether slice s contain element that deep equals
// element el.
//
// If s type is not slice, return false.
func SliceContain(slice, el interface{}) bool {
	sKind := reflect.TypeOf(slice).Kind()
	if sKind != reflect.Slice {
		return false
	}

	sVal := reflect.ValueOf(slice)
	for i := 0; i < sVal.Len(); i++ {
		sElem := sVal.Index(i)
		if reflect.DeepEqual(sElem.Interface(), el) {
			return true
		}
	}

	return false
}

// SliceFind returns first element that passes condition(s) in callback
// function.
//
// It returns nil if none of the elements pass.
func SliceFind(slice interface{}, cb SlicePredicateCallback) interface{} {
	sKind := reflect.TypeOf(slice).Kind()
	if sKind != reflect.Slice {
		return false
	}

	sVal := reflect.ValueOf(slice)
	for i := 0; i < sVal.Len(); i++ {
		el := sVal.Index(i)

		pass := cb(el.Interface(), i)
		if pass {
			return el.Interface()
		}
	}

	return nil
}

// SliceFindIndex returns index of the first element that passes condition(s)
// in callback function.
//
// It returns negative integer if none of the elements pass.
func SliceFindIndex(slice interface{}, cb SlicePredicateCallback) int {
	sKind := reflect.TypeOf(slice).Kind()
	if sKind != reflect.Slice {
		return -1
	}

	sVal := reflect.ValueOf(slice)
	for i := 0; i < sVal.Len(); i++ {
		el := sVal.Index(i)

		pass := cb(el.Interface(), i)
		if pass {
			return i
		}
	}

	return -1
}
