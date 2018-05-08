package model

const (
	// AddressType is the schema type for address values
	AddressType = "address"
	// IndexType is the schema type for index values
	IndexType = "index"
	// IntegerType is the schema type for int values
	IntegerType = "integer"
	// FloatType is the schema type for float values
	FloatType = "float"
	// BoolType is the schema type for bool values
	BoolType = "boolean"
	// DateTimeType is the schema type for date/time values
	DateTimeType = "dateTime"
	// OrdinalType is the schema type for ordinal values
	OrdinalType = "ordinal"
	// CategoricalType is the schema type for categorical values
	CategoricalType = "categorical"
	// NumericalType is the schema type for numerical values
	NumericalType = "numerical"
	// TextType is the schema type for text values
	TextType = "text"
	// CityType is the schema type for city values
	CityType = "city"
	// CountryType is the schema type for country values
	CountryType = "country"
	// EmailType is the schema type for email values
	EmailType = "email"
	// LatitudeType is the schema type for latitude values
	LatitudeType = "latitude"
	// LongitudeType is the schema type for longitude values
	LongitudeType = "longitude"
	// PhoneType is the schema type for phone values
	PhoneType = "phone"
	// PostalCodeType is the schema type for postal code values
	PostalCodeType = "postal_code"
	// StateType is the schema type for state values
	StateType = "state"
	// URIType is the schema type for URI values
	URIType = "uri"
	// ImageType is the schema type for Image values
	ImageType = "image"
	// UnknownType is the schema type for unknown values
	UnknownType = "unknown"
)

var (
	categoricalTypes = map[string]bool{
		CategoricalType: true,
		OrdinalType:     true,
		BoolType:        true,
		AddressType:     true,
		CityType:        true,
		CountryType:     true,
		EmailType:       true,
		PhoneType:       true,
		PostalCodeType:  true,
		StateType:       true,
		URIType:         true,
		UnknownType:     true}
	numericalTypes = map[string]bool{
		LongitudeType: true,
		LatitudeType:  true,
		FloatType:     true,
		IntegerType:   true,
		IndexType:     true,
		DateTimeType:  true}
	floatingPointTypes = map[string]bool{
		LongitudeType: true,
		LatitudeType:  true,
		FloatType:     true}
	ta2TypeMap = map[string]string{
		AddressType:     "string",
		IndexType:       "integer",
		IntegerType:     "integer",
		FloatType:       "real",
		BoolType:        "boolean",
		DateTimeType:    "dateTime",
		OrdinalType:     "categorical",
		CategoricalType: "categorical",
		NumericalType:   "real",
		TextType:        "string",
		CityType:        "string",
		CountryType:     "string",
		EmailType:       "string",
		LatitudeType:    "real",
		LongitudeType:   "real",
		PhoneType:       "string",
		PostalCodeType:  "string",
		StateType:       "string",
		URIType:         "string",
	}
)

// IsNumerical indicates whether or not a schema type is numeric for the purposes
// of analysis.
func IsNumerical(typ string) bool {
	return numericalTypes[typ]
}

// IsFloatingPoint indicates whether or not a schema type is a floating point
// value.
func IsFloatingPoint(typ string) bool {
	return floatingPointTypes[typ]
}

// IsCategorical indicates whether or not a schema type is categorical for the purposes
// of analysis.
func IsCategorical(typ string) bool {
	return categoricalTypes[typ]
}

// IsText indicates whether or not a schema type is text for the purposes
// of analysis.
func IsText(typ string) bool {
	return typ == TextType
}

// MapTA2Type maps a type to a simple type.
func MapTA2Type(typ string) string {
	return ta2TypeMap[typ]
}
