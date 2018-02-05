import _ from 'lodash';
import Vue from 'vue';
import { Dictionary } from './dict'
import { getters as routeGetters } from '../store/route/module';
import { overlayRouteEntry } from './routes';

/**
 * Empty filter, omitting no documents.
 * @constant {string}
 */
export const EMPTY_FILTER = 'empty';

/**
 * Categorical filter, omitting documents that do not contain the provided
 * categories in the variable.
 * @constant {string}
 */
export const CATEGORICAL_FILTER = 'categorical';

/**
 * Numerical filter, omitting documents that do not fall within the provided
 * variable range.
 * @constant {string}
 */
export const NUMERICAL_FILTER = 'numerical';

export interface Filter {
	name: string;
	type: string;
	enabled: boolean;
	min?: number;
	max?: number;
	categories?: string[];
}

export interface FilterParams {
	filters: Filter[];
	size?: number;
}

/**
 * Decodes the filters from the route string into an array.
 *
 * @param {string} filters - The filters from the route query string.
 *
 * @returns {Filter[]} The decoded filter object.
 */
export function decodeFilters(filters: string): Filter[] {
	if (_.isEmpty(filters)) {
		return [];
	}
	return JSON.parse(atob(filters)) as Filter[];
}

/**
 * Decodes the filters from the route string into a dictionary.
 *
 * @param {string} filters - The filters from the route query string.
 *
 * @returns {Dictionary<Filter>} The decoded filter object.
 */
export function decodeFiltersDictionary(filters: string): Dictionary<Filter> {
	const arr = decodeFilters(filters);
	const map = {};
	arr.forEach(filter => {
		map[filter.name] = filter;
	});
	return map;
}


/**
 * Encodes the map of filter objects into a map of route query strings.
 *
 * @param {Filter[]} filters - The filter objects.
 *
 * @returns {string} The encoded route query strings.
 */
export function encodeFilters(filters: Filter[]): string {
	if (_.isEmpty(filters)) {
		return null;
	}
	return btoa(JSON.stringify(filters));
}

/**
 * Encodes the filter object into a query param string for an HTTP request.
 *
 * @param {Filter} filter - The filter object.
 *
 * @returns {string} The HTTP query param strings.
 */
export function encodeQueryParam(filter: Filter): string {
	if (isDisabled(filter)) {
		return `${encodeURIComponent(filter.name)}`;
	}
	switch (getFilterType(filter)) {
		case NUMERICAL_FILTER:
			return `${encodeURIComponent(filter.name)}=${NUMERICAL_FILTER},${filter.min},${filter.max}`;

		case CATEGORICAL_FILTER:
			return `${encodeURIComponent(filter.name)}=${CATEGORICAL_FILTER},${filter.categories.join(',')}`;
	}
	return null;
}

/**
 * Encodes the filter objects into a single query param string for an HTTP
 * request.
 *
 * @param {Filter[]} filters - The filter objects.
 *
 * @returns {string} The HTTP query param strings.
 */
export function encodeQueryParams(filters: Filter[]): string {
	const params: string[] = [];
	_.forEach(filters, filter => {
		const param = encodeQueryParam(filter);
		if (param !== null) {
			params.push(param);
		}
	});
	return params.length > 0 ? `?${params.join('&')}` : '';
}

export function overlayFilter(dst: Filter, src: Filter): Filter {
	// only override empty filters with typed filters
	if (dst.type === EMPTY_FILTER && src.type !== EMPTY_FILTER) {
		dst.type = src.type;
	}
	dst.enabled = _.defaultTo(src.enabled, dst.enabled);
	dst.min = _.defaultTo(src.min, dst.min);
	dst.max = _.defaultTo(src.max, dst.max);
	dst.categories = _.defaultTo(src.categories, dst.categories);
	return dst;
}

/**
 * Updates the route with the provided route filter key and value. The function
 * will add, modify, or remove the filter as necessary.
 *
 * @param {string} filters - The existing route filter string.
 * @param {Filter} filter - The filter.
 *
 * @returns {string} The updated route filter strings.
 */
export function updateFilter(filters: string, filter: Filter): string {
	// decode the provided filters
	const decoded = decodeFilters(filters);
	// get or create the filter
	let index = _.findIndex(decoded, existing => {
		return existing.name === filter.name;
	})

	let target = null;
	if (index === -1) {
		// add filter
		target = filter;
		decoded.push(filter);
		index = decoded.length - 1;
	} else {
		// overlay onto existing
		target = decoded[index];
		overlayFilter(target, filter);
	}

	// empty enabled filter is default, remove it
	if (getFilterType(target) === EMPTY_FILTER && isEnabled(target)) {
		decoded.splice(index, 1);
	}

	// encode the filters back into a url string
	return encodeFilters(decoded);
}

export function updateFilterRoute(component: Vue, filter: Filter) {
	// retrieve the filters from the route
	const filters = routeGetters.getRouteFilters(component.$store);
	// merge the updated filters back into the route query params
	const updated = updateFilter(filters, filter);
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		filters: updated
	});
	component.$router.push(entry);
}

interface NumericalFilterParams {
	from: {
		label: string[]
	},
	to: {
		label: string[]
	}
};

type CategoricalFilterParams = string[];

export function createNumericalFilter(key: string, value: NumericalFilterParams): Filter {
	return {
		name: key,
		type: NUMERICAL_FILTER,
		enabled: true,
		min: parseFloat(value.from.label[0]),
		max: parseFloat(value.to.label[0])
	};
}

export function createCategoricalFilter(key: string, value: CategoricalFilterParams): Filter {
	return {
		name: key,
		type: CATEGORICAL_FILTER,
		enabled: true,
		categories: value
	};
}

/**
 * Returns the filter type symbol.
 *
 * @param {Object} filter - The filter object or string.
 *
 * @returns {string} The filter type.
 */
export function getFilterType(filter: Filter): string {
	return (filter) ? filter.type : EMPTY_FILTER;
}

/**
 * Returns whether or not the filter is enabled.
 *
 * @param {Filter} filter - The filter object.
 *
 * @returns {bool} Whether or not the filter is enabled.
 */
export function isEnabled(filter: Filter): boolean {
	if (filter) {
		return filter.enabled;
	}
	return true;
}

/**
 * Returns whether or not the filter is disabled.
 *
 * @param {Filter} filter - The filter object.
 *
 * @returns {bool} Whether or not the filter is disabled.
 */
export function isDisabled(filter: Filter): boolean {
	return !isEnabled(filter);
}
