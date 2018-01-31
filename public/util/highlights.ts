import { Store } from 'vuex';
import { Data } from '../store/data/index';
import { Dictionary } from '../util/dict';
import { Filter, CATEGORICAL_FILTER, NUMERICAL_FILTER } from '../util/filters';
import { getters as routeGetters } from '../store/route/module';
import { getters as dataGetters } from '../store/data/module';
import { overlayRouteEntry} from '../util/routes'
import _ from 'lodash';
import Vue from 'vue';

export interface Range {
	to: number;
	from: number;
}

export interface HighlightRoot {
	context: string;
	key: string;
	value: any;
}

export interface Highlights {
	root: HighlightRoot;
	values: Dictionary<string[]>;
}

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

// Highlights table rows with values that are currently marked as highlighted.  Uses a supplied highlight
// context ID to enure that something like a table selection doesn't trigger additional table highlight
// updates.
export function updateTableHighlights(tableData: Dictionary<any>[], highlightValues: Highlights, highlightContext: string) {

	// skip highlighting when the context is the originating table
	if (_.get(highlightValues, 'root.context', highlightContext) !== highlightContext) {
		// for the table, we're interested only in rows that have data that matches the value/range
		// described by the highlight root
		_.forEach(tableData, (row, rowNum) => {
			const value = row[highlightValues.root.key];
			// range case (root selection is numerical facet)
			if (_.get(highlightValues, 'root.value.from', NaN) <= value &&
				_.get(highlightValues, 'root.value.to', NaN) >= value) {
				row._rowVariant = 'info';
			} else if (_.get(highlightValues, 'root.value') === value) {
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
				const args = smoothScroll ? { behavior: 'smooth' } : {};
				selectedElem[0].scrollIntoView(args);
			}
		}
	});
}

export function createFilterFromHighlightRoot(highlightRoot: HighlightRoot): Filter {
	if (_.isString(highlightRoot.value)) {
		return {
			name: highlightRoot.key,
			type: CATEGORICAL_FILTER,
			enabled: true,
			categories: [highlightRoot.value]
		};
	}
	return {
		name: highlightRoot.key,
		type: NUMERICAL_FILTER,
		enabled: true,
		min: highlightRoot.value.from,
		max: highlightRoot.value.to
	};
}

export function parseHighlightValues(data: Data): Dictionary<string[]> {
	const highlights: Map<string, Set<any>> = new Map();
	for (let rowIdx = 0; rowIdx < data.values.length; rowIdx++) {
		for (const [colIdx, col] of data.columns.entries()) {
			const val = data.values[rowIdx][colIdx];
			let colData = highlights.get(col);
			if (!colData) {
				colData = new Set<string>();
				highlights.set(col, colData);
			}
			colData.add(val);
		}
	}
	const storeHighlights: Dictionary<string[]> = {};
	for (const [key, values] of highlights) {
		storeHighlights[key] = Array.from(values);
	}
	return storeHighlights;
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

export function getHighlights(store: Store<any>): Highlights {
	const rootHighlights = routeGetters.getDecodedHighlightRoot(store);
	if (!rootHighlights) {
		return {} as Highlights;
	}
	const values = dataGetters.getHighlightedValues(store);
	return {
		root: rootHighlights,
		values: values
	};
}
