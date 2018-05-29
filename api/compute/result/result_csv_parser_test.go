package compute

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
	assert.Equal(t, []interface{}{int64(0), []interface{}{"alpha", "bravo"}, "foxtrot"}, result[1])
	assert.Equal(t, []interface{}{int64(1), []interface{}{"charlie", "delta's oscar"}, "hotel"}, result[2])
}
