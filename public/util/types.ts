import _ from "lodash";
import store from "../store/store";
import { Dictionary } from "./dict";
import { getters as datasetGetters } from "../store/dataset/module";
import { D3M_INDEX_FIELD } from "../store/dataset/index";

// TODO: Convert these to enums.
export const CLUSTER_PREFIX = "_cluster_";
export const GEOCODED_LAT_PREFIX = "_lat_";
export const GEOCODED_LON_PREFIX = "_lon_";
export const DATETIME_UNIX_ADJUSTMENT = 1000;

// Action Types Reuse In Similar Places to Data Types
// These are listed in the facet menu too, but Aren't Posted Back to
// as type change, but can take actions to change a compound facet
// back to its components, or change aspects of the facet's display

export const EXPLODE_ACTION_TYPE = "explode";
export const EXPAND_ACTION_TYPE = "expand";
export const COLLAPSE_ACTION_TYPE = "collapse";

// NOTE: these are copied from `distil-compute/model/schema_types.go` and
// should be kept up to date in case of changes.
// TODO: Convert these to enums.
export const ADDRESS_TYPE = "address";
export const INDEX_TYPE = "index";
export const INTEGER_TYPE = "integer";
export const REAL_TYPE = "real";
export const REAL_LIST_TYPE = "realList";
export const REAL_VECTOR_TYPE = "realVector";
export const BOOL_TYPE = "boolean";
export const DATE_TIME_TYPE = "dateTime";
export const DATE_TIME_LOWER_TYPE = "datetime";
export const TIMESTAMP_TYPE = "timestamp";
export const ORDINAL_TYPE = "ordinal";
export const CATEGORICAL_TYPE = "categorical";
export const TEXT_TYPE = "string";
export const CITY_TYPE = "city";
export const STATE_TYPE = "state";
export const COUNTRY_TYPE = "country";
export const COUNTRY_CODE_TYPE = "country_code";
export const EMAIL_TYPE = "email";
export const LATITUDE_TYPE = "latitude";
export const LONGITUDE_TYPE = "longitude";
export const PHONE_TYPE = "phone";
export const POSTAL_CODE_TYPE = "postal_code";
export const URI_TYPE = "uri";
export const IMAGE_TYPE = "image";
export const REMOTE_SENSING_TYPE = "remote_sensing";
export const TIMESERIES_TYPE = "timeseries";
export const UNKNOWN_TYPE = "unknown";
export const GEOCOORDINATE_TYPE = "geocoordinate";
export const GEOBOUNDS_TYPE = "geobounds";

const TYPES_TO_LABELS: Dictionary<string> = {
  [INTEGER_TYPE]: "Integer",
  [REAL_TYPE]: "Decimal",
  [REAL_LIST_TYPE]: "List",
  [REAL_VECTOR_TYPE]: "Vector",
  [LATITUDE_TYPE]: "Latitude",
  [LONGITUDE_TYPE]: "Longitude",
  [TEXT_TYPE]: "Text",
  [CATEGORICAL_TYPE]: "Categorical",
  [ORDINAL_TYPE]: "Ordinal",
  [ADDRESS_TYPE]: "Address",
  [CITY_TYPE]: "City",
  [STATE_TYPE]: "State/Province",
  [COUNTRY_TYPE]: "Country",
  [EMAIL_TYPE]: "Email",
  [PHONE_TYPE]: "Phone Number",
  [POSTAL_CODE_TYPE]: "Postal Code",
  [COUNTRY_CODE_TYPE]: "Country Code",
  [URI_TYPE]: "URI",
  [TIMESTAMP_TYPE]: "Timestamp",
  [DATE_TIME_TYPE]: "Date/Time",
  [BOOL_TYPE]: "Boolean",
  [IMAGE_TYPE]: "Image",
  [REMOTE_SENSING_TYPE]: "Satellite Image",
  [TIMESERIES_TYPE]: "Timeseries",
  [UNKNOWN_TYPE]: "Unknown",
  [GEOCOORDINATE_TYPE]: "Geocoordinate",
  [GEOBOUNDS_TYPE]: "Geobounds",
};

const LABELS_TO_TYPES = _.invert(TYPES_TO_LABELS);

const INTEGER_TYPES = [INTEGER_TYPE];

