package model

const (
	// IntegerType is the schema type for int values
	IntegerType = "integer"
	// FloatType is the schema type for float values
	FloatType = "float"
	// DateTimeType is the schema type for date/time values
	DateTimeType = "dateTime"
	// OrdinalType is the schema type for ordinal values
	OrdinalType = "ordinal"
	// CategoricalType is the schema type for categorical values
	CategoricalType = "categorical"
	// TextType is the schema type for text values
	TextType = "text"
)

// IsNumerical indicates whether or not a schema type is numeric for the purposes
// of analysis.
func IsNumerical(typ string) bool {
	return typ == IntegerType ||
		typ == FloatType ||
		typ == DateTimeType
}

// IsCategorical indicates whether or not a schema type is categorical for the purposes
// of analysis.
func IsCategorical(typ string) bool {
	return typ == CategoricalType ||
		typ == OrdinalType ||
		typ == TextType
}
