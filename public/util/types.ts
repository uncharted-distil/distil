import _ from 'lodash';
import { store } from '../store/storeProvider';
import { Dictionary } from './dict';
import { getters as datasetGetters } from '../store/dataset/module';
import { D3M_INDEX_FIELD } from '../store/dataset/index';
import { FEATURE_FILTER, CATEGORICAL_FILTER, NUMERICAL_FILTER } from '../util/filters';

const EMAIL_REGEX = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
const URI_REGEX = /^(?:(?:(?:https?|ftp):)?\/\/)(?:\S+(?::\S*)?@)?(?:(?!(?:10|127)(?:\.\d{1,3}){3})(?!(?:169\.254|192\.168)(?:\.\d{1,3}){2})(?!172\.(?:1[6-9]|2\d|3[0-1])(?:\.\d{1,3}){2})(?:[1-9]\d?|1\d\d|2[01]\d|22[0-3])(?:\.(?:1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.(?:[1-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(?:(?:[a-z\u00a1-\uffff0-9]-*)*[a-z\u00a1-\uffff0-9]+)(?:\.(?:[a-z\u00a1-\uffff0-9]-*)*[a-z\u00a1-\uffff0-9]+)*(?:\.(?:[a-z\u00a1-\uffff]{2,})))(?::\d{2,5})?(?:[/?#]\S*)?$/i;
const BOOL_REGEX = /^(0|1|true|false|t|f)$/i;
const PHONE_REGEX = /^(\+\d{1,2}\s)?\(?\d{3}\)?[\s.-]\d{3}[\s.-]\d{4}$/
const IMAGE_REGEX = /\.(gif|jpg|jpeg|png|tif|tiff|bmp)$/i;
const META_PREFIX = '_feature_';

const TYPES_TO_LABELS: Dictionary<string> = {
	integer: 'Integer',
	int4: 'Integer',
	float: 'Decimal',
	float8: 'Decimal',
	latitude: 'Latitude',
	longitude: 'Longitude',
	text: 'Text',
	categorical: 'Categorical',
	ordinal: 'Ordinal',
	address: 'Address',
	city: 'City',
	state: 'State/Province',
	country: 'Country',
	email: 'Email',
	phone: 'Phone Number',
	postal_code: 'Postal Code',
	uri: 'URI',
	keyword: 'Keyword',
	dateTime: 'Date/Time',
	boolean: 'Boolean',
	image: 'Image',
	timeseries: 'Time Series',
	unknown: 'Unknown'
};

const LABELS_TO_TYPES = _.invert(TYPES_TO_LABELS);

const INTEGER_TYPES = [
	'integer',
	'int4'
];

const FLOATING_POINT_TYPES = [
	'float',
	'float8',
	'latitude',
	'longitude'
];

const META_TYPES = [
	'image',
	'timeseries'
];

const NUMERIC_TYPES = INTEGER_TYPES.concat(FLOATING_POINT_TYPES);

const TEXT_TYPES = [
	'text',
	'image',
	'timeseries',
	'categorical',
	'ordinal',
	'address',
	'city',
	'state',
	'country',
	'country_code',
	'email',
	'phone',
	'postal_code',
	'uri',
	'keyword',
	'dateTime',
	'boolean',
	'unknown'
];

const BOOL_SUGGESTIONS = [
	'text',
	'categorical',
	'boolean',
	'integer',
	'keyword',
	'unknown'
];

const EMAIL_SUGGESTIONS = [
	'text',
	'email',
	'unknown'
];

const URI_SUGGESTIONS = [
	'text',
	'uri',
	'unknown'
];

const PHONE_SUGGESTIONS = [
	'text',
	'integer',
	'phone',
	'unknown'
];

const TEXT_SUGGESTIONS = [
	'text',
	'categorical',
	'ordinal',
	'integer',
	'address',
	'city',
	'state',
	'country',
	'postal_code',
	'keyword',
	'dateTime',
	'image',
	'unknown'
];

const INTEGER_SUGGESTIONS = [
	'integer',
	'float',
	'latitude',
	'longitude',
	'categorical',
	'ordinal',
	'unknown'
];

const DECIMAL_SUGGESTIONS = [
	'integer',
	'float',
	'latitude',
	'longitude',
	'unknown'
];

const IMAGE_SUGGESTIONS = [
	'image',
	'text',
	'categorical'
];

const BASIC_SUGGESTIONS = [
	'integer',
	'float',
	'categorical',
	'ordinal',
	'text',
	'image',
	'timeseries',
	'unknown'
];

export function getVarType(varname: string): string {
	return datasetGetters.getVariableTypesMap(store())[varname];
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
		case 'longitude':
		case 'latitude':
			return colValue.toFixed(6);
	}
	return colValue.toFixed(4);
}

export function getFilterType(type: string): string {
	if (isMetaType(type)) {
		return FEATURE_FILTER;
	} else if (isNumericType(type)) {
		return NUMERICAL_FILTER;
	}
	return CATEGORICAL_FILTER;
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

export function isMetaType(type: string): boolean {
	return META_TYPES.indexOf(type) !== -1;
}

export function addMetaPrefix(varName: string): string {
	return `${META_PREFIX}${varName}`;
}

export function removeMetaPrefix(varName: string): string {
	return varName.replace(META_PREFIX, '');
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

export function guessTypeByValue(value: any): string[] {
	if (value === undefined) {
		return TEXT_SUGGESTIONS;
	}
	if (_.isArray(value)) {
		let types = [];
		value.forEach(val => {
			types = types.concat(guessTypeByValue(val));
		});
		return _.uniq(types);
	}
	if (BOOL_REGEX.test(value)) {
		return BOOL_SUGGESTIONS;
	}
	if (_.isNumber(value) || !_.isNaN(_.toNumber(value))) {
		const num = _.toNumber(value);
		return _.isInteger(num) ? INTEGER_SUGGESTIONS : DECIMAL_SUGGESTIONS
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
	};
	console.warn(`No type exists for label \`${label}\``);
	return label;
}
