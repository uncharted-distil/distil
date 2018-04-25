import { Store } from 'vuex';
import { Data, Highlight, HighlightRoot } from '../store/data/index';
import { Dictionary } from '../util/dict';
import { Filter, CATEGORICAL_FILTER, NUMERICAL_FILTER } from '../util/filters';
import { getters as routeGetters } from '../store/route/module';
import { getters as dataGetters } from '../store/data/module';
import { overlayRouteEntry} from '../util/routes'
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

export function createFilterFromHighlightRoot(highlightRoot: HighlightRoot, mode: string): Filter {
	if (!highlightRoot || highlightRoot.value == null) {
		return null;
	}
	if (_.isString(highlightRoot.value)) {
		return {
			name: highlightRoot.key,
			type: CATEGORICAL_FILTER,
			mode: mode,
			categories: [highlightRoot.value]
		};
	}
	if (highlightRoot.value.from !== undefined && highlightRoot.value.to !== undefined) {
		return {
			name: highlightRoot.key,
			type: NUMERICAL_FILTER,
			mode: mode,
			min: highlightRoot.value.from,
			max: highlightRoot.value.to
		};
	}
	return null;
}

export function parseHighlightSamples(data: Data): Dictionary<string[]>  {
	const samples: Dictionary<string[]> = {};
	for (let rowIdx = 0; rowIdx < data.values.length; rowIdx++) {
		for (const [colIdx, col] of data.columns.entries()) {
			const val = data.values[rowIdx][colIdx];
			let colData = samples[col];
			if (!colData) {
				colData = [];
				samples[col] = colData;
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
			samples: dataGetters.getHighlightedSamples(store),
			summaries: dataGetters.getHighlightedSummaries(store)
		}
	};
}
