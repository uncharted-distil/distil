import _ from 'lodash';
import { SuggestedType } from '../store/data/index';

const LOW_PROBABILITY = 0.33;
const MED_PROBABILITY = 0.66;
const DEFAULT_PROBABILITY = 0.5;

const INTEGER_TYPES = [
	'integer',
	'long'
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
	'long',
	'float',
	'latitude',
	'longitude',
	'categorical',
	'ordinal'
];

const FLOAT_SUGGESTIONS = [
	'integer',
	'long',
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

export function isNumericType(type: string): boolean {
	return NUMERIC_TYPES.indexOf(type) !== -1;
}

export function isFloatingPointType(type: string): boolean {
	return FLOATING_POINT_TYPES.indexOf(type) !== -1;
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

export function addSuggestions(current: SuggestedType[], suggestions: string[], probability: number): SuggestedType[] {
	suggestions.forEach((suggestion: string) => {
		// check if already exists
		const index = _.findIndex(current, (s: SuggestedType) => {
			return s.type === suggestion;
		});
		if (index === -1) {
			// add
			current.push({
				type: suggestion,
				probability: probability
			})
		}
	});
	return current;
}

export function addMissingSuggestions(suggested: SuggestedType[], type: string): SuggestedType[] {
	// copy current suggestions by value
	const current = suggested ? suggested.slice() : [];
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
