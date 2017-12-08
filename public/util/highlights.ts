import { Range } from '../store/data/index';
import { Dictionary } from '../util/dict';
import _ from 'lodash';

// Updates table rows that fall into the highlight range
export function updateTableHighlights(data: Dictionary<any>, highlightRanges: Range) {
	// if highlight feature ranges is empty, clear everything
	if (_.isEmpty(highlightRanges)) {
		_.forEach(data, row => row._rowVariant = null);
		return;
	}

	// highlight any table row that has a value in the feature range
	_.forIn(data, row => {
		_.forIn(highlightRanges, (range, name) => {
			if (row[name] >= range.from && row[name] <= range.to) {
				row._rowVariant = 'info';
			} else {
				row._rowVariant = null;
			}
		});
	});
}
