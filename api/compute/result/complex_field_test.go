package compute

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserUnescaped(t *testing.T) {
	field := &ComplexField{Buffer: "  ['c ar'  , 'plane', 'b* oat']"}
	field.Init()

	err := field.Parse()
	assert.NoError(t, err)

	field.Execute()
	assert.Equal(t, []interface{}{"c ar", "plane", "b* oat"}, field.arrayElements.elements)
}

func TestParserEscaped(t *testing.T) {
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
