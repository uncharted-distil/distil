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
	"testing"

	"github.com/stretchr/testify/assert"
)

func JSON(t *testing.T, data string) map[string]interface{} {
	var j map[string]interface{}
	err := json.Unmarshal([]byte(data), &j)
	assert.Nil(t, err)
	return j
}

func TestGet(t *testing.T) {
	// should return true if the value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"obj": {}
			}
		}`)
	_, ok := Get(j, "test", "obj")
	assert.True(t, ok)

	// should return the root node if no path is provided
	j = JSON(t,
		`{
			"test": {
				"obj": {}
			}
		}`)
	val, ok := Get(j)
	assert.True(t, ok)
	assert.Equal(t, val, j)

	// return false if the value doesn't exist in the provided path
	j = JSON(t, `{}`)
	_, ok = Get(j, "missing", "path")
	assert.False(t, ok)

	// should return a true value if a value exists in the provided path
	j = JSON(t,
		`{
			"test": {
				"child": {}
			}
		}`)
	_, ok = Get(j, "test", "child")
	assert.True(t, ok)

	// should return a map[string]interface{} if the value is a map[string]interface{}
	j = JSON(t,
		`{
			"test": {
				"child": {
					"a": "a",
					"b": "b"
				}
			}
		}`)
	val, ok = Get(j, "test", "child")
	assert.True(t, ok)
	assert.Equal(t, val["a"].(string), "a")
	assert.Equal(t, val["b"].(string), "b")

	// should return a false value if value does not exist in the provided path
	j = JSON(t, `{}`)
	_, ok = Get(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not a map[string]interface{}
	j = JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok = Get(j, "test", "int")
	assert.False(t, ok)
}

func TestExists(t *testing.T) {
	// should return true if the value exists in the provided path"
	j := JSON(t, `{
			"test": {
				"obj": {}
			}
		}`)
	exists := Exists(j, "test", "obj")
	assert.True(t, exists)

	// should return false if value does not exist in the provided path
	j = JSON(t,
		`{
			"test": {
				"obj": {}
			}
		}`)
	exists = Exists(j, "test", "missing")
	assert.False(t, exists)
}

func TestInterface(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"float": 123.0
			}
		}`)
	_, ok := Interface(j, "test", "float")
	assert.True(t, ok)

	// should return a false value if value does not exist in the provided path
	j = JSON(t, `{}`)
	_, ok = Float(j, "test", "missing")
	assert.False(t, ok)
}

func TestFloat(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"float": 123.0
			}
		}`)
	_, ok := Float(j, "test", "float")
	assert.True(t, ok)

	// should return a float64 if the value is a float
	j = JSON(t,
		`{
			"test": {
				"float": 123.0
			}
		}`)
	val, ok := Float(j, "test", "float")
	assert.True(t, ok)
	assert.Equal(t, val, 123.0)

	// should return a float64 if the value is an int
	j = JSON(t,
		`{
			"test": {
				"int": 123
			}
		}`)
	val, ok = Float(j, "test", "int")
	assert.True(t, ok)
	assert.Equal(t, val, 123.0)

	// should return a false value if value does not exist in the provided path
	j = JSON(t, `{}`)
	_, ok = Float(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not a float
	j = JSON(t,
		`{
			"test": {
				"text": "hello"
			}
		}`)
	_, ok = Float(j, "test", "text")
	assert.False(t, ok)
}

func TestInt(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"int": 123
			}
		}`)
	_, ok := Int(j, "test", "int")
	assert.True(t, ok)

	// should return an int if the value is an int
	j = JSON(t,
		`{
			"test": {
				"int": 123
			}
		}`)
	val, ok := Int(j, "test", "int")
	assert.True(t, ok)
	assert.Equal(t, val, 123)

	// should return an int if the value is a float
	j = JSON(t,
		`{
			"test": {
				"float": 123.0
			}
		}`)
	val, ok = Int(j, "test", "float")
	assert.True(t, ok)
	assert.Equal(t, val, 123)

	// should return a false value if value does not exist in the provided path
	j = JSON(t, `{}`)
	_, ok = Int(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an int
	j = JSON(t,
		`{
			"test": {
				"text": "hello"
			}
		}`)
	_, ok = Int(j, "test", "text")
	assert.False(t, ok)
}

func TestString(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"text": "hello"
			}
		}`)
	_, ok := String(j, "test", "text")
	assert.True(t, ok)

	// should return a string if the value is a string
	j = JSON(t,
		`{
			"test": {
				"text": "hello"
			}
		}`)
	val, ok := String(j, "test", "text")
	assert.True(t, ok)
	assert.Equal(t, val, "hello")

	// should return a false value if value does not exist in the provided path
	j = JSON(t, `{}`)
	_, ok = String(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not a string
	j = JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok = String(j, "test", "int")
	assert.False(t, ok)
}

func TestBool(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"bool": true
			}
		}`)
	_, ok := Bool(j, "test", "bool")
	assert.True(t, ok)

	// should return a bool if the value is a bool
	j = JSON(t,
		`{
			"test": {
				"bool": false
			}
		}`)
	val, ok := Bool(j, "test", "bool")
	assert.True(t, ok)
	assert.False(t, val)

	// should return a false value if value does not exist in the provided path
	j = JSON(t, `{}`)
	_, ok = Bool(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not a bool
	j = JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok = Bool(j, "test", "int")
	assert.False(t, ok)
}

