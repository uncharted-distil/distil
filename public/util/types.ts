import _ from 'lodash';
import { SuggestedType } from '../store/data/index';

const LOW_PROBABILITY = 0.33;
const MED_PROBABILITY = 0.66;

const NUMERIC_TYPES = [
	"integer",
	"index",
	"long",
	"float",
	"double",
	"latitude",
	"longitude"
];

const TEXT_TYPES = [
	"text",
	"categorical",
	"ordinal",
	"address",
	"city",
	"state",
	"country",
	"email",
	"phone",
	"postal_code",
	"uri",
	"keyword",
	"dateTime",
	"boolean"
];

export function isNumericType(type: string): boolean {
	return NUMERIC_TYPES.indexOf(type) !== -1;
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

export function addMissingSuggestions(suggested: SuggestedType[], type: string) {
	const all = suggested.slice();
	if (isNumericType(type)) {
		NUMERIC_TYPES.forEach((nt: string) => {
			const exists = _.findIndex(suggested, (s: SuggestedType) => {
				return s.type === nt;
			}) !== -1;
			if (!exists) {
				// add
				all.push({
					type: nt,
					probability: 0.5
				})
			}
		});
	} else {
		TEXT_TYPES.forEach((tt: string) => {
			const exists = _.findIndex(suggested, (s: SuggestedType) => {
				return s.type === tt;
			}) !== -1;
			if (!exists) {
				// add
				all.push({
					type: tt,
					probability: 0.5
				})
			}
		});
	}
	return all;
}
