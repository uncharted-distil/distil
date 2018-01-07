import { RangeHighlights, ValueHighlights } from '../store/data/index';
import { Dictionary } from '../util/dict';
import _ from 'lodash';

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
		_.forEach(tableData, row => {
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
		_.forEach(tableData, row => {
			_.forEach(row, (value, name) => {
				if (highlightValues.values[name] && highlightValues.values[name] === value) {
					row._rowVariant = 'info';
					return false;
				} else {
					row._rowVariant = false;
				}
			});
		});
	}
}