const COMPUTED_VAR_PREFIXES = [
  CLUSTER_PREFIX,
  GEOCODED_LAT_PREFIX,
  GEOCODED_LON_PREFIX,
];

const FLOATING_POINT_TYPES = [
  REAL_TYPE,
  REAL_LIST_TYPE,
  REAL_VECTOR_TYPE,
  LATITUDE_TYPE,
  LONGITUDE_TYPE,
  GEOBOUNDS_TYPE,
];

const LIST_TYPES = [REAL_LIST_TYPE, REAL_VECTOR_TYPE, GEOBOUNDS_TYPE];

const CLUSTER_TYPES = [
  IMAGE_TYPE,
  REMOTE_SENSING_TYPE,
  GEOCOORDINATE_TYPE,
  GEOBOUNDS_TYPE,
  TIMESERIES_TYPE,
];

const NUMERIC_TYPES = INTEGER_TYPES.concat(FLOATING_POINT_TYPES);

const TEXT_TYPES = [
  TEXT_TYPE,
  IMAGE_TYPE,
  REMOTE_SENSING_TYPE,
  TIMESERIES_TYPE,
  CATEGORICAL_TYPE,
  ORDINAL_TYPE,
  ADDRESS_TYPE,
  CITY_TYPE,
  STATE_TYPE,
  COUNTRY_TYPE,
  EMAIL_TYPE,
  PHONE_TYPE,
  POSTAL_CODE_TYPE,
  URI_TYPE,
  DATE_TIME_TYPE,
  BOOL_TYPE,
  UNKNOWN_TYPE,
];

const LOCATION_TYPES = [
  ADDRESS_TYPE,
  CITY_TYPE,
  STATE_TYPE,
  COUNTRY_TYPE,
  COUNTRY_CODE_TYPE,
  POSTAL_CODE_TYPE,
];

const REAL_COLLECTION_TYPES = [
  REAL_VECTOR_TYPE,
  REAL_LIST_TYPE,
  GEOBOUNDS_TYPE,
];

const TIME_TYPES = [DATE_TIME_TYPE, DATE_TIME_LOWER_TYPE, TIMESTAMP_TYPE];

const TEXT_SIMPLE_TYPES = [
  TEXT_TYPE,
  ADDRESS_TYPE,
  CITY_TYPE,
  STATE_TYPE,
  COUNTRY_TYPE,
  EMAIL_TYPE,
  PHONE_TYPE,
  POSTAL_CODE_TYPE,
  URI_TYPE,
  DATE_TIME_TYPE,
  BOOL_TYPE,
];

const GEOLOCATED_TYPES = [GEOCOORDINATE_TYPE, GEOBOUNDS_TYPE];

const IMAGE_TYPES = [IMAGE_TYPE, REMOTE_SENSING_TYPE];

const BOOL_SUGGESTIONS = [TEXT_TYPE, CATEGORICAL_TYPE, BOOL_TYPE, INTEGER_TYPE];

const EMAIL_SUGGESTIONS = [TEXT_TYPE, EMAIL_TYPE];

const URI_SUGGESTIONS = [TEXT_TYPE, URI_TYPE, UNKNOWN_TYPE];

const TIME_SUGGESTIONS = [DATE_TIME_TYPE, TEXT_TYPE, CATEGORICAL_TYPE];

const PHONE_SUGGESTIONS = [TEXT_TYPE, INTEGER_TYPE, PHONE_TYPE];

const TEXT_SUGGESTIONS = [
  TEXT_TYPE,
  CATEGORICAL_TYPE,
  INTEGER_TYPE,
  REAL_TYPE,
  ORDINAL_TYPE,
  ADDRESS_TYPE,
  CITY_TYPE,
  STATE_TYPE,
  COUNTRY_TYPE,
  POSTAL_CODE_TYPE,
  DATE_TIME_TYPE,
  IMAGE_TYPE,
];

const INTEGER_SUGGESTIONS = [
  INTEGER_TYPE,
  REAL_TYPE,
  LATITUDE_TYPE,
  LONGITUDE_TYPE,
  CATEGORICAL_TYPE,
  ORDINAL_TYPE,
  TIMESTAMP_TYPE,
];

const DECIMAL_SUGGESTIONS = [
  INTEGER_TYPE,
  REAL_TYPE,
  REAL_LIST_TYPE,
  REAL_VECTOR_TYPE,
  LATITUDE_TYPE,
  LONGITUDE_TYPE,
];

