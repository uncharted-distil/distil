
import localStorage from 'store';

export interface Grouping {
	dataset: string;
	idCol: string;
	type: string;
	properties: Object;
}

export function createGrouping(grouping: Grouping) {
	let groupings = localStorage.get(grouping.dataset);
	if (!groupings) {
		groupings = [];
	}
	groupings.push(grouping);
	localStorage.set(grouping.dataset, groupings);
}

export function getGroupings(dataset: string): Grouping[] {
	return localStorage.get(dataset);
}
