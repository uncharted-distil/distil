import _ from 'lodash';

const EMAIL_REGEX = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
const URI_REGEX = /^(?:(?:(?:https?|ftp):)?\/\/)(?:\S+(?::\S*)?@)?(?:(?!(?:10|127)(?:\.\d{1,3}){3})(?!(?:169\.254|192\.168)(?:\.\d{1,3}){2})(?!172\.(?:1[6-9]|2\d|3[0-1])(?:\.\d{1,3}){2})(?:[1-9]\d?|1\d\d|2[01]\d|22[0-3])(?:\.(?:1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.(?:[1-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(?:(?:[a-z\u00a1-\uffff0-9]-*)*[a-z\u00a1-\uffff0-9]+)(?:\.(?:[a-z\u00a1-\uffff0-9]-*)*[a-z\u00a1-\uffff0-9]+)*(?:\.(?:[a-z\u00a1-\uffff]{2,})))(?::\d{2,5})?(?:[/?#]\S*)?$/i;
const BOOL_REGEX = /^(0|1|true|false|t|f)$/i;
const PHONE_REGEX = /^(\+\d{1,2}\s)?\(?\d{3}\)?[\s.-]\d{3}[\s.-]\d{4}$/

const INTEGER_TYPES = [
	'integer',
];

const FLOATING_POINT_TYPES = [
	'decimal',
	'latitude',
	'longitude'
];

const NUMERIC_TYPES = INTEGER_TYPES.concat(FLOATING_POINT_TYPES);

const TEXT_TYPES = [
	'text',
	'categorical',
	'ordinal',
	'address',
	'city',
	'state',
	'country',
	'email',
	'phone',
	'postal_code',
	'uri',
	'keyword',
	'dateTime',
	'boolean'
];

const BOOL_SUGGESTIONS = [
	'text',
	'categorical',
	'boolean',
	'integer',
	'keyword'
];

const EMAIL_SUGGESTIONS = [
	'text',
	'email'
];

const URI_SUGGESTIONS = [
	'text',
	'uri'
];

const PHONE_SUGGESTIONS= [
	'text',
	'integer',
	'phone'
];

const TEXT_SUGGESTIONS = [
	'text',
	'categorical',
	'ordinal',
	'address',
	'city',
	'state',
	'country',
	'postal_code',
	'keyword',
	'dateTime'
];

const INTEGER_SUGGESTIONS = [
	'integer',
	'decimal',
	'latitude',
	'longitude',
	'categorical',
	'ordinal'
];

const DECIMAL_SUGGESTIONS = [
	'integer',
	'decimal',
	'latitude',
	'longitude'
];

const BASIC_SUGGESTIONS = [
	'integer',
	'decimal',
	'categorical',
	'ordinal',
	'text'
];

export function formatValue(colValue: any, colType: string): any {
	if (!colType || colType === '') {
		if (_.isNumber(colValue)) {
			return _.isInteger(colValue) ? colValue : colValue.toFixed(4);
		}
		return colValue;
	}
	if (isTextType(colType)) {
		return colValue;
	}
	if (_.isInteger(colValue)) {
		return colValue;
	}
	switch (colType) {
		case 'longitude':
		case 'latitude':
			return colValue.toFixed(6);
	}
	return colValue.toFixed(4);
}

export function isNumericType(type: string): boolean {
	return NUMERIC_TYPES.indexOf(type) !== -1;
}

export function isFloatingPointType(type: string): boolean {
	return FLOATING_POINT_TYPES.indexOf(type) !== -1;
}

export function isTextType(type: string): boolean {
	return TEXT_TYPES.indexOf(type) !== -1;
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
	if (value.match(EMAIL_REGEX)) {
		return EMAIL_SUGGESTIONS;
	}
	if (value.match(URI_REGEX)) {
		return URI_SUGGESTIONS;
	}
	if (value.match(PHONE_REGEX)) {
		return PHONE_SUGGESTIONS;
	}
	return TEXT_SUGGESTIONS;
}
