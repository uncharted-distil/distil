import { Store } from 'vuex';
import { TableData } from '../store/dataset/index';
import { Highlight, HighlightRoot } from '../store/highlights/index';
import { Dictionary } from '../util/dict';
import { Filter, NUMERICAL_FILTER } from '../util/filters';
import { getters as routeGetters } from '../store/route/module';
import { getters as highlightGetters } from '../store/highlights/module';
import { overlayRouteEntry } from '../util/routes'
import { FilterParams } from '../util/filters'
import { getFilterType, getVarType, isMetaType, addMetaPrefix } from '../util/types'
import _ from 'lodash';
import Vue from 'vue';

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

export function createFilterFromHighlightRoot(store: Store<any>, highlightRoot: HighlightRoot, mode: string): Filter {
	if (!highlightRoot || highlightRoot.value == null) {
		return null;
	}
	// inject metadata prefix for metadata vars
	let key = highlightRoot.key;
	const type = getVarType(store, key);
	if (isMetaType(type)) {
		key = addMetaPrefix(key);
	}
	const filterType = getFilterType(type);
	if (_.isString(highlightRoot.value)) {
		return {
			key: key,
			type: filterType,
			mode: mode,
			categories: [highlightRoot.value]
		};
	}
	if (highlightRoot.value.from !== undefined && highlightRoot.value.to !== undefined) {
		return {
			key: key,
			type: NUMERICAL_FILTER,
			mode: mode,
			min: highlightRoot.value.from,
			max: highlightRoot.value.to
		};
	}
	return null;
}

export function addHighlightToFilterParams(store: any, filterParams: FilterParams, highlightRoot: HighlightRoot, mode: string): FilterParams {
	const params = _.cloneDeep(filterParams);
	const highlightFilter = createFilterFromHighlightRoot(store, highlightRoot, mode);
	if (highlightFilter) {
		params.filters.push(highlightFilter);
	}
	return params;
}

export function parseHighlightSamples(data: TableData): Dictionary<string[]>  {
	const samples: Dictionary<string[]> = {};
	if (!data) {
		return samples;
	}
	for (let rowIdx = 0; rowIdx < data.values.length; rowIdx++) {
		for (const [colIdx, col] of data.columns.entries()) {
			const val = data.values[rowIdx][colIdx];
			let colData = samples[col.key];
			if (!colData) {
				colData = [];
				samples[col.key] = colData;
			}
			colData.push(val);
		}
	}
	return samples;
}

export function updateHighlightRoot(component: Vue, highlightRoot: HighlightRoot) {
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		highlights: encodeHighlights(highlightRoot),
		row: null // clear row
	});
	component.$router.push(entry);
}

export function clearHighlightRoot(component: Vue) {
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		highlights: null,
		row: null // clear row
	});
	component.$router.push(entry);
}

export function getHighlights(store: Store<any>): Highlight {
	return {
		root: routeGetters.getDecodedHighlightRoot(store),
		values: {
			samples: highlightGetters.getHighlightedSamples(store),
			summaries: highlightGetters.getHighlightedSummaries(store)
		}
	};
}
