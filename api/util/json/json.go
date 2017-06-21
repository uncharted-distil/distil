package json

import (
	"encoding/json"
)

func get(json map[string]interface{}, path ...string) (interface{}, bool) {
	child := json
	last := len(path) - 1
	var val interface{} = child
	for index, key := range path {
		// does a child exists?
		v, ok := child[key]
		if !ok {
			return nil, false
		}
		// is it the target?
		if index == last {
			val = v
			break
		}
		// if not, does it have children to traverse?
		c, ok := v.(map[string]interface{})
		if !ok {
			return nil, false
		}
		child = c
	}
	return val, true
}

func array(json map[string]interface{}, path ...string) ([]interface{}, bool) {
	v, ok := get(json, path...)
	if !ok {
		return nil, false
	}
	val, ok := v.([]interface{})
	if !ok {
		return nil, false
	}
	return val, true
}

// Get returns a map[string]interface{} under the given path.
func Get(json map[string]interface{}, path ...string) (map[string]interface{}, bool) {
	v, ok := get(json, path...)
	if !ok {
		return nil, ok
	}
	val, ok := v.(map[string]interface{})
	if !ok {
		return nil, ok
	}
	return val, true
}

// Exists returns true if something exists under the provided path.
func Exists(json map[string]interface{}, path ...string) bool {
	_, ok := get(json, path...)
	return ok
}

// Interface returns an interface{} under the given path.
func Interface(json map[string]interface{}, path ...string) (interface{}, bool) {
	return get(json, path...)
}

// String returns a string property under the given path.
func String(json map[string]interface{}, path ...string) (string, bool) {
	v, ok := get(json, path...)
	if !ok {
		return "", false
	}
	val, ok := v.(string)
	if !ok {
		return "", false
	}
	return val, true
}

// Bool returns a bool property under the given key.
func Bool(json map[string]interface{}, path ...string) (bool, bool) {
	v, ok := get(json, path...)
	if !ok {
		return false, false
	}
	val, ok := v.(bool)
	if !ok {
		return false, false
	}
	return val, true
}

// Float returns a float property under the given key.
func Float(json map[string]interface{}, path ...string) (float64, bool) {
	v, ok := get(json, path...)
	if !ok {
		return 0, false
	}
	flt, ok := v.(float64)
	if !ok {
		return 0, false
	}
	return flt, true
}

// Int returns an int property under the given key.
func Int(json map[string]interface{}, path ...string) (int, bool) {
	v, ok := get(json, path...)
	if !ok {
		return 0, false
	}
	flt, ok := v.(float64)
	if !ok {
		return 0, false
	}
	return int(flt), true
}

// Array returns an []map[string]interface{} property under the given key.
func Array(json map[string]interface{}, path ...string) ([]map[string]interface{}, bool) {
	vs, ok := array(json, path...)
	if !ok {
		return nil, false
	}
	vals := make([]map[string]interface{}, len(vs))
	for i, v := range vs {
		val, ok := v.(map[string]interface{})
		if !ok {
			return nil, false
		}
		vals[i] = val
	}
	return vals, true
}

// InterfaceArray returns a []interface{} under the given path.
func InterfaceArray(json map[string]interface{}, path ...string) ([]interface{}, bool) {
	return array(json, path...)
}

// FloatArray returns a []float64 property under the given key.
func FloatArray(json map[string]interface{}, path ...string) ([]float64, bool) {
	vs, ok := array(json, path...)
	if !ok {
		return nil, false
	}
	flts := make([]float64, len(vs))
	for i, v := range vs {
		flt, ok := v.(float64)
		if !ok {
			return nil, false
		}
		flts[i] = flt
	}
	return flts, true
}

// IntArray returns an []int64 property under the given key.
func IntArray(json map[string]interface{}, path ...string) ([]int, bool) {
	vs, ok := array(json, path...)
	if !ok {
		return nil, false
	}
	ints := make([]int, len(vs))
	for i, v := range vs {
		flt, ok := v.(float64)
		if !ok {
			return nil, false
		}
		ints[i] = int(flt)
	}
	return ints, true
}

// StringArray returns an []string property under the given key.
func StringArray(json map[string]interface{}, path ...string) ([]string, bool) {
	vs, ok := array(json, path...)
	if !ok {
		return nil, false
	}
	strs := make([]string, len(vs))
	for i, v := range vs {
		val, ok := v.(string)
		if !ok {
			return nil, false
		}
		strs[i] = val
	}
	return strs, true
}

// BoolArray returns an []bool property under the given key.
func BoolArray(json map[string]interface{}, path ...string) ([]bool, bool) {
	vs, ok := array(json, path...)
	if !ok {
		return nil, false
	}
	bools := make([]bool, len(vs))
	for i, v := range vs {
		val, ok := v.(bool)
		if !ok {
			return nil, false
		}
		bools[i] = val
	}
	return bools, true
}

// Map returns a map[string]map[string]interface{} of all child nodes
// under the given path.
func Map(json map[string]interface{}, path ...string) (map[string]map[string]interface{}, bool) {
	sub, ok := Get(json, path...)
	if !ok {
		return nil, false
	}
	children := make(map[string]map[string]interface{})
	for k, v := range sub {
		c, ok := v.(map[string]interface{})
		if ok {
			children[k] = c
		}
	}
	return children, true
}

// Marshal marhsals JSON into a byte slice, convenience wrapper for the native
// package so no need to import both and get a name collision.
func Marshal(j interface{}) ([]byte, error) {
	return json.Marshal(j)
}

// Unmarshal unmarshals JSON and returns a newly instantiated map.
func Unmarshal(data []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal(data, &m)
	if nil != err {
		return nil, err
	}
	return m, nil
}

// UnmarshalArray unmarshals an array of JSON and returns a newly instantiated
// array of maps.
func UnmarshalArray(data []byte) ([]map[string]interface{}, error) {
	var arr []map[string]interface{}
	err := json.Unmarshal(data, &arr)
	if nil != err {
		return nil, err
	}
	return arr, nil
}

// Copy will copy the JSON data deeply by value, this process involves
// marshalling and then unmarshalling the data.
func Copy(j map[string]interface{}) (map[string]interface{}, error) {
	bytes, err := Marshal(j)
	if err != nil {
		return nil, err
	}
	return Unmarshal(bytes)
}
