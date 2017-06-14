package json

import (
	"github.com/stretchr/testify/assert"
)

func JSON(t, t *testing.T, data string) map[string]interface{} {
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
	_, ok := json.Get(j, "test", "obj")
	assert.True(t, ok)

	// should return the root node if no path is provided
	j := JSON(t,
		`{
			"test": {
				"obj": {}
			}
		}`)
	val, ok := json.Get(j)
	assert.True(t, ok)
	assert.Equal(t, val, j)

	// return false if the value doesn't exist in the provided path
	j := JSON(t, `{}`)
	_, ok := json.Get(j, "missing", "path")
	assert.False(t, ok)
}

func TestExists(t *testing.T) {
	// should return true if the value exists in the provided path"
	j := JSON(t, `{
			"test": {
				"obj": {}
			}
		}`)
	exists := json.Exists(j, "test", "obj")
	assert.True(t, exists)

	// should return false if value does not exist in the provided path
	j := JSON(t,
		`{
			"test": {
				"obj": {}
			}
		}`)
	exists := json.Exists(j, "test", "missing")
	assert.False(t, exists)
}

func TestFloat(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"float": 123.0
			}
		}`)
	_, ok := json.Float(j, "test", "float")
	assert.True(t, ok)

	// should return a float64 if the value is a float
	j := JSON(t,
		`{
			"test": {
				"float": 123.0
			}
		}`)
	val, ok := json.Float(j, "test", "float")
	assert.True(t, ok)
	assert.Equal(t, val, 123.0)

	// should return a float64 if the value is an int
	j := JSON(t,
		`{
			"test": {
				"int": 123
			}
		}`)
	val, ok := json.Float(j, "test", "int")
	assert.True(t, ok)
	assert.Equal(t, val, 123.0)

	// should return a false value if value does not exist in the provided path
	j := JSON(t, `{}`)
	_, ok := json.Float(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not a float
	j := JSON(t,
		`{
			"test": {
				"string": "hello"
			}
		}`)
	_, ok := json.Float(j, "test", "string")
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
	_, ok := json.Int(j, "test", "int")
	assert.True(t, ok)

	// should return an int if the value is an int
	j := JSON(t,
		`{
			"test": {
				"int": 123
			}
		}`)
	val, ok := json.Int(j, "test", "int")
	assert.True(t, ok)
	assert.Equal(t, val, 123)

	// should return an int if the value is a float
	j := JSON(t,
		`{
			"test": {
				"float": 123.0
			}
		}`)
	val, ok := json.Int(j, "test", "float")
	assert.True(t, ok)
	assert.Equal(t, val, 123)

	// should return a false value if value does not exist in the provided path
	j := JSON(t, `{}`)
	_, ok := json.Int(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an int
	j := JSON(t,
		`{
			"test": {
				"string": "hello"
			}
		}`)
	_, ok := json.Int(j, "test", "string")
	assert.False(t, ok)
}

func TestString(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"string": "hello"
			}
		}`)
	_, ok := json.String(j, "test", "string")
	assert.True(t, ok)

	// should return a string if the value is a string
	j := JSON(t,
		`{
			"test": {
				"string": "hello"
			}
		}`)
	val, ok := json.String(j, "test", "string")
	assert.True(t, ok)
	assert.Equal(t, val, "hello")

	// should return a false value if value does not exist in the provided path
	j := JSON(t, `{}`)
	_, ok := json.String(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not a string
	j := JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok := json.String(j, "test", "int")
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
	_, ok := json.Bool(j, "test", "bool")
	assert.True(t, ok)

	// should return a bool if the value is a bool
	j := JSON(t,
		`{
			"test": {
				"bool": false
			}
		}`)
	val, ok := json.Bool(j, "test", "bool")
	assert.True(t, ok)
	assert.False(t, val)

	// should return a false value if value does not exist in the provided path
	j := JSON(t, `{}`)
	_, ok := json.Bool(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not a bool
	j := JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok := json.Bool(j, "test", "int")
	assert.False(t, ok)
}

func TestGetChild(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"child": {}
			}
		}`)
	_, ok := json.GetChild(j, "test", "child")
	assert.True(t, ok)

	// should return a map[string]interface{} if the value is a map[string]interface{}
	j := JSON(t,
		`{
			"test": {
				"child": {
					"a": "a",
					"b": "b"
				}
			}
		}`)
	val, ok := json.GetChild(j, "test", "child")
	assert.True(t, ok)
	assert.Equal(t, val["a"].(string), "a")
	assert.Equal(t, val["b"].(string), "b")

	// should return a false value if value does not exist in the provided path
	j := JSON(t, `{}`)
	_, ok := json.GetChild(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not a map[string]interface{}
	j := JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok := json.GetChild(j, "test", "int")
	assert.False(t, ok)
}

func TestArray(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"array": [0, 1, "hello", true]
			}
		}`)
	_, ok := json.Array(j, "test", "array")
	assert.True(t, ok)

	// should return a []interface{} if the value is an array
	j := JSON(t,
		`{
			"test": {
				"array": [0, 1, "hello", true]
			}
		}`)
	val, ok := json.Array(j, "test", "array")
	assert.True(t, ok)
	assert.Equal(t, len(val), 4)

	// should return a false value if value does not exist in the provided path
	j := JSON(t, `{}`)
	_, ok := json.Array(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an array
	j := JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok := json.Array(j, "test", "int")
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
	_, ok := json.FloatArray(j, "test", "array")
	assert.True(t, ok)

	// should return a []float64 if the value is a float array
	j := JSON(t,
		`{
			"test": {
				"array": [0, 1, 0.1, 0.2]
			}
		}`)
	val, ok := json.FloatArray(j, "test", "array")
	assert.True(t, ok)
	assert.Equal(t, val[0], 0.0)
	assert.Equal(t, val[1], 1.0)
	assert.Equal(t, val[2], 0.1)
	assert.Equal(t, val[3], 0.2)

	// should return a false value if value does not exist in the provided path
	j := JSON(t, `{}`)
	_, ok := json.FloatArray(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an array
	j := JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok := json.FloatArray(j, "test", "int")
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
	_, ok := json.IntArray(j, "test", "array")
	assert.True(t, ok)

	// should return a []int if the value is an int array
	j := JSON(t,
		`{
			"test": {
				"array": [0, 1, 0.1, 0.2]
			}
		}`)
	val, ok := json.IntArray(j, "test", "array")
	assert.True(t, ok)
	assert.Equal(t, val[0], 0)
	assert.Equal(t, val[1], 1)
	assert.Equal(t, val[2], 0)
	assert.Equal(t, val[3], 0)

	// should return a false value if value does not exist in the provided path
	j := JSON(t, `{}`)
	_, ok := json.IntArray(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an array
	j := JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok := json.IntArray(j, "test", "int")
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
	_, ok := json.StringArray(j, "test", "array")
	assert.True(t, ok)

	// should return a []string if the value is a string array
	j := JSON(t,
		`{
			"test": {
				"array": ["a", "b", "see", "dee"]
			}
		}`)
	val, ok := json.StringArray(j, "test", "array")
	assert.True(t, ok)
	assert.Equal(t, val[0], "a")
	assert.Equal(t, val[1], "b")
	assert.Equal(t, val[2], "see")
	assert.Equal(t, val[3], "dee")

	// should return a false value if value does not exist in the provided path
	j := JSON(t, `{}`)
	_, ok := json.StringArray(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an array
	j := JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok := json.StringArray(j, "test", "int")
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
	_, ok := json.BoolArray(j, "test", "array")
	assert.True(t, ok)

	// should return a []bool if the value is a bool array
	j := JSON(t,
		`{
			"test": {
				"array": [true, false, false, true]
			}
		}`)
	val, ok := json.BoolArray(j, "test", "array")
	assert.True(t, ok)
	assert.True(t, val[0])
	assert.False(t, val[1])
	assert.False(t, val[2])
	assert.True(t, val[3])

	// should return a false value if value does not exist in the provided path
	j := JSON(t, `{}`)
	_, ok := json.BoolArray(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an array
	j := JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok := json.BoolArray(j, "test", "int")
	assert.False(t, ok)
}

func TestGetChildArray(t *testing.T) {
	// should return a true value if a value exists in the provided path
	j := JSON(t,
		`{
			"test": {
				"array": [{}, {}]
			}
		}`)
	_, ok := json.GetChildArray(j, "test", "array")
	assert.True(t, ok)

	// should return a []map[string]interface{} if the value is an array of nodes
	j := JSON(t,
		`{
			"test": {
				"array": [{
					"a": "a"
				}]
			}
		}`)
	val, ok := json.GetChildArray(j, "test", "array")
	assert.True(t, ok)
	assert.Equal(t, val[0]["a"], "a")

	// should return a false value if value does not exist in the provided path
	j := JSON(t, `{}`)
	_, ok := json.GetChildArray(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an array
	j := JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok := json.GetChildArray(j, "test", "int")
	assert.False(t, ok)

}

func TestGetChildMap(t *testing.T) {
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
	_, ok := json.GetChildMap(j, "test", "children")
	assert.True(t, ok)

	// should return a map[string]map[string]interface{} if the value is a map of nodes
	j := JSON(t,
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
	val, ok := json.GetChildMap(j, "test", "children")
	assert.True(t, ok)
	assert.Equal(t, val["a"]["val"], "a")
	assert.Equal(t, val["b"]["val"], "b")
	assert.Equal(t, val["c"]["val"], "c")

	// should return a false value if value does not exist in the provided path
	j := JSON(t, `{}`)
	_, ok := json.GetChildMap(j, "test", "missing")
	assert.False(t, ok)

	// should return a false value if value is not an array
	j := JSON(t,
		`{
			"test": {
				"int": 5
			}
		}`)
	_, ok := json.GetChildMap(j, "test", "int")
	assert.False(t, ok)
}

func TestGetRandomChild(t *testing.T) {
	// should return a true if there is at least one nested object
	j := JSON(t,
		`{
			"test": {
				"a": {},
				"b": {},
				"c": {}
			}
		}`)
	_, _, ok := json.GetRandomChild(j, "test")
	assert.True(t, ok)

	// should return a map[string]interface{} if there is at least one nested object
	j := JSON(t,
		`{
			"test": {
				"child" : {
					"a": "a"
				}
			}
		}`)
	key, val, ok := json.GetRandomChild(j, "test")
	assert.True(t, ok)
	assert.Equal(t, key, "child")
	assert.Equal(t, val["a"], "a")

	// should return a false if there are no nested objects
	j := JSON(t,
		`{
				"test": {}
			}`)
	_, _, ok := json.GetRandomChild(j, "test")
	assert.False(t, ok)
}
