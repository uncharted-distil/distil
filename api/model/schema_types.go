package model

//TODO: GET RID OF IntegerType IF WE CAN TO REDUCE CONFUSION!
const (
	// AddressType is the schema type for address values
	AddressType = "address"
	// IntType is the schema type for int values
	IntType = "int"
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
		URIType:         true}
	numericalTypes = map[string]bool{
		LongitudeType: true,
		LatitudeType:  true,
		FloatType:     true,
		IntType:       true,
		DateTimeType:  true}
)

// IsNumerical indicates whether or not a schema type is numeric for the purposes
// of analysis.
func IsNumerical(typ string) bool {
	return numericalTypes[typ]
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
