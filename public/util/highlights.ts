import { Highlight } from '../store/dataset/index';
import { Filter, FilterParams, CATEGORICAL_FILTER, NUMERICAL_FILTER,
	BIVARIATE_FILTER, FEATURE_FILTER, TIMESERIES_FILTER } from '../util/filters';
import { getters as routeGetters } from '../store/route/module';
import { getters as datasetGetters } from '../store/dataset/module';
import { overlayRouteEntry } from '../util/routes';
import { getVarType, isFeatureType, addFeaturePrefix, isClusterType, addClusterPrefix, isTimeType } from '../util/types';
import _ from 'lodash';
import store from '../store/store';
import VueRouter from 'vue-router';

export function encodeHighlights(highlight: Highlight): string {
	if (_.isEmpty(highlight)) {
		return null;
	}
	return btoa(JSON.stringify(highlight));
}

export function decodeHighlights(highlight: string): Highlight {
	if (_.isEmpty(highlight)) {
		return null;
	}
	return JSON.parse(atob(highlight)) as Highlight;
}

export function createFilterFromHighlight(highlight: Highlight, mode: string): Filter {
	if (!highlight || highlight.value === null) {
		return null;
	}

	// inject metadata prefix for metadata vars
	let key = highlight.key;

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
			categories: [highlight.value]
		};
	}

	if (_.isString(highlight.value)) {
		return {
			key: key,
			type: CATEGORICAL_FILTER,
			mode: mode,
			categories: [highlight.value]
		};
	}

	const isTimeseriesAnalysis = !!routeGetters.getRouteTimeseriesAnalysis(store);
	if (isTimeseriesAnalysis) {
		// TODO: fix this later
		return null;
	}

	if (highlight.value.from !== undefined &&
		highlight.value.to !== undefined) {

		// TODO: we currently have no support for filter timeseries data by
		// ranges and handle it in the client.
		if (grouping && grouping.type === TIMESERIES_FILTER) {
			return null;
		}

		return {
			key: key,
			type: NUMERICAL_FILTER,
			mode: mode,
			min: highlight.value.from,
			max: highlight.value.to
		};
	}
	if (highlight.value.minX !== undefined &&
		highlight.value.maxX !== undefined &&
		highlight.value.minY !== undefined &&
		highlight.value.maxY !== undefined) {
		return {
			key: key,
			type: BIVARIATE_FILTER,
			mode: mode,
			minX: highlight.value.minX,
			maxX: highlight.value.maxX,
			minY: highlight.value.minY,
			maxY: highlight.value.maxY,
		};
	}
	return null;
}

export function addHighlightToFilterParams(filterParams: FilterParams, highlight: Highlight, mode: string): FilterParams {
	const params = _.cloneDeep(filterParams);
	const highlightFilter = createFilterFromHighlight(highlight, mode);
	if (highlightFilter) {
		params.filters.push(highlightFilter);
	}
	return params;
}

export function updateHighlight(router: VueRouter, highlight: Highlight) {
	const entry = overlayRouteEntry(routeGetters.getRoute(store), {
		highlights: encodeHighlights(highlight),
		row: null // clear row
	});
	router.push(entry);
}

export function clearHighlight(router: VueRouter) {
	const entry = overlayRouteEntry(routeGetters.getRoute(store), {
		highlights: null,
		row: null // clear row
	});
	router.push(entry);
}