const COORDINATE_SUGGESTIONS = [
  INTEGER_TYPE,
  REAL_TYPE,
  REAL_LIST_TYPE,
  REAL_VECTOR_TYPE,
  LATITUDE_TYPE,
  LONGITUDE_TYPE,
  CATEGORICAL_TYPE,
  ORDINAL_TYPE,
];

const IMAGE_SUGGESTIONS = [IMAGE_TYPE, TEXT_TYPE, CATEGORICAL_TYPE];

const REAL_COLLECTION_SUGGESTIONS = [
  REAL_LIST_TYPE,
  REAL_VECTOR_TYPE,
  GEOBOUNDS_TYPE,
];

const LATITUDE_GROUPING_TYPES = [LATITUDE_TYPE, REAL_TYPE];

const LONGITUDE_GROUPING_TYPES = [LONGITUDE_TYPE, REAL_TYPE];

const TIME_GROUPING_TYPES = [INTEGER_TYPE, DATE_TIME_TYPE, TIMESTAMP_TYPE];

const VALUE_GROUPING_TYPES = [INTEGER_TYPE, REAL_TYPE];

export const BASIC_SUGGESTIONS = [
  INTEGER_TYPE,
  REAL_TYPE,
  CATEGORICAL_TYPE,
  ORDINAL_TYPE,
  TEXT_TYPE,
  IMAGE_TYPE,
  DATE_TIME_TYPE,
  TIMESTAMP_TYPE,
  UNKNOWN_TYPE,
];

const EQUIV_TYPES = {
  [INTEGER_TYPE]: [INTEGER_TYPE],
  [REAL_TYPE]: [REAL_TYPE],
  [REAL_VECTOR_TYPE]: [REAL_VECTOR_TYPE],
  [REAL_LIST_TYPE]: [REAL_LIST_TYPE],
  [LATITUDE_TYPE]: [LATITUDE_TYPE],
  [LONGITUDE_TYPE]: [LONGITUDE_TYPE],
  [TEXT_TYPE]: [TEXT_TYPE],
  [CATEGORICAL_TYPE]: [CATEGORICAL_TYPE],
  [ORDINAL_TYPE]: [ORDINAL_TYPE],
  [ADDRESS_TYPE]: [ADDRESS_TYPE],
  [CITY_TYPE]: [CITY_TYPE],
  [STATE_TYPE]: [STATE_TYPE],
  [COUNTRY_TYPE]: [COUNTRY_TYPE],
  [EMAIL_TYPE]: [EMAIL_TYPE],
  [PHONE_TYPE]: [PHONE_TYPE],
  [POSTAL_CODE_TYPE]: [POSTAL_CODE_TYPE],
  [URI_TYPE]: [URI_TYPE],
  [DATE_TIME_TYPE]: [DATE_TIME_TYPE, DATE_TIME_LOWER_TYPE],
  [DATE_TIME_LOWER_TYPE]: [DATE_TIME_TYPE, DATE_TIME_LOWER_TYPE],
  [BOOL_TYPE]: [BOOL_TYPE],
  [IMAGE_TYPE]: [IMAGE_TYPE],
  [REMOTE_SENSING_TYPE]: [REMOTE_SENSING_TYPE],
  [TIMESERIES_TYPE]: [TIMESERIES_TYPE],
  [UNKNOWN_TYPE]: [UNKNOWN_TYPE],
};

