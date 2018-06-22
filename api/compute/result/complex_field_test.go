package result

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserSingleQuoted(t *testing.T) {
	field := &ComplexField{Buffer: "  ['c ar'  , '\\'plane', 'b* oat']"} // single quote can be escaped in python
	field.Init()

	err := field.Parse()
	assert.NoError(t, err)

	field.Execute()
	assert.Equal(t, []interface{}{"c ar", "'plane", "b* oat"}, field.arrayElements.elements)
}

func TestParserDoubleQuoted(t *testing.T) {
	field := &ComplexField{Buffer: "[\"&car\"  , \"\\plane\", \"boat's\"]"}
	field.Init()

	err := field.Parse()
	assert.NoError(t, err)

	field.Execute()
	assert.Equal(t, []interface{}{"&car", "\\plane", "boat's"}, field.arrayElements.elements)
}

func TestParserValues(t *testing.T) {
	field := &ComplexField{Buffer: "[10, 20, 30, \"forty  &*\"]"}
	field.Init()

	err := field.Parse()
	field.PrintSyntaxTree()
	assert.NoError(t, err)

	field.Execute()
	assert.Equal(t, []interface{}{"10", "20", "30", "forty  &*"}, field.arrayElements.elements)
}

func TestParserFail(t *testing.T) {
	field := &ComplexField{Buffer: "[&*&, \"car\"  , \"plane\", \"boat's\"]"}
	field.Init()

	err := field.Parse()
	assert.Error(t, err)
}

func TestParserNested(t *testing.T) {
	field := &ComplexField{Buffer: "[[10, 20, 30, [alpha, bravo]], [40, 50, 60]]"}
	field.Init()

	err := field.Parse()
	field.PrintSyntaxTree()
	assert.NoError(t, err)

	field.Execute()

	assert.Equal(t, []interface{}{"alpha", "bravo"}, field.arrayElements.elements[0].([]interface{})[3].([]interface{}))
	assert.Equal(t, []interface{}{"40", "50", "60"}, field.arrayElements.elements[1].([]interface{}))
}
