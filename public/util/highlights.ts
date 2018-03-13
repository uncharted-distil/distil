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

// Highlight table rows with values that are currently marked as highlighted.  Uses a supplied highlight
// context ID to enure that something like a table selection doesn't trigger additional table highlight
// updates.
export function updateTableHighlights(tableData: Dictionary<any>[], highlight: Highlight, highlightContext: string) {

	// skip highlighting when the context is the originating table
	if (_.get(highlight, 'root.context', highlightContext) !== highlightContext) {
		// for the table, we're interested only in rows that have data that matches the value/range
		// described by the highlight root
		_.forEach(tableData, (row, rowNum) => {
			const value = row[highlight.root.key];
			// range case (root selection is numerical facet)
			if (_.get(highlight, 'root.value.from', NaN) <= value &&
				_.get(highlight, 'root.value.to', NaN) >= value) {
				row._rowVariant = 'info';
			} else if (_.get(highlight, 'root.value') === value) {
				// single value case (root selection is a categorical facet)
				row._rowVariant = 'info';
			} else {
				row._rowVariant = null;
			}
		});
	}
}

// Scrolls table to first highlighted row
export function scrollToFirstHighlight(component: Vue, refName: string, smoothScroll: boolean) {
	// No support for scroll-to in the bootstrap table.  Author suggests finding the
	// the row element and using the scrollIntoView function - we put this into a nextTick()
	// because need it to kick off after the virtual DOM update
	component.$nextTick(() => {
		const tableRef = <Vue>component.$refs[refName];
		if (tableRef && tableRef.$el) {
			const selectedElem = $(tableRef.$el).find('.table-info');
			if (selectedElem && selectedElem.length > 0) {
				// Enabling smooth scrolling seems to cause some sort of contention within the browser
				// resulting in only one table scrolling.  We can enable it for the select screen, but
				// not the result screen.
				const args: ScrollIntoViewOptions = smoothScroll ? { behavior: 'smooth' } : {};
				selectedElem[0].scrollIntoView(args);
			}
		}
	});
}

export function createFilterFromHighlightRoot(highlightRoot: HighlightRoot, mode: string): Filter {
	if (highlightRoot.value == null) {
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
	});
	component.$router.push(entry);
}

export function clearHighlightRoot(component: Vue) {
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		highlights: null
	});
	component.$router.push(entry);
}

export function getHighlights(store: Store<any>): Highlight {
	const rootHighlights = routeGetters.getDecodedHighlightRoot(store);
	if (!rootHighlights) {
		return {} as Highlight;
	}
	return {
		root: rootHighlights,
		values: {
			samples: dataGetters.getHighlightedSamples(store),
			summaries: dataGetters.getHighlightedSummaries(store)
		}
	};
}