const TYPE_TO_SUGGESTIONS = {
  [ADDRESS_TYPE]: TEXT_SUGGESTIONS,
  [INDEX_TYPE]: TEXT_SUGGESTIONS,
  [INTEGER_TYPE]: INTEGER_SUGGESTIONS,
  [REAL_TYPE]: DECIMAL_SUGGESTIONS,
  [REAL_LIST_TYPE]: [...DECIMAL_SUGGESTIONS, ...REAL_COLLECTION_SUGGESTIONS],
  [REAL_VECTOR_TYPE]: [...DECIMAL_SUGGESTIONS, ...REAL_COLLECTION_SUGGESTIONS],
  [BOOL_TYPE]: BOOL_SUGGESTIONS,
  [DATE_TIME_TYPE]: TIME_SUGGESTIONS,
  [TIMESTAMP_TYPE]: TIME_SUGGESTIONS,
  [ORDINAL_TYPE]: TEXT_SUGGESTIONS,
  [CATEGORICAL_TYPE]: TEXT_SUGGESTIONS,
  [TEXT_TYPE]: TEXT_SUGGESTIONS,
  [CITY_TYPE]: TEXT_SUGGESTIONS,
  [STATE_TYPE]: TEXT_SUGGESTIONS,
  [COUNTRY_TYPE]: TEXT_SUGGESTIONS,
  [COUNTRY_CODE_TYPE]: TEXT_SUGGESTIONS,
  [EMAIL_TYPE]: EMAIL_SUGGESTIONS,
  [LATITUDE_TYPE]: COORDINATE_SUGGESTIONS,
  [LONGITUDE_TYPE]: COORDINATE_SUGGESTIONS,
  [PHONE_TYPE]: PHONE_SUGGESTIONS,
  [POSTAL_CODE_TYPE]: TEXT_SUGGESTIONS,
  [URI_TYPE]: URI_SUGGESTIONS,
  [IMAGE_TYPE]: IMAGE_SUGGESTIONS,
  [REMOTE_SENSING_TYPE]: REMOTE_SENSING_TYPE,
  [TIMESERIES_TYPE]: TIME_SUGGESTIONS,
  [GEOBOUNDS_TYPE]: [...REAL_COLLECTION_SUGGESTIONS, ...DECIMAL_SUGGESTIONS],
  [UNKNOWN_TYPE]: TEXT_SUGGESTIONS,
};

export const UNSUPPORTED_TARGET_TYPES = new Set([
  IMAGE_TYPE,
  REMOTE_SENSING_TYPE,
  GEOCOORDINATE_TYPE,
  GEOBOUNDS_TYPE,
  TEXT_TYPE,
]);

export const RANKABLE_VARIABLE_TYPES = new Set([
  ...INTEGER_TYPES,
  ...FLOATING_POINT_TYPES,
  BOOL_TYPE,
  DATE_TIME_TYPE,
  TIMESTAMP_TYPE,
  CATEGORICAL_TYPE,
  ORDINAL_TYPE,
  CITY_TYPE,
  STATE_TYPE,
  COUNTRY_TYPE,
  COUNTRY_CODE_TYPE,
]);

export function isEquivalentType(a: string, b: string): boolean {
  const equiv = EQUIV_TYPES[a];
  if (!equiv) {
    console.warn(
      `Unable to find equivalent types for type '${a}', type unrecognized`,
    );
    return false;
  }
  const matches = equiv.filter((type: string) => {
    return type === b;
  });
  return matches.length > 0;
}

export function normalizedEquivalentType(rawType: string): string {
  const normalizedType = EQUIV_TYPES[rawType];
  if (!normalizedType) {
    return rawType;
  }
  return normalizedType[0];
}

export function getVarType(varname: string): string {
  return datasetGetters.getVariableTypesMap(store)[varname];
}

export function formatValue(colValue: any, colType: string): any {
  // If there is no assigned schema, fix precision for a number, pass through otherwise.
  if (!colType || colType === "" || colType === D3M_INDEX_FIELD) {
    if (_.isNumber(colValue)) {
      return _.isInteger(colValue) ? colValue : colValue.toFixed(4);
    }
    return colValue;
  }

  // If the schema type is numeric and the value is a number stored as a string,
  // parse it and format again.
  if (
    isNumericType(colType) &&
    !_.isNumber(colValue) &&
    !_.isNaN(Number.parseFloat(colValue))
  ) {
    return formatValue(Number.parseFloat(colValue), colType);
  }

  // If the schema type is an integer, round.
  if (isIntegerType(colType)) {
    return Math.round(colValue).toFixed(0);
  }

  // If the schema type is text or not float, pass through.
  if (isTextType(colType) || !isFloatingPointType(colType)) {
    return colValue;
  }

  if (colValue === "") {
    return colValue;
  }

  // We've got a floating point value - set precision based on
  // type.
  switch (colType) {
    case LONGITUDE_TYPE:
    case LATITUDE_TYPE:
      return colValue.toFixed(6);
    case REAL_LIST_TYPE:
    case REAL_VECTOR_TYPE:
      return colValue;
  }
  return colValue.toFixed ? colValue.toFixed(4) : colValue;
}

