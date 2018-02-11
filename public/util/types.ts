import _ from 'lodash';

const LOW_PROBABILITY = 0.33;
const MED_PROBABILITY = 0.66;
const DEFAULT_PROBABILITY = 0.5;

const INTEGER_TYPES = [
	'integer',
];

const FLOATING_POINT_TYPES = [
	'float',
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

const INTEGER_SUGGESTIONS = [
	'integer',
	'float',
	'latitude',
	'longitude',
	'categorical',
	'ordinal'
];

const FLOAT_SUGGESTIONS = [
	'integer',
	'float',
	'latitude',
	'longitude'
];

const TEXT_SUGGESTIONS = [
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

export function formatValue(colValue: any, colType: string): any {
	// If there is no assigned schema, fix precision for a number, pass through otherwise.
	if (!colType || colType === '') {
		return _.isNumber(colValue) ? colValue.toFixed(4) : colValue;
	}

	// If the schema type is numeric and the value is a number stored as a string,
	// parse it and format again.
	if (isNumericType(colType) && 
		!_.isNumber(colValue) && !_.isNaN(Number.parseFloat(colValue))) {
		return formatValue(Number.parseFloat(colValue), colType);
	}

	// If the schema type is an integer, round.
	if (isIntegerType(colType)) {
		return Math.round(colValue);
	}

	// If the schema type is text or not float, pass through.
	if (isTextType(colType) || !isFloatingPointType(colType)) {
		return colValue;
	}

	// We've got a floating point value - set precision based on
	// type.
	switch (colType) {
		case 'float':
			return colValue.toFixed(4);
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

export function isIntegerType(type: string): boolean {
	return INTEGER_TYPES.indexOf(type) !== -1;
}

export function isTextType(type: string): boolean {
	return TEXT_TYPES.indexOf(type) !== -1;
}

export function probabilityCategoryText(probability: number): string {
	if (probability < LOW_PROBABILITY) {
		return 'Low';
	}
	if (probability < MED_PROBABILITY) {
		return 'Med';
	}
	return 'High';
}

export function probabilityCategoryClass(probability: number): string {
	if (probability < LOW_PROBABILITY) {
		return 'text-danger';
	}
	if (probability < MED_PROBABILITY) {
		return 'text-warning';
	}
	return 'text-success';
}

export function addSuggestions(current: string[], suggestions: string[], probability: number): string[] {
	suggestions.forEach((suggestion: string) => {
		// check if already exists
		const index = _.findIndex(current, (s: string) => {
			return s === suggestion;
		});
		if (index === -1) {
			// add
			current.push(suggestion);
		}
	});
	return current;
}

export function addMissingSuggestions(type: string): string[] {
	// copy current suggestions by value
	const current = [];
	if (isNumericType(type)) {
		if (isFloatingPointType(type)) {
			// float
			return addSuggestions(current, FLOAT_SUGGESTIONS, DEFAULT_PROBABILITY);
		}
		// integer
		return addSuggestions(current, INTEGER_SUGGESTIONS, DEFAULT_PROBABILITY);
	}
	// text
	return addSuggestions(current, TEXT_SUGGESTIONS, DEFAULT_PROBABILITY);
}
