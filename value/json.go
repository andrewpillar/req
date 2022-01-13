package value

import (
	"io"
	"encoding/json"
)

func decodeJson(a interface{}) Value {
	var val Value = Zero{}

	switch v := a.(type) {
	case string:
		val = String{Value: v}
	case float64:
		val = Int{Value: int64(v)}
	case bool:
		val = Bool{Value: v}
	case []interface{}:
		arr := &Array{
			set:   make(map[uint32]struct{}),
			Items: make([]Value, 0, len(v)),
		}

		for _, a := range v {
			arr.Items = append(arr.Items, decodeJson(a))
		}

		arr.hashItems()

		val = arr
	case map[string]interface{}:
		obj := Object{
			Pairs: make(map[string]Value),
		}

		for k, a := range v {
			obj.Pairs[k] = decodeJson(a)
		}
		val = obj
	}
	return val
}

func DecodeJSON(r io.Reader) (Value, error) {
	var p interface{}

//	m := make(map[string]interface{})

	if err := json.NewDecoder(r).Decode(&p); err != nil {
		return nil, err
	}
	return decodeJson(p), nil

//	obj := Object{
//		Pairs: make(map[string]Value),
//	}
//
//	for k, v := range m {
//		obj.Pairs[k] = decodeJson(v)
//	}
//	return obj, nil
}