func TestInferfaceArray(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"array": [0, 1, 0.1, 0.2]
			}
		}`)
	_, ok := InterfaceArray(j, "test", "array")
	assert.True(t, ok)

	// should return a false value if value does not exist in the provided path
	j = JSON(t, `{}`)
	_, ok = InterfaceArray(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an array
	j = JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok = InterfaceArray(j, "test", "int")
	assert.False(t, ok)
}

func TestFloatArray(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"array": [0, 1, 0.1, 0.2]
			}
		}`)
	_, ok := FloatArray(j, "test", "array")
	assert.True(t, ok)

	// should return a []float64 if the value is a float array
	j = JSON(t,
		`{
			"test": {
				"array": [0, 1, 0.1, 0.2]
			}
		}`)
	val, ok := FloatArray(j, "test", "array")
	assert.True(t, ok)
	assert.Equal(t, val[0], 0.0)
	assert.Equal(t, val[1], 1.0)
	assert.Equal(t, val[2], 0.1)
	assert.Equal(t, val[3], 0.2)

	// should return a false value if value does not exist in the provided path
	j = JSON(t, `{}`)
	_, ok = FloatArray(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an array
	j = JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok = FloatArray(j, "test", "int")
	assert.False(t, ok)
}

func TestIntArray(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"array": [0, 1, 0.1, 0.2]
			}
		}`)
	_, ok := IntArray(j, "test", "array")
	assert.True(t, ok)

	// should return a []int if the value is an int array
	j = JSON(t,
		`{
			"test": {
				"array": [0, 1, 0.1, 0.2]
			}
		}`)
	val, ok := IntArray(j, "test", "array")
	assert.True(t, ok)
	assert.Equal(t, val[0], 0)
	assert.Equal(t, val[1], 1)
	assert.Equal(t, val[2], 0)
	assert.Equal(t, val[3], 0)

	// should return a false value if value does not exist in the provided path
	j = JSON(t, `{}`)
	_, ok = IntArray(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an array
	j = JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok = IntArray(j, "test", "int")
	assert.False(t, ok)
}

func TestStringArray(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"array": ["a", "b", "see", "dee"]
			}
		}`)
	_, ok := StringArray(j, "test", "array")
	assert.True(t, ok)

	// should return a []string if the value is a string array
	j = JSON(t,
		`{
			"test": {
				"array": ["a", "b", "see", "dee"]
			}
		}`)
	val, ok := StringArray(j, "test", "array")
	assert.True(t, ok)
	assert.Equal(t, val[0], "a")
	assert.Equal(t, val[1], "b")
	assert.Equal(t, val[2], "see")
	assert.Equal(t, val[3], "dee")

	// should return a false value if value does not exist in the provided path
	j = JSON(t, `{}`)
	_, ok = StringArray(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an array
	j = JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok = StringArray(j, "test", "int")
	assert.False(t, ok)
}

func TestBoolArray(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"array": [true, false, false, true]
			}
		}`)
	_, ok := BoolArray(j, "test", "array")
	assert.True(t, ok)

	// should return a []bool if the value is a bool array
	j = JSON(t,
		`{
			"test": {
				"array": [true, false, false, true]
			}
		}`)
	val, ok := BoolArray(j, "test", "array")
	assert.True(t, ok)
	assert.True(t, val[0])
	assert.False(t, val[1])
	assert.False(t, val[2])
	assert.True(t, val[3])

	// should return a false value if value does not exist in the provided path
	j = JSON(t, `{}`)
	_, ok = BoolArray(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an array
	j = JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok = BoolArray(j, "test", "int")
	assert.False(t, ok)
}

func TestArray(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"array": [{}, {}]
			}
		}`)
	_, ok := Array(j, "test", "array")
	assert.True(t, ok)

	// should return a []map[string]interface{} if the value is an array of nodes
	j = JSON(t,
		`{
			"test": {
				"array": [{
					"a": "a"
				}]
			}
		}`)
	val, ok := Array(j, "test", "array")
	assert.True(t, ok)
	assert.Equal(t, val[0]["a"], "a")

	// should return a false value if value does not exist in the provided path
	j = JSON(t, `{}`)
	_, ok = Array(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an array
	j = JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok = Array(j, "test", "int")
	assert.False(t, ok)

}

func TestMap(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"children": {
					"a": {},
					"b": {},
					"c": {}
				}
			}
		}`)
	_, ok := Map(j, "test", "children")
	assert.True(t, ok)

	// should return a map[string]map[string]interface{} if the value is a map of nodes
	j = JSON(t,
		`{
			"test": {
				"children": {
					"a": {
						"val": "a"
					},
					"b": {
						"val": "b"
					},
					"c": {
						"val": "c"
					}
				}
			}
		}`)
	val, ok := Map(j, "test", "children")
	assert.True(t, ok)
	assert.Equal(t, val["a"]["val"], "a")
	assert.Equal(t, val["b"]["val"], "b")
	assert.Equal(t, val["c"]["val"], "c")

	// should return a false value if value does not exist in the provided path
	j = JSON(t, `{}`)
	_, ok = Map(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an array
	j = JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok = Map(j, "test", "int")
	assert.False(t, ok)
}
