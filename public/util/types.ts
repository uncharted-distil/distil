import _ from 'lodash';
import store from '../store/store';
import { Dictionary } from './dict';
import { getters as datasetGetters } from '../store/dataset/module';
import { D3M_INDEX_FIELD } from '../store/dataset/index';

const EMAIL_REGEX = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
const URI_REGEX = /^(?:(?:(?:https?|ftp):)?\/\/)(?:\S+(?::\S*)?@)?(?:(?!(?:10|127)(?:\.\d{1,3}){3})(?!(?:169\.254|192\.168)(?:\.\d{1,3}){2})(?!172\.(?:1[6-9]|2\d|3[0-1])(?:\.\d{1,3}){2})(?:[1-9]\d?|1\d\d|2[01]\d|22[0-3])(?:\.(?:1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.(?:[1-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(?:(?:[a-z\u00a1-\uffff0-9]-*)*[a-z\u00a1-\uffff0-9]+)(?:\.(?:[a-z\u00a1-\uffff0-9]-*)*[a-z\u00a1-\uffff0-9]+)*(?:\.(?:[a-z\u00a1-\uffff]{2,})))(?::\d{2,5})?(?:[/?#]\S*)?$/i;
const BOOL_REGEX = /^(0|1|true|false|t|f)$/i;
const PHONE_REGEX = /^(\+\d{1,2}\s)?\(?\d{3}\)?[\s.-]\d{3}[\s.-]\d{4}$/;
const IMAGE_REGEX = /\.(gif|jpg|jpeg|png|tif|tiff|bmp)$/i;

export const FEATURE_PREFIX = '_feature_';
export const CLUSTER_PREFIX = '_cluster_';
export const GEOCODED_LAT_PREFIX = '_lat_';
export const GEOCODED_LON_PREFIX = '_lon_';

// NOTE: these are copied from `distil-compute/model/schema_types.go` and
// should be kept up to date in case of changes.

export const ADDRESS_TYPE = 'address';
export const INDEX_TYPE = 'index';
export const INTEGER_TYPE = 'integer';
export const REAL_TYPE = 'real';
export const REAL_VECTOR_TYPE = 'realVector';
export const BOOL_TYPE = 'boolean';
export const DATE_TIME_TYPE = 'dateTime';
export const DATE_TIME_LOWER_TYPE = 'datetime';
export const TIMESTAMP_TYPE = 'timestmap';
export const ORDINAL_TYPE = 'ordinal';
export const CATEGORICAL_TYPE = 'categorical';
export const TEXT_TYPE = 'text';
export const CITY_TYPE = 'city';
export const STATE_TYPE = 'state';
export const COUNTRY_TYPE = 'country';
export const COUNTRY_CODE_TYPE = 'country_code';
export const EMAIL_TYPE = 'email';
export const LATITUDE_TYPE = 'latitude';
export const LONGITUDE_TYPE = 'longitude';
export const PHONE_TYPE = 'phone';
export const POSTAL_CODE_TYPE = 'postal_code';
export const URI_TYPE = 'uri';
export const IMAGE_TYPE = 'image';
export const TIMESERIES_TYPE = 'timeseries';
export const UNKNOWN_TYPE = 'unknown';

const TYPES_TO_LABELS: Dictionary<string> = {
	[INTEGER_TYPE]: 'Integer',
	[REAL_TYPE]: 'Decimal',
	[REAL_VECTOR_TYPE]: 'Vector',
	[LATITUDE_TYPE]: 'Latitude',
	[LONGITUDE_TYPE]: 'Longitude',
	[TEXT_TYPE]: 'Text',
	[CATEGORICAL_TYPE]: 'Categorical',
	[ORDINAL_TYPE]: 'Ordinal',
	[ADDRESS_TYPE]: 'Address',
	[CITY_TYPE]: 'City',
	[STATE_TYPE]: 'State/Province',
	[COUNTRY_TYPE]: 'Country',
	[EMAIL_TYPE]: 'Email',
	[PHONE_TYPE]: 'Phone Number',
	[POSTAL_CODE_TYPE]: 'Postal Code',
	[COUNTRY_CODE_TYPE]: 'Country Code',
	[URI_TYPE]: 'URI',
	[DATE_TIME_TYPE]: 'Date/Time',
	[BOOL_TYPE]: 'Boolean',
	[IMAGE_TYPE]: 'Image',
	[TIMESERIES_TYPE]: 'Timeseries',
	[UNKNOWN_TYPE]: 'Unknown'
};

const LABELS_TO_TYPES = _.invert(TYPES_TO_LABELS);

const INTEGER_TYPES = [
	INTEGER_TYPE
];

const COMPUTED_VAR_PREFIXES = [
	FEATURE_PREFIX,
	CLUSTER_PREFIX,
	GEOCODED_LAT_PREFIX,
	GEOCODED_LON_PREFIX,
];

const FLOATING_POINT_TYPES = [
	REAL_TYPE,
	REAL_VECTOR_TYPE,
	LATITUDE_TYPE,
	LONGITUDE_TYPE
];

const FEATURE_TYPES = [
	IMAGE_TYPE
];

const CLUSTER_TYPES = [
	TIMESERIES_TYPE
];

const NUMERIC_TYPES = INTEGER_TYPES.concat(FLOATING_POINT_TYPES);

const TEXT_TYPES = [
	TEXT_TYPE,
	IMAGE_TYPE,
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
	UNKNOWN_TYPE
];

const LOCATION_TYPES = [
	ADDRESS_TYPE,
	CITY_TYPE,
	STATE_TYPE,
	COUNTRY_TYPE,
	COUNTRY_CODE_TYPE,
	POSTAL_CODE_TYPE
];

const TIME_TYPES = [
	DATE_TIME_TYPE,
	DATE_TIME_LOWER_TYPE,
	TIMESTAMP_TYPE
];

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
	UNKNOWN_TYPE
];

const BOOL_SUGGESTIONS = [
	TEXT_TYPE,
	CATEGORICAL_TYPE,
	BOOL_TYPE,
	INTEGER_TYPE,
	UNKNOWN_TYPE
];

const EMAIL_SUGGESTIONS = [
	TEXT_TYPE,
	EMAIL_TYPE,
	UNKNOWN_TYPE
];

const URI_SUGGESTIONS = [
	TEXT_TYPE,
	URI_TYPE,
	UNKNOWN_TYPE
];

const TIME_SUGGESTIONS = [
	DATE_TIME_TYPE,
	TEXT_TYPE,
	CATEGORICAL_TYPE,
	UNKNOWN_TYPE
];

const PHONE_SUGGESTIONS = [
	TEXT_TYPE,
	INTEGER_TYPE,
	PHONE_TYPE,
	UNKNOWN_TYPE
];

const TEXT_SUGGESTIONS = [
	TEXT_TYPE,
	CATEGORICAL_TYPE,
	ORDINAL_TYPE,
	ADDRESS_TYPE,
	CITY_TYPE,
	STATE_TYPE,
	COUNTRY_TYPE,
	POSTAL_CODE_TYPE,
	DATE_TIME_TYPE,
	IMAGE_TYPE,
	UNKNOWN_TYPE
];

const INTEGER_SUGGESTIONS = [
	INTEGER_TYPE,
	REAL_TYPE,
	LATITUDE_TYPE,
	LONGITUDE_TYPE,
	CATEGORICAL_TYPE,
	ORDINAL_TYPE,
	TIMESERIES_TYPE,
	UNKNOWN_TYPE
];

const DECIMAL_SUGGESTIONS = [
	INTEGER_TYPE,
	REAL_TYPE,
	REAL_VECTOR_TYPE,
	LATITUDE_TYPE,
	LONGITUDE_TYPE,
	TIMESERIES_TYPE,
	UNKNOWN_TYPE
];

const IMAGE_SUGGESTIONS = [
	IMAGE_TYPE,
	TEXT_TYPE,
	CATEGORICAL_TYPE
];

export const BASIC_SUGGESTIONS = [
	INTEGER_TYPE,
	REAL_TYPE,
	CATEGORICAL_TYPE,
	ORDINAL_TYPE,
	TEXT_TYPE,
	IMAGE_TYPE,
	DATE_TIME_TYPE,
	TIMESERIES_TYPE,
	UNKNOWN_TYPE
];

const EQUIV_TYPES = {
	[INTEGER_TYPE]: [ INTEGER_TYPE ],
	[REAL_TYPE]: [ REAL_TYPE ],
	[REAL_VECTOR_TYPE]: [ REAL_VECTOR_TYPE ],
	[LATITUDE_TYPE]: [ LATITUDE_TYPE ],
	[LONGITUDE_TYPE]: [ LONGITUDE_TYPE ],
	[TEXT_TYPE]:  [ TEXT_TYPE ],
	[CATEGORICAL_TYPE]: [ CATEGORICAL_TYPE ],
	[ORDINAL_TYPE]: [ ORDINAL_TYPE ],
	[ADDRESS_TYPE]: [ ADDRESS_TYPE ],
	[CITY_TYPE]: [ CITY_TYPE ],
	[STATE_TYPE]: [ STATE_TYPE ],
	[COUNTRY_TYPE]: [ COUNTRY_TYPE ],
	[EMAIL_TYPE]: [ EMAIL_TYPE ],
	[PHONE_TYPE]: [ PHONE_TYPE ],
	[POSTAL_CODE_TYPE]: [ POSTAL_CODE_TYPE ],
	[URI_TYPE]: [ URI_TYPE ],
	[DATE_TIME_TYPE]: [ DATE_TIME_TYPE ],
	[DATE_TIME_LOWER_TYPE]: [ DATE_TIME_TYPE ],
	[BOOL_TYPE]: [ BOOL_TYPE ],
	[IMAGE_TYPE]: [ IMAGE_TYPE ],
	[TIMESERIES_TYPE]: [ TIMESERIES_TYPE ],
	[UNKNOWN_TYPE]: [ UNKNOWN_TYPE ]
};

export function isEquivalentType(a: string, b: string): boolean {
	const equiv = EQUIV_TYPES[a];
	if (!equiv) {
		console.warn(`Unable to find equivalent types for type '${a}', type unrecognized`);
		return false;
	}
	const matches = equiv.filter((type: string) => {
		return type === b;
	});
	return matches.length > 0;
}

export function normalizedEquivalentType (rawType: string): string {
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
	if (!colType || colType === '' || colType === D3M_INDEX_FIELD) {
		if (_.isNumber(colValue)) {
			return _.isInteger(colValue) ? colValue : colValue.toFixed(4);
		}
		return colValue;
	}

	// If the schema type is numeric and the value is a number stored as a string,
	// parse it and format again.
	if (isNumericType(colType) &&
		!_.isNumber(colValue) && !_.isNaN(Number.parseFloat(colValue))) {
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

	if (colValue === '') {
		return colValue;
	}

	// We've got a floating point value - set precision based on
	// type.
	switch (colType) {
		case LONGITUDE_TYPE:
		case LATITUDE_TYPE:
			return colValue.toFixed(6);
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

export function isLocationType(type: string): boolean {
	return LOCATION_TYPES.indexOf(type) !== -1;
}

export function isTextSimpleType(type: string): boolean {
	return TEXT_SIMPLE_TYPES.indexOf(type) !== -1;
}

export function isFeatureType(type: string): boolean {
	return FEATURE_TYPES.indexOf(type) !== -1;
}

export function addFeaturePrefix(varName: string): string {
	return `${FEATURE_PREFIX}${varName}`;
}

export function hasComputedVarPrefix(varName: string): boolean {
	return Boolean(COMPUTED_VAR_PREFIXES.find(prefix => varName.indexOf(prefix) === 0));
}

export function removeFeaturePrefix(varName: string): string {
	return varName.replace(FEATURE_PREFIX, '');
}

export function isClusterType(type: string): boolean {
	return CLUSTER_TYPES.indexOf(type) !== -1;
}

export function addClusterPrefix(varName: string): string {
	return `${CLUSTER_PREFIX}${varName}`;
}

export function removeClusterPrefix(varName: string): string {
	return varName.replace(CLUSTER_PREFIX, '');
}

export function isJoinable(type: string, otherType: string): boolean {
	const isSameType = type === otherType;
	const isBothNumericType = isNumericType(type) && isNumericType(otherType);
	return isSameType || isBothNumericType;
}

export function addTypeSuggestions(type: string, values: any[]): string[] {
	let suggestions = guessTypeByValue(values);
	if (!suggestions || suggestions.length === 0) {
		suggestions = BASIC_SUGGESTIONS;
	}
	return suggestions;
}

export function guessTypeByType(type: string): string[] {
	if (isNumericType(type)) {
		return isFloatingPointType(type) ? DECIMAL_SUGGESTIONS : INTEGER_SUGGESTIONS;
	}
	return TEXT_SUGGESTIONS;
}

function combineTypeWithUnion(types: string[][]): string[] {
	let res = [];
	types.forEach(ts => {
		res = res.concat(ts);
	});
	return _.uniq(res);
}

function combineTypeWithIntersection(types: string[][]): string[] {
	const counts = {};
	types.forEach(ts => {
		ts.forEach(type => {
			if (counts[type] === undefined) {
				counts[type] = 0;
			}
			counts[type]++;
		});
	});
	const res = [];
	_.forIn(counts, (val, key) => {
		if (val === types.length) {
			res.push(key);
		}
	});
	return res;
}

function combineSampledTypes(types: string[][]): string[] {
	const USE_INTERSECTION = true;
	if (USE_INTERSECTION) {
		return combineTypeWithIntersection(types);
	}
	return combineTypeWithUnion(types);
}

function checkBooleanValues(values: any) {
	if (values.length === 2) {
		const boolTest = values.filter(val => BOOL_REGEX.test(val));
		if (boolTest[0] && boolTest[1]) {
			return true;
		}
	}
	return false;
}

export function guessTypeByValue(value: any): string[] {
	if (value === undefined) {
		return TEXT_SUGGESTIONS;
	}
	if (_.isArray(value)) {
		const types = [];
		const uniqValues = _.uniq(value);
		checkBooleanValues(uniqValues)
			? types.push(BOOL_SUGGESTIONS)
			: value.forEach(val => types.push(guessTypeByValue(val)));
		return combineSampledTypes(types);
	}
	if (_.isNumber(value) || !_.isNaN(_.toNumber(value))) {
		const num = _.toNumber(value);
		return _.isInteger(num) ? INTEGER_SUGGESTIONS : DECIMAL_SUGGESTIONS;
	}
	if (EMAIL_REGEX.test(value)) {
		return EMAIL_SUGGESTIONS;
	}
	if (URI_REGEX.test(value)) {
		return URI_SUGGESTIONS;
	}
	if (PHONE_REGEX.test(value)) {
		return PHONE_SUGGESTIONS;
	}
	if (IMAGE_REGEX.test(value)) {
		return IMAGE_SUGGESTIONS;
	}
	if (Date.parse(value)) {
		return TIME_SUGGESTIONS;
	}
	return TEXT_SUGGESTIONS;
}


/**
 * Returns a UI-ready label for a given schema type.
 */
export function getLabelFromType(schemaType: string) {
	if (_.has(TYPES_TO_LABELS, schemaType)) {
		return TYPES_TO_LABELS[schemaType];
	}
	console.warn(`No label exists for type \`${schemaType}\` - using type as default label`);
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
