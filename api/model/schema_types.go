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
	// RealVectorType is the schema type for a vector of real values
	RealVectorType = "realVector"
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
	// StringType is the schema type for Image values
	StringType = "string"
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
	// TA2ImageType is the TA2 semantic type for image data
	TA2ImageType = "http://schema.org/ImageObject"
	// TA2TimeSeriesType is the TA2 semantic type for timeseries data
	TA2TimeSeriesType = "https://metadata.datadrivendiscovery.org/types/Timeseries"

	// D3M Dataset Type Keys - these are the types used by for the D3M dataset format.

	// BooleanSchemaType is the schema doc type for boolean data
	BooleanSchemaType = "boolean"
	// IntegerSchemaType is the schema doc type for integer data
	IntegerSchemaType = "integer"
	// RealSchemaType is the schema doc type for real data
	RealSchemaType = "real"
	// StringSchemaType is the schema doc type for string/text data
	StringSchemaType = "string"
	// CategoricalSchemaType is the schema doc type for categorical data
	CategoricalSchemaType = "categorical"
	// DatetimeSchemaType is the schema doc type for datetime data
	DatetimeSchemaType = "dateTime"
	// RealVectorSchemaType is the schema doc type for a vector of real data
	RealVectorSchemaType = "realVector"
	// JSONSchemaType is the schema doc type for json data
	JSONSchemaType = "json"
	// GeoJSONSchemaType is the schema doc type for geo json data
	GeoJSONSchemaType = "geojson"
	// ImageSchemaType is the schema doc type for image data
	ImageSchemaType = "image"
	// TimeSeriesSchemaType is the schema doc type for image data
	TimeSeriesSchemaType = "timeseries"

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
		TimeSeriesType:  true,
	}
	numericalTypes = map[string]bool{
		LongitudeType: true,
		LatitudeType:  true,
		FloatType:     true,
		RealType:      true,
		IntegerType:   true,
		IndexType:     true,
	}
	floatingPointTypes = map[string]bool{
		LongitudeType: true,
		LatitudeType:  true,
		RealType:      true,
		FloatType:     true,
	}

	// Maps from Distil internal type to TA2 supported type
	ta2TypeMap = map[string]string{
		StringType:      TA2StringType,
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
		CityType:        TA2CategoricalType,
		CountryType:     TA2CategoricalType,
		EmailType:       TA2StringType,
		LatitudeType:    TA2RealType,
		LongitudeType:   TA2RealType,
		PhoneType:       TA2StringType,
		PostalCodeType:  TA2StringType,
		StateType:       TA2CategoricalType,
		URIType:         TA2StringType,
		ImageType:       TA2StringType,
		TimeSeriesType:  TA2TimeSeriesType,
		UnknownType:     TA2StringType,
	}

	// Maps from Distil internal type to D3M dataset doc type
	schemaTypeMap = map[string]string{
		StringType:      StringSchemaType,
		AddressType:     StringSchemaType,
		IndexType:       IntegerSchemaType,
		IntegerType:     IntegerSchemaType,
		FloatType:       RealSchemaType,
		RealType:        RealSchemaType,
		BoolType:        BooleanSchemaType,
		DateTimeType:    DatetimeSchemaType,
		OrdinalType:     CategoricalSchemaType,
		CategoricalType: CategoricalSchemaType,
		NumericalType:   RealSchemaType,
		TextType:        StringSchemaType,
		CityType:        StringSchemaType,
		CountryType:     StringSchemaType,
		EmailType:       StringSchemaType,
		LatitudeType:    RealSchemaType,
		LongitudeType:   RealSchemaType,
		PhoneType:       StringSchemaType,
		PostalCodeType:  StringSchemaType,
		StateType:       StringSchemaType,
		URIType:         StringSchemaType,
		ImageType:       ImageSchemaType,
		TimeSeriesType:  TimeSeriesSchemaType,
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
	return typ == TextType || typ == StringType
}

// IsVector indicates whether or not a schema type is a vector for the purposes
// of analysis.
func IsVector(typ string) bool {
	return typ == RealVectorType
}

// IsImage indicates whether or not a schema type is an image for the purposes
// of analysis.
func IsImage(typ string) bool {
	return typ == ImageType
}

// IsTimeSeries indicates whether or not a schema type is an timeseries for the purposes
// of analysis.
func IsTimeSeries(typ string) bool {
	return typ == TimeSeriesType
}

// IsDateTime indicates whether or not a schema type is a date time for the purposes
// of analysis.
func IsDateTime(typ string) bool {
	return typ == DateTimeType
}

// HasMetadataVar indicates whether or not a schema type has a corresponding metadata var.
func HasMetadataVar(typ string) bool {
	return IsImage(typ) || IsTimeSeries(typ)
}

// MapTA2Type maps an internal Distil type to a TA2 type.
func MapTA2Type(typ string) string {
	return ta2TypeMap[typ]
}

// MapSchemaType maps a type to a D3M dataset doc type.
func MapSchemaType(typ string) string {
	return schemaTypeMap[typ]
}
