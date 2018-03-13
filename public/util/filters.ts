import _ from 'lodash';
import Vue from 'vue';
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


/**
 * Include filter, excluding documents that do not fall within the filter.
 * @constant {string}
 */
export const INCLUDE_FILTER = 'include';

/**
 * Exclude filter, excluding documents that fall outside the filter.
 * @constant {string}
 */
export const EXCLUDE_FILTER = 'exclude';

export interface Filter {
	name: string;
	type: string;
	mode: string;
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

function addFilter(filters: string, filter: Filter): string {
	const decoded = decodeFilters(filters);
	decoded.filters.push(filter);
	return encodeFilters(decoded);
}

function removeFilter(filters: string, filter: Filter): string {
	// decode the provided filters
	const decoded = decodeFilters(filters);
	const index = _.findIndex(decoded.filters, f => {
		return _.isEqual(f, filter);
	});
	if (index !== -1) {
		decoded.filters.splice(index, 1);
	}
	// encode the filters back into a url string
	return encodeFilters(decoded);
}

export function addFilterToRoute(component: Vue, filter: Filter) {
	// retrieve the filters from the route
	const filters = routeGetters.getRouteFilters(component.$store);
	// merge the updated filters back into the route query params
	const updated = addFilter(filters, filter);
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		filters: updated
	});
	component.$router.push(entry);
}

export function removeFilterFromRoute(component: Vue, filter: Filter) {
	// retrieve the filters from the route
	const filters = routeGetters.getRouteFilters(component.$store);
	// merge the updated filters back into the route query params
	const updated = removeFilter(filters, filter);
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		filters: updated
	});
	component.$router.push(entry);
}

export function removeFiltersByName(component: Vue, name: string) {
	// retrieve the filters from the route
	const filters = routeGetters.getRouteFilters(component.$store);
	const decoded = decodeFilters(filters);
	decoded.filters = decoded.filters.filter(filter => {
		return (filter.name !== name);
	});
	const encoded = encodeFilters(decoded);
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		filters: encoded
	});
	component.$router.push(entry);
}
