import _ from 'lodash';
import Vue from 'vue';
import { Dictionary } from './dict'
import { getters as routeGetters } from '../store/route/module';
import { overlayRouteEntry } from './routes';

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
	min?: number;
	max?: number;
	categories?: string[];
}

export interface FilterParams {
	filters: Filter[];
	variables: string[];
	size?: number;
}

/**
 * Decodes the filters from the route string into an array.
 *
 * @param {string} filters - The filters from the route query string.
 *
 * @returns {Filter[]} The decoded filter object.
 */
export function decodeFilters(filters: string): FilterParams {
	if (_.isEmpty(filters)) {
		return {
			filters: [],
			variables: []
		};
	}
	return JSON.parse(atob(filters)) as FilterParams;
}

/**
 * Decodes the filters from the route string into a dictionary.
 *
 * @param {string} filters - The filters from the route query string.
 *
 * @returns {Dictionary<Filter>} The decoded filter object.
 */
export function decodeFiltersDictionary(filters: string): Dictionary<Filter> {
	const params = decodeFilters(filters);
	const map = {};
	params.filters.forEach(filter => {
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
export function encodeFilters(filters: FilterParams): string {
	if (_.isEmpty(filters)) {
		return null;
	}
	return btoa(JSON.stringify(filters));
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
	let index = _.findIndex(decoded.filters, existing => {
		return existing.name === filter.name;
	})

	if (index === -1) {
		// add filter
		decoded.filters.push(filter);
	} else {
		// replace existing
		decoded.filters[index] = filter;
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
