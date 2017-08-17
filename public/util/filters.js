import _ from 'lodash';

/**
 * Empty filter, omitting no documents.
 * @constant {Symbol}
 */
export const EMPTY_FILTER = Symbol('empty');
export const EMPTY_FILTER_ID = 'empty';

/**
 * Categorical filter, omitting documents that do not contain the provided
 * categories in the variable.
 * @constant {Symbol}
 */
export const CATEGORICAL_FILTER = Symbol('categorical');
export const CATEGORICAL_FILTER_ID = 'categorical';

/**
 * Numerical filter, omitting documents that do not fall within the provided
 * variable range.
 * @constant {Symbol}
 */

export const NUMERICAL_FILTER = Symbol('numerical');
export const NUMERICAL_FILTER_ID = 'numerical';

/**
 * Decodes the filter from the route into an object:
 * Ex:
 *     input: ("varName", "1,numerical,1,9")
 *
 *     output: {
 *         name: "VarName",
 *         enabled: true,
 *         type: "numerical",
 *         min: 1,
 *         max: 9
 *     }
 *
 * @param {string} filterName - The name of the filter
 * @param {Object} filter - The filter string from the route
 *
 * @returns {Object} The decoded filter object.
 */
export function decodeFilter(filterName, filter) {
	if (!filter) {
		return null;
	}
	const values = filter.split(',');
	if (values.length >= 2) {
		const enabled = values[0] === '1';
		const type = values[1];
		switch (type) {
			case NUMERICAL_FILTER_ID:
				return {
					name: filterName,
					enabled: enabled,
					min: _.toNumber(values[2]),
					max: _.toNumber(values[3])
				};
			case CATEGORICAL_FILTER_ID:
				return {
					name: filterName,
					enabled: enabled,
					categories: values.slice(2)
				};
			case EMPTY_FILTER_ID:
				return {
					name: filterName,
					enabled: enabled
				};
			default:
				console.warn(`invalid filter type of ${type}`);
				return null;
		}
	}
	// enabled empty filter
	return null;
}

/**
 * Decodes the map of filters from the route into objects.
 *
 * @param {Object} filters - The filters from the route query string.
 *
 * @returns {Object} The decoded filter object.
 */
export function decodeFilters(filters) {
	const results = {};
	_.forEach(filters, (filter, filterName) => {
		const decoded = decodeFilter(filterName, filter);
		if (decoded) {
			results[decoded.name] = decoded;
		}
	});
	return results;
}

/**
 * Encodes the filter object into the filter route query string:
 *
 *     input: {
 *         name: "VarName",
 *         enabled: true,
 *         type: "numerical",
 *         min: 1,
 *         max: 9
 *     }
 *     ouput: "VarName=1,numerical,1,9"
 *
 * @param {Object} filter - The filter object.
 *
 * @returns {string} The encoded filter route string.
 */
export function encodeFilter(filter) {
	if (!filter) {
		return null;
	}
	const enabled = filter.enabled ? '1' : '0';
	switch (getFilterType(filter)) {
		case EMPTY_FILTER:
			return `${enabled},${EMPTY_FILTER_ID}`;

		case NUMERICAL_FILTER:
			return `${enabled},${NUMERICAL_FILTER_ID},${filter.min},${filter.max}`;

		case CATEGORICAL_FILTER:
			return `${enabled},${CATEGORICAL_FILTER_ID},${filter.categories.join(',')}`;
	}
	return null;
}

/**
 * Encodes the map of filter objects into a map of route query strings.
 *
 * @param {Object} filters - The filter objects.
 *
 * @returns {Object} The encoded route query strings.
 */
export function encodeFilters(filters) {
	const results = {};
	_.forEach(filters, (filter, name) => {
		const encoded = encodeFilter(filter);
		if (encoded !== null) {
			results[name] = encoded;
		}
	});
	return results;
}

/**
 * Encodes the filter object into a query param string for an HTTP request.
 *
 * @param {Object} filter - The filter object.
 *
 * @returns {Object} The HTTP query param strings.
 */
export function encodeQueryParam(filter) {
	if (isDisabled(filter)) {
		return `${encodeURIComponent(filter.name)}`;
	}
	switch (getFilterType(filter)) {
		case NUMERICAL_FILTER:
			return `${encodeURIComponent(filter.name)}=${NUMERICAL_FILTER_ID},${filter.min},${filter.max}`;

		case CATEGORICAL_FILTER:
			return `${encodeURIComponent(filter.name)}=${CATEGORICAL_FILTER_ID},${filter.categories.join(',')}`;
	}
	return null;
}

/**
 * Encodes the filter objects into a single query param string for an HTTP
 * request.
 *
 * @param {Object} filters - The filter objects.
 *
 * @returns {string} The HTTP query param strings.
 */
export function encodeQueryParams(filters) {
	const params = [];
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
 * @param {Object} filters - The route filter strings.
 * @param {string} key - The filter key.
 * @param {Object} values - The filter values.
 *
 * @returns {Object} The updated route filter strings.
 */
export function updateFilter(filters, key, values) {
	// decode the provided filters
	const decoded = decodeFilters(filters);
	// get or create the filter
	let filter = decoded[key];
	if (!filter) {
		filter = {
			name: key,
			enabled: true
		};
		decoded[key] = filter;
	}
	// add the filter values
	_.forIn(values, (v, k) => {
		filter[k] = v;
	});
	const encoded = encodeFilters(decoded);
	// empty enabled filter is default, so remove it
	if (getFilterType(filter) === EMPTY_FILTER && isEnabled(filter)) {
		encoded[key] = undefined;
	}
	return encoded;
}

/**
 * Returns the filter type symbol.
 *
 * @param {string|Object} filter - The filter object or string.
 *
 * @returns {Symbol} The filter type symbol.
 */
export function getFilterType(filter) {
	if (_.isString(filter)) {
		filter = decodeFilter(filter);
	}
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
 * @param {string|Object} filter - The filter object or string.
 *
 * @returns {bool} Whether or not the filter is enabled.
 */
export function isEnabled(filter) {
	if (_.isString(filter)) {
		// name doesn't matter in this decode context
		filter = decodeFilter('filter', filter);
	}
	if (filter) {
		return filter.enabled;
	}
	return true;
}

/**
 * Returns whether or not the filter is disabled.
 *
 * @param {string|Object} filter - The filter object or string.
 *
 * @returns {bool} Whether or not the filter is disabled.
 */
export function isDisabled(filter) {
	return !isEnabled(filter);
}
