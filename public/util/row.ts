import { RowSelection, Row } from '../store/highlights/index';
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

export function getNumIncludedRows(selection: RowSelection,): number {
	if (!selection) {
		return 0;
	}
	return selection.rows.filter(r => r.included).length;
}

export function getNumExcludedRows(selection: RowSelection,): number {
	if (!selection) {
		return 0;
	}
	return selection.rows.filter(r => !r.included).length;
}

export function isRowSelected(selection: RowSelection, index: number): boolean {
	if (!selection) {
		return false;
	}
	for (let i=0; i<selection.rows.length; i++) {
		if (selection.rows[i].index === index) {
			return true;
		}
	}
	return false;
}

export function updateRowSelection(component: Vue, context: string, selection: RowSelection, row: Row) {

	if (!selection || selection.context !== context) {
		selection = {
			context: context,
			rows: []
		};
	}
	if (row) {
		selection.rows.push(row);
	}

	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		row: encodeRowSelection(selection),
	});
	component.$router.push(entry);
}

export function clearRowSelection(component: Vue) {
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		row: null
	});
	component.$router.push(entry);
}
