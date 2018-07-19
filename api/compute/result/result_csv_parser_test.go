package result

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCSVResultParser(t *testing.T) {
	result, err := ParseResultCSV("./testdata/test.csv")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	fmt.Printf("%v", result)

	assert.Equal(t, []interface{}{"idx", "col a", "col b"}, result[0])
	assert.Equal(t, []interface{}{"0", []interface{}{"alpha", "bravo"}, "foxtrot"}, result[1])
	assert.Equal(t, []interface{}{"1", []interface{}{"charlie", "delta's oscar"}, "hotel"}, result[2])
	assert.Equal(t, []interface{}{"2", []interface{}{"a", "[", "b"}, []interface{}{"c", "\"", "e"}}, result[3])
	assert.Equal(t, []interface{}{"3", []interface{}{"a", "['\"", "b"}, []interface{}{"c", "\"", "e"}}, result[4])
	assert.Equal(t, []interface{}{"4", []interface{}{"-10.001", "20.1"}, []interface{}{"30", "40"}}, result[5])
	assert.Equal(t, []interface{}{"5", []interface{}{"int"}, []interface{}{"0.989599347114563"}}, result[6])
	assert.Equal(t, []interface{}{"7", []interface{}{"int", "categorical"}, []interface{}{"0.9885959029197693", "1"}}, result[8])
}
