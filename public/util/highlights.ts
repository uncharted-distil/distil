import { Highlight, HighlightRoot } from '../store/highlights/index';
import { Filter, FilterParams, CATEGORICAL_FILTER, NUMERICAL_FILTER,
	BIVARIATE_FILTER, FEATURE_FILTER, TIMESERIES_FILTER } from '../util/filters';
import { getters as routeGetters } from '../store/route/module';
import { getters as datasetGetters } from '../store/dataset/module';
import { getters as highlightGetters } from '../store/highlights/module';
import { overlayRouteEntry } from '../util/routes';
import { getVarType, isFeatureType, addFeaturePrefix, isClusterType, addClusterPrefix, isTimeType } from '../util/types';
import _ from 'lodash';
import store from '../store/store';
import VueRouter from 'vue-router';

export function encodeHighlights(highlightRoot: HighlightRoot): string {
	if (_.isEmpty(highlightRoot)) {
		return null;
	}
	return btoa(JSON.stringify(highlightRoot));
}

export function decodeHighlights(highlightRoot: string): HighlightRoot {
	if (_.isEmpty(highlightRoot)) {
		return null;
	}
	return JSON.parse(atob(highlightRoot)) as HighlightRoot;
}

export function createFilterFromHighlightRoot(highlightRoot: HighlightRoot, mode: string): Filter {
	if (!highlightRoot || highlightRoot.value === null) {
		return null;
	}

	// inject metadata prefix for metadata vars
	let key = highlightRoot.key;

	const variables = datasetGetters.getVariables(store);

	const variable = variables.find(v => v.colName === key);
	let grouping = null;
	if (variable && variable.grouping) {
		if (variable.grouping.type === 'timeseries') {
			key = variable.grouping.properties.clusterCol;
			key = addClusterPrefix(key);
		}
		grouping = variable.grouping;
	}

	const type = getVarType(key);

	if (isFeatureType(type)) {
		key = addFeaturePrefix(key);
		return {
			key: key,
			type: FEATURE_FILTER,
			mode: mode,
			categories: [highlightRoot.value]
		};
	}

	if (_.isString(highlightRoot.value)) {
		return {
			key: key,
			type: CATEGORICAL_FILTER,
			mode: mode,
			categories: [highlightRoot.value]
		};
	}

	const isTimeseriesAnalysis = !!routeGetters.getRouteTimeseriesAnalysis(store);
	if (isTimeseriesAnalysis) {
		// TODO: fix this later
		return null;
	}

	if (highlightRoot.value.from !== undefined &&
		highlightRoot.value.to !== undefined) {

		// TODO: we currently have no support for filter timeseries data by
		// ranges and handle it in the client.
		if (grouping && grouping.type === TIMESERIES_FILTER) {
			return null;
		}

		return {
			key: key,
			type: NUMERICAL_FILTER,
			mode: mode,
			min: highlightRoot.value.from,
			max: highlightRoot.value.to
		};
	}
	if (highlightRoot.value.minX !== undefined &&
		highlightRoot.value.maxX !== undefined &&
		highlightRoot.value.minY !== undefined &&
		highlightRoot.value.maxY !== undefined) {
		return {
			key: key,
			type: BIVARIATE_FILTER,
			mode: mode,
			minX: highlightRoot.value.minX,
			maxX: highlightRoot.value.maxX,
			minY: highlightRoot.value.minY,
			maxY: highlightRoot.value.maxY,
		};
	}
	return null;
}

export function addHighlightToFilterParams(filterParams: FilterParams, highlightRoot: HighlightRoot, mode: string): FilterParams {
	const params = _.cloneDeep(filterParams);
	const highlightFilter = createFilterFromHighlightRoot(highlightRoot, mode);
	if (highlightFilter) {
		params.filters.push(highlightFilter);
	}
	return params;
}

export function updateHighlightRoot(router: VueRouter, highlightRoot: HighlightRoot) {
	const entry = overlayRouteEntry(routeGetters.getRoute(store), {
		highlights: encodeHighlights(highlightRoot),
		row: null // clear row
	});
	router.push(entry);
}

export function clearHighlightRoot(router: VueRouter) {
	const entry = overlayRouteEntry(routeGetters.getRoute(store), {
		highlights: null,
		row: null // clear row
	});
	router.push(entry);
}

export function getHighlights(): Highlight {
	return {
		root: routeGetters.getDecodedHighlightRoot(store),
		values: {
			summaries: highlightGetters.getHighlightedSummaries(store)
		}
	};
}
