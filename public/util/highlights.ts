import { RangeHighlights, ValueHighlights } from '../store/data/index';
import { Dictionary } from '../util/dict';
import _ from 'lodash';

// Highlights table rows with values that are currently marked as highlighted.  Uses a supplied highligh
// context ID to enure that something like a table selection doesn't trigger additional table highlight
// updates.
export function updateTableHighlights(
	tableData: Dictionary<any>,
	highlightRanges: RangeHighlights,
	highlightValues: ValueHighlights,
	highlightContext: string) {

		// if highlights are empty, clear everything
	if (_.isEmpty(highlightRanges.ranges) && _.isEmpty(highlightValues.values)) {
		_.forEach(tableData, row => row._rowVariant = null);
		return;
	}

	// skip highlighting when the context is the originating table
	if (!_.isEmpty(highlightRanges.ranges) && highlightRanges.context !== highlightContext) {
		// highlight any table row that has a value in the feature range
		_.forIn(tableData, row => {
			_.forIn(highlightRanges.ranges, (range, name) => {
				if (row[name] >= range.from && row[name] <= range.to) {
					row._rowVariant = 'info';
				} else {
					row._rowVariant = null;
				}
			});
		});
	} else if (!_.isEmpty(highlightValues.values) && highlightValues.context !== highlightContext) {
		// highlight any table row that is in the value map
		_.forIn(tableData, row => {
			_.forIn(highlightValues.values, (value, name) => row._rowVariant = row[name] === value ? 'info' : null);
		});
	}
}
