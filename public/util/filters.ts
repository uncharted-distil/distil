/**
 *
 *    Copyright © 2021 Uncharted Software Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

import _ from "lodash";
import { getters as routeGetters } from "../store/route/module";
import { overlayRouteEntry } from "./routes";
import store from "../store/store";
import VueRouter from "vue-router";

/**
 * Categorical filter, omitting documents that do not contain the provided
 * categories in the variable.
 * @constant {string}
 */
export const CATEGORICAL_FILTER = "categorical";

/**
 * Numerical filter, omitting documents that do not fall within the provided
 * variable range.
 * @constant {string}
 */
export const NUMERICAL_FILTER = "numerical";

/**
 * Datetime filter, omitting documents that do not fall within the provided
 * variable range.
 * @constant {string}
 */
export const DATETIME_FILTER = "datetime";

/**
 * Bivariate filter, omitting documents that do not fall within the provided
 * variable range.
 * @constant {string}
 */
export const BIVARIATE_FILTER = "bivariate";

/**
 * GeoBounds filter, omitting documents that do not fall within the provided
 * variable range.
 * @constant {string}
 */
export const GEOBOUNDS_FILTER = "geobounds";

/**
 * Timeseries filter, omitting documents that do not fall within the provided
 * timeseries range.
 * @constant {string}
 */
export const TIMESERIES_FILTER = "timeseries";

/**
 * Row filter, omitting documents that have the specified d3mIndices;
 * @constant {string}
 */
export const ROW_FILTER = "row";

/**
 * Cluster filter, omitting documents that have the specified cluster value;
 * @constant {string}
 */
export const CLUSTER_FILTER = "cluster";

/**
 * Vector filter, omitting documents that have the specified vector value;
 * @constant {string}
 */
export const VECTOR_FILTER = "vector";

/**
 * Text filter, omitting documents that have the specified text value;
 * @constant {string}
 */
export const TEXT_FILTER = "text";

/**
 * Geocoordinate filter, omitting documents that have the specified bounding box value;
 * @constant {string}
 */
export const GEOCOORDINATE_FILTER = "geocoordinate";

/**
 * Include filter, excluding documents that do not fall within the filter.
 * @constant {string}
 */
export const INCLUDE_FILTER = "include";

/**
 * Exclude filter, excluding documents that fall outside the filter.
 * @constant {string}
 */
export const EXCLUDE_FILTER = "exclude";

export function invertFilter(filter: string): string {
  return filter === INCLUDE_FILTER ? EXCLUDE_FILTER : INCLUDE_FILTER;
}

export interface Filter {
  type: string;
  mode: string;
  displayName?: string;
  key?: string;
  min?: number;
  max?: number;
  minX?: number;
  maxX?: number;
  minY?: number;
  maxY?: number;
  nestedType?: string;
  categories?: string[];
  d3mIndices?: string[];
}

export interface FilterObject {
  list: Filter[];
  invert?: boolean;
}

export interface FilterParams {
  highlights: FilterObject;
  filters: FilterObject;
  variables: string[];
  size?: number;
  dataMode?: string;
  isHighlight?: boolean;
}

/**
 * Decodes the filters from the route string into an array.
 *
 * @param {string} filters - The filters from the route query string.
 *
 * @returns {Filter[]} The decoded filter object.
 */
