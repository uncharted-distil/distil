import { RangeHighlights, ValueHighlights } from '../store/data/index';
import { Dictionary } from '../util/dict';
import _ from 'lodash';
import Vue from 'vue';

// Highlights table rows with values that are currently marked as highlighted.  Uses a supplied highligh
// context ID to enure that something like a table selection doesn't trigger additional table highlight
// updates.
export function updateTableHighlights(
	tableData: Dictionary<any>[],
	highlightRanges: RangeHighlights,
	highlightValues: ValueHighlights,
	highlightContext: string) {

	// skip highlighting when the context is the originating table
	if (!_.isEmpty(highlightRanges.ranges) && highlightRanges.context !== highlightContext) {
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
	} else if (!_.isEmpty(highlightValues.values) && highlightValues.context !== highlightContext) {
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
