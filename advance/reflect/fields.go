package reflect

import (
	"errors"
	"fmt"
	"reflect"
)

func IterateFields(val any) {
	res, err := iterateFields(val)

	if err != nil {
		fmt.Println(err)
		return
	}
	for k, v := range res {
		fmt.Println(k, v)
	}
}

func iterateFields(val any) (map[string]any, error) {
	if val == nil {
		return nil, errors.New("不能为nil")
	}

	typ := reflect.TypeOf(val)
	refVal := reflect.ValueOf(val)
	numField := typ.NumField()
	res := make(map[string]any, numField)
	for i := 0; i < numField; i++ {
		fdType := typ.Field(i)
		res[fdType.Name] = refVal.Field(i).Interface()
	}
	return res, nil
}
