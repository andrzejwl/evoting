package pow

import (
	"errors"
	"reflect"
)

func Exists(slice interface{}, val interface{}) bool {
	anySlice, ok := CreateAnyTypeSlice(slice)
	if !ok {
		panic(errors.New("provided slice argument is not a slice"))
	}
	for _, item := range anySlice {
		if item == val {
			return true
		}
	}
	return false
}

func CreateAnyTypeSlice(slice interface{}) ([]interface{}, bool) {
	val, ok := isSlice(slice)

	if !ok {
		return nil, false
	}

	sliceLen := val.Len()

	out := make([]interface{}, sliceLen)

	for i := 0; i < sliceLen; i++ {
		out[i] = val.Index(i).Interface()
	}

	return out, true
}

//Determine whether it is slcie data
func isSlice(arg interface{}) (val reflect.Value, ok bool) {
	val = reflect.ValueOf(arg)

	if val.Kind() == reflect.Slice {
		ok = true
	}

	return
}
