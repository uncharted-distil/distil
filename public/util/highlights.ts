import { RangeHighlights, ValueHighlights, Data } from '../store/data/index';
import { Dictionary } from '../util/dict';
import { Filter } from '../util/filters';
import { getVarFromTarget } from '../util/data';
import { getters as routeGetters } from '../store/route/module';
import { mutations as dataMutations, actions as dataActions } from '../store/data/module';
import _ from 'lodash';
import Vue from 'vue';
import { AxiosPromise } from 'axios';

// Highlights table rows with values that are currently marked as highlighted.  Uses a supplied highligh
// context ID to enure that something like a table selection doesn't trigger additional table highlight
// updates.
export function updateTableHighlights(
	tableData: Dictionary<any>[],
	highlightRanges: RangeHighlights,
	highlightValues: ValueHighlights,
	highlightContext: string) {

	// skip highlighting when the context is the originating table
	if (!_.isEmpty(highlightRanges.ranges) && highlightRanges.root.context !== highlightContext) {
		// highlight any table row that has a value in the feature range
		_.forEach(tableData, (row, rowNum) => {
			_.forEach(row, (value, name) => {
				const range = highlightRanges.ranges[name];
				if (range && range.from <= value && range.to >= value) {
					row._rowVariant = 'info';
					return false;
				}
				row._rowVariant = null;
			});
		});
	} else if (!_.isEmpty(highlightValues.values) && highlightValues.root.context !== highlightContext) {
		// highlight any table row that is in the value map
		_.forEach(tableData, (row, rowNum) => {
			_.forEach(row, (value, name) => {
				if (highlightValues.values[name] && highlightValues.values[name] === value) {
					row._rowVariant = 'info';
					return false;
				} else {
					row._rowVariant = null;
				}
			});
		});
	}
}

// Scrolls table to first highlighted row
export function scrollToFirstHighlight(component: Vue, refName: string) {
	// No support for scroll-to in the bootstrap table.  Author suggests finding the
	// the row element and using the scrollIntoView function - we put this into a nextTick()
	// because need it to kick off after the virtual DOM update
	component.$nextTick(() => {
		const tableRef = <Vue>component.$refs[refName];
		if (tableRef && tableRef.$el) {
			const selectedElem = $(tableRef.$el).find('.table-info');
			if (selectedElem && selectedElem.length > 0) {
				// Enabling smooth scrolling seems to cause some sort of contention within the browser
				// resulting in only one table scrolling.  Need to use default 'instant' scroll for now.
				// Could also use
				// selectedElem[0].scrollIntoView({ behavior: 'smooth'});
				selectedElem[0].scrollIntoView();
			}
		}
	});
}

// Given a key/value from a facet/histogram click event and a corresponding filter,
// generate a set of value highlights.
export function updateDataHighlights(component: Vue, context: string,  key: string, value: string, selectFilter: Filter) {
	const dataset = routeGetters.getRouteDataset(component.$store);
	const filters = Array.from(routeGetters.getDecodedFilters(component.$store));

	const index = _.findIndex(filters, f => f.name === key);
	if (index < 0) {
		filters.push(selectFilter);
	} else {
		filters[index] = selectFilter;
	}
	// fetch the data using the supplied filtered
	const resultPromise = dataActions.fetchData(component.$store, { dataset: dataset, filters: filters, inclusive: true });
	updateHighlights(component, resultPromise, context, key, value);
}

// Given a key/value from a facet/histogram click event and a corresponding filter,
// generate a set of value highlights.
export function updateResultHighlights(component: Vue, context: string, key: string, value: string, selectFilter: Filter) {
	const dataset = routeGetters.getRouteDataset(component.$store);
	const filters = Array.from(routeGetters.getDecodedFilters(component.$store));
	const pipelineId = routeGetters.getRoutePipelineId(component.$store);

	selectFilter.name = getVarFromTarget(selectFilter.name);

	const index = _.findIndex(filters, f => f.name.toLowerCase() === selectFilter.name.toLowerCase());
	if (index < 0) {
		filters.push(selectFilter);
	} else {
		filters[index] = selectFilter;
	}

	// fetch the data using the supplied filtered
	const resultPromise = dataActions.fetchResults(component.$store, { pipelineId: pipelineId, dataset: dataset, filters: filters });
	updateHighlights(component, resultPromise, context, key, value);
}

// Given returned data,
function updateHighlights(component: Vue, promise: AxiosPromise<Data>, context: string, key: string, value: string) {
	promise.then(response => {
		const data = response.data;
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
		dataMutations.highlightFeatureValues(component.$store, {
			root: {
				context: context,
				key: key,
				value: value
			},
			values: storeHighlights
		});
	})
	.catch(error => console.error(error));
}