export function decodeFilters(filters: string): FilterObject {
  if (_.isEmpty(filters)) {
    return { list: [] };
  }
  return { list: JSON.parse(atob(filters)) as Filter[] };
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
 * Resolves any redundant row include / excludes such that there are only a
 * maximum of two row filters, one for includes, one for excludes.
 */
function dedupeRowFilters(filters: FilterObject): Filter[] {
  const rowFilters = filters.list.filter(
    (filter) => filter.type === ROW_FILTER
  );
  const remaining = filters.list.filter((filter) => filter.type !== ROW_FILTER);

  const included = {};
  const excluded = {};
  const d3mIndices = {};

  rowFilters.forEach((filter, filterIndex) => {
    filter.d3mIndices.forEach((d3mIndex) => {
      if (filter.mode === INCLUDE_FILTER) {
        included[d3mIndex] = filterIndex;
      } else {
        excluded[d3mIndex] = filterIndex;
      }
      d3mIndices[d3mIndex] = true;
    });
  });

  const includes = {
    type: ROW_FILTER,
    mode: INCLUDE_FILTER,
    d3mIndices: [],
  };
  const excludes = {
    type: ROW_FILTER,
    mode: EXCLUDE_FILTER,
    d3mIndices: [],
  };

  _.keys(d3mIndices).forEach((d3mIndex) => {
    const includedIndex = included[d3mIndex];
    const excludedIndex = excluded[d3mIndex];

    // NOTE: filters should be in the order they are created
    if (includedIndex >= 0 && excludedIndex >= 0) {
      // if excluded and then included, omit filter entirely
      return;
    }

    if (includedIndex >= 0) {
      includes.d3mIndices.push(d3mIndex);
      return;
    }

    if (excludedIndex >= 0) {
      excludes.d3mIndices.push(d3mIndex);
    }
  });

  if (includes.d3mIndices.length > 0) {
    remaining.push(includes);
  }

  if (excludes.d3mIndices.length > 0) {
    remaining.push(excludes);
  }

  return remaining;
}

function addFilter(filters: string, filter: Filter | Filter[]): string {
  const decoded = decodeFilters(filters);
  if (Array.isArray(filter)) {
    decoded.list = [...decoded.list, ...filter];
  } else {
    decoded.list.push(filter as Filter);
  }
  return encodeFilters(dedupeRowFilters(decoded));
}

function removeFilter(filters: string, filter: Filter): string {
  // decode the provided filters
  const decoded = decodeFilters(filters);
  const index = _.findIndex(decoded.list, (f) => {
    return _.isEqual(f, filter);
  });
  if (index !== -1) {
    decoded.list.splice(index, 1);
  }
  // encode the filters back into a url string
  return encodeFilters(decoded.list);
}

export function hasFilterInRoute(variable: string): boolean {
  // retrieve the filters from the route

  const filters = routeGetters.getRouteFilters(store);
  const decoded = decodeFilters(filters);
  return (
    decoded.list.filter((filter) => {
      return filter.key && filter.key === variable;
    }).length > 0
  );
}

export function addFilterToRoute(router: VueRouter, filter: Filter | Filter[]) {
  // retrieve the filters from the route
  const filters = routeGetters.getRouteFilters(store);
  // merge the updated filters back into the route query params
  const updated = addFilter(filters, filter);
  const entry = overlayRouteEntry(routeGetters.getRoute(store), {
    filters: updated,
  });
  router.push(entry).catch((err) => console.warn(err));
}

export function removeFilterFromRoute(router: VueRouter, filter: Filter) {
  // retrieve the filters from the route
  const filters = routeGetters.getRouteFilters(store);
  // merge the updated filters back into the route query params
  const updated = removeFilter(filters, filter);
  const entry = overlayRouteEntry(routeGetters.getRoute(store), {
    filters: updated,
  });
  router.push(entry).catch((err) => console.warn(err));
}

export function removeFiltersByName(router: VueRouter, key: string) {
  // retrieve the filters from the route
  const filters = routeGetters.getRouteFilters(store);
  const decoded = decodeFilters(filters);
  decoded.list = decoded.list.filter((filter) => {
    return filter.key !== key;
  });
  deepUpdateFiltersInRoute(router, decoded.list);
}

export function deepUpdateFiltersInRoute(router: VueRouter, filters: Filter[]) {
  const encoded = encodeFilters(filters);
  const entry = overlayRouteEntry(routeGetters.getRoute(store), {
    filters: encoded,
  });
  router.push(entry).catch((err) => console.warn(err));
}
