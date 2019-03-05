import _ from 'lodash';
import localStorage from 'store';
import { getters as routeGetters } from '../store/route/module';
import store from '../store/store';
import { Dictionary } from './dict';

export interface GroupingProperties {
	xCol: string;
	yCol: string;
	clusterCol: string;
}

export interface Grouping {
	dataset: string;
	idCol: string;
	type: string;
	hidden: Dictionary<boolean>;
	properties?: GroupingProperties;
}

export function createGrouping(grouping: Grouping) {
	let groupings = localStorage.get(grouping.dataset);
	if (!groupings) {
		groupings = {};
	}
	groupings[grouping.idCol] = grouping;
	localStorage.set(`groupings:${grouping.dataset}`, groupings);
}

export function getGroupings(): Grouping[] {
	const dataset = routeGetters.getRouteDataset(store);
	const groups = localStorage.get(`groupings:${dataset}`);
	return _.map(groups, group => group);
}
