package model

const (
	// Internal Type Keys.  These reflect types produced by semantic typing
	// analytics, and are not the set that is consumable by a downstream TA2
	// system.

	// AddressType is the schema type for address values
	AddressType = "address"
	// IndexType is the schema type for index values
	IndexType = "index"
	// IntegerType is the schema type for int values
	IntegerType = "integer"
	// FloatType is the schema type for float values
	FloatType = "float"
	// RealType is the schema type for real values, and is equivalent to FloatType
	RealType = "real"
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
	// TimeSeriesType is the schema type for Image values
	TimeSeriesType = "timeseries"
	// UnknownType is the schema type for unknown values
	UnknownType = "unknown"

	// TA2 Semantic Type Keys - defined in
	// https://gitlab.com/datadrivendiscovery/d3m/blob/devel/d3m/metadata/schemas/v0/definitions.json
	// These are the agreed upond set of types that are consumable by a downstream TA2 system.

	// TA2StringType is the semantic type reprsenting a text/string
	TA2StringType = "http://schema.org/Text"
	// TA2IntegerType is the TA2 semantic type for an integer value
	TA2IntegerType = "http://schema.org/Integer"
	// TA2RealType is the TA2 semantic type for a real value
	TA2RealType = "http://schema.org/Float"
	// TA2BooleanType is the TA2 semantic type for a boolean value
	TA2BooleanType = "http://schema.org/Boolean"
	// TA2LocationType is the TA2 semantic type for a location value
	TA2LocationType = "https://metadata.datadrivendiscovery.org/types/Location"
	// TA2TimeType is the TA2 semantic type for a time value
	TA2TimeType = "https://metadata.datadrivendiscovery.org/types/Time"
	// TA2CategoricalType is the TA2 semantic type for categorical data
	TA2CategoricalType = "https://metadata.datadrivendiscovery.org/types/CategoricalData"
	// TA2OrdinalType is the TA2 semantic type for ordinal (ordered categorical) data
	TA2OrdinalType = "https://metadata.datadrivendiscovery.org/types/OrdinalData"

	// TA2 Role keys

	// TA2TargetType is the semantic type indicating a prediction target
	TA2TargetType = "https://metadata.datadrivendiscovery.org/types/Target"
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
		UnknownType:     true,
		TimeSeriesType:  true}
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
		AddressType:     TA2StringType,
		IndexType:       TA2IntegerType,
		IntegerType:     TA2IntegerType,
		FloatType:       TA2RealType,
		RealType:        TA2RealType,
		BoolType:        TA2BooleanType,
		DateTimeType:    TA2TimeType,
		OrdinalType:     TA2CategoricalType,
		CategoricalType: TA2CategoricalType,
		NumericalType:   TA2RealType,
		TextType:        TA2StringType,
		CityType:        TA2StringType,
		CountryType:     TA2StringType,
		EmailType:       TA2StringType,
		LatitudeType:    TA2RealType,
		LongitudeType:   TA2RealType,
		PhoneType:       TA2StringType,
		PostalCodeType:  TA2StringType,
		StateType:       TA2StringType,
		URIType:         TA2StringType,
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

// IsImage indicates whether or not a schema type is an image for the purposes
// of analysis.
func IsImage(typ string) bool {
	return typ == ImageType
}

// MapTA2Type maps a type to a simple type.
func MapTA2Type(typ string) string {
	return ta2TypeMap[typ]
}
