//
//   Copyright © 2021 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package json

import (
	"encoding/json"
	"io/ioutil"
)

// RawMessage is an alias for json.RawMessage
type RawMessage json.RawMessage

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
		return nil, false
	}
	val, ok := v.(map[string]interface{})
	if !ok {
		return nil, false
	}
	return val, true
}

// Struct fills a struct under the given path.
func Struct(j map[string]interface{}, arg interface{}, path ...string) bool {
	v, ok := get(j, path...)
	if !ok || v == nil {
		return false
	}
	bs, err := Marshal(v)
	if nil != err {
		return false
	}
	err = json.Unmarshal(bs, &arg)
	return err == nil
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

// StringDefault returns a string property under the given key, if it doesn't
// exist, it will return the provided default.
func StringDefault(json map[string]interface{}, def string, path ...string) string {
	v, ok := String(json, path...)
	if ok {
		return v
	}
	return def
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

// FloatDefault returns a float property under the given key, if it doesn't
// exist, it will return the provided default.
func FloatDefault(json map[string]interface{}, def float64, path ...string) float64 {
	v, ok := Float(json, path...)
	if ok {
		return v
	}
	return def
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

// IntDefault returns an int property under the given key, if it doesn't
// exist, it will return the provided default.
func IntDefault(json map[string]interface{}, def int, path ...string) int {
	v, ok := Int(json, path...)
	if ok {
		return v
	}
	return def
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

// DoubleArray returns an [][]map[string]interface{} property under the given key.
func DoubleArray(json map[string]interface{}, path ...string) ([][]map[string]interface{}, bool) {
	vs, ok := array(json, path...)
	if !ok {
		return nil, false
	}
	result := make([][]map[string]interface{}, len(vs))
	for j, v := range vs {
		val, ok := v.([]interface{})
		if !ok {
			return nil, false
		}
		vals := make([]map[string]interface{}, len(val))
		for i, v := range val {
			prop, ok := v.(map[string]interface{})
			if !ok {
				return nil, false
			}
			vals[i] = prop
		}
		result[j] = vals
	}
	return result, true
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

// MapToStruct converts a map[string]interface{} to its struct equivalent.
func MapToStruct(res interface{}, arg map[string]interface{}) error {
	bs, err := Marshal(arg)
	if nil != err {
		return nil
	}
	return json.Unmarshal(bs, &res)
}

// StructToMap converts a struct to its map[string]interface{} equivalent.
func StructToMap(arg interface{}) map[string]interface{} {
	bs, err := Marshal(arg)
	if nil != err {
		return nil
	}
	m, err := Unmarshal(bs)
	if nil != err {
		return nil
	}
	return m
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

// Copy will copy the JSON data deeply by value, this process involves
// marshalling and then unmarshalling the data.
func Copy(j map[string]interface{}) (map[string]interface{}, error) {
	bytes, err := Marshal(j)
	if err != nil {
		return nil, err
	}
	return Unmarshal(bytes)
}

// ReadFile will read a json file into a byte array and unmarshal it to a map[string]interface
func ReadFile(file string) (map[string]interface{}, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return Unmarshal(buffer)
}
