import _ from 'lodash';

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

/**
 * Decodes the map of filters from the route into objects.
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
 * Encodes the map of filter objects into a map of route query strings.
 *
 * @param {Filter[]} filters - The filter objects.
 *
 * @returns {string} The encoded route query strings.
 */
export function encodeFilters(filters: Filter[]): string {
	if (_.isEmpty(filters)) {
		return undefined;
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
	if (index === -1) {
		// does not exist yet
		decoded.push(filter)
		index = decoded.length - 1;
	} else {
		// use existing
		filter = decoded[index];
	}
	// overlay the new filter values
	_.forIn(filter, (v, k) => {
		filter[k] = v;
	});
	// set the type field
	filter.type = getFilterType(filter);
	// empty enabled filter is default, remove it
	if (getFilterType(filter) === EMPTY_FILTER && isEnabled(filter)) {
		decoded.splice(index, 1);
	}
	// encode the filters back into a url string
	return encodeFilters(decoded);
}

/**
 * Returns the filter type symbol.
 *
 * @param {Object} filter - The filter object or string.
 *
 * @returns {string} The filter type.
 */
export function getFilterType(filter: Filter): string {
	if (filter) {
		if (_.has(filter, 'categories')) {
			return CATEGORICAL_FILTER;
		}
		if (_.has(filter, 'min') && _.has(filter, 'max')) {
			return NUMERICAL_FILTER;
		}
	}
	return EMPTY_FILTER;
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
