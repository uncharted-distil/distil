import { RowSelection } from '../store/data/index';
import { getters as routeGetters } from '../store/route/module';
import { overlayRouteEntry} from '../util/routes'
import _ from 'lodash';
import Vue from 'vue';

export function encodeRowSelection(row: RowSelection): string {
	if (_.isEmpty(row)) {
		return null;
	}
	return btoa(JSON.stringify(row));
}

export function decodeRowSelection(row: string): RowSelection {
	if (_.isEmpty(row)) {
		return null;
	}
	return JSON.parse(atob(row)) as RowSelection;
}

export function updateTableRowSelection(items: any, selection: RowSelection, context: string) {
	// skip highlighting when the context is the originating table
	if (!selection) {
		return items;
	}

	if (selection.context !== context) {
		return items;
	}
	// // clear selections
	_.forEach(items, (row, rowNum) => {
		row._rowVariant = null;
	});
	// add selection
	if (items[selection.index]) {
		items[selection.index]._rowVariant = 'info';
	}
	return items;
}

export function updateRowSelection(component: Vue, row: RowSelection) {
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		row: encodeRowSelection(row),
	});
	component.$router.push(entry);
}

export function clearRowSelection(component: Vue) {
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		row: null
	});
	component.$router.push(entry);
}