export function isNumericType(type: string): boolean {
  return NUMERIC_TYPES.indexOf(type) !== -1;
}

export function isFloatingPointType(type: string): boolean {
  return FLOATING_POINT_TYPES.indexOf(type) !== -1;
}

export function isIntegerType(type: string): boolean {
  return INTEGER_TYPES.indexOf(type) !== -1;
}

export function isTextType(type: string): boolean {
  return TEXT_TYPES.indexOf(type) !== -1;
}

export function isTimeType(type: string): boolean {
  return TIME_TYPES.indexOf(type) !== -1;
}

export function isCollectionType(type: string): boolean {
  return REAL_COLLECTION_TYPES.indexOf(type) !== -1;
}

export function isLocationType(type: string): boolean {
  return LOCATION_TYPES.indexOf(type) !== -1;
}

export function isTextSimpleType(type: string): boolean {
  return TEXT_SIMPLE_TYPES.indexOf(type) !== -1;
}

export function isListType(type: string): boolean {
  return LIST_TYPES.indexOf(type) !== -1;
}

export function isGeoLocatedType(type: string): boolean {
  return GEOLOCATED_TYPES.indexOf(type) !== -1;
}

export function isImageType(type: string): boolean {
  return IMAGE_TYPES.indexOf(type) !== -1;
}

export function hasComputedVarPrefix(varName: string): boolean {
  return Boolean(
    COMPUTED_VAR_PREFIXES.find((prefix) => varName.indexOf(prefix) === 0),
  );
}

export function isClusterType(type: string): boolean {
  return CLUSTER_TYPES.indexOf(type) !== -1;
}

export function isLatitudeGroupType(type: string): boolean {
  return LATITUDE_GROUPING_TYPES.indexOf(type) !== -1;
}

export function isLongitudeGroupType(type: string): boolean {
  return LONGITUDE_GROUPING_TYPES.indexOf(type) !== -1;
}

export function isTimeGroupType(type: string): boolean {
  return TIME_GROUPING_TYPES.indexOf(type) !== -1;
}

export function isValueGroupType(type: string): boolean {
  return VALUE_GROUPING_TYPES.indexOf(type) !== -1;
}

export function addClusterPrefix(varName: string): string {
  return `${CLUSTER_PREFIX}${varName}`;
}

export function removeClusterPrefix(varName: string): string {
  return varName.replace(CLUSTER_PREFIX, "");
}

export function isJoinable(type: string, otherType: string): boolean {
  const isSameType = type === otherType;
  const isBothNumericType = isNumericType(type) && isNumericType(otherType);
  return isSameType || isBothNumericType;
}

/**
 * Returns true if a given variable can act as a target, false otherwise.
 */
export function isUnsupportedTargetVar(
  varName: string,
  varType: string,
): boolean {
  return UNSUPPORTED_TARGET_TYPES.has(varType) || hasComputedVarPrefix(varName);
}

/**
 * Returns ture if a given variable type can be processed as part of the feature ranking
 * pipeline, false otherwise.
 */
export function isRankableVariableType(varType: string): boolean {
  return RANKABLE_VARIABLE_TYPES.has(varType);
}

export function addTypeSuggestions(types: any[]): string[] {
  const suggestions = types.reduce((allSuggestions, type) => {
    allSuggestions = allSuggestions.concat(
      getSuggestionsByType(normalizedEquivalentType(type)),
    );
    return allSuggestions;
  }, []);
  suggestions.push(UNKNOWN_TYPE);
  return _.uniq(suggestions);
}

export function getSuggestionsByType(type: string): string[] {
  const types = TYPE_TO_SUGGESTIONS[type];
  if (types.length === 0) {
    return BASIC_SUGGESTIONS;
  }
  return types;
}

/**
 * Returns a UI-ready label for a given schema type.
 */
export function getLabelFromType(schemaType: string) {
  if (_.has(TYPES_TO_LABELS, schemaType)) {
    return TYPES_TO_LABELS[schemaType];
  }
  console.warn(
    `No label exists for type \`${schemaType}\` - using type as default label`,
  );
  return schemaType;
}

/**
 * Returns a schema type from a UI label
 */
export function getTypeFromLabel(label: string) {
  if (_.has(LABELS_TO_TYPES, label)) {
    return LABELS_TO_TYPES[label];
  }
  console.warn(`No type exists for label \`${label}\``);
  return label;
}
