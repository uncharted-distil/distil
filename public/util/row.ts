import { RowSelection, Row } from '../store/highlights/index';
import { D3M_INDEX_FIELD } from '../store/dataset/index';
import { getters as routeGetters } from '../store/route/module';
import { getters as dataGetters } from '../store/dataset/module';
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

export function getNumIncludedRows(component: Vue, selection: RowSelection): number {
	if (!selection || selection.d3mIndices.length === 0) {
		return 0;
	}
	const includedData = dataGetters.getIncludedTableDataItems(component.$store);
	const d3mIndices = {};
	selection.d3mIndices.forEach(index => {
		d3mIndices[index] = true;
	});
	return includedData.filter(data => d3mIndices[data[D3M_INDEX_FIELD]]).length;
}

export function getNumExcludedRows(component: Vue, selection: RowSelection,): number {
	if (!selection || selection.d3mIndices.length === 0) {
		return 0;
	}
	const excludedData = dataGetters.getExcludedTableDataItems(component.$store);
	const d3mIndices = {};
	selection.d3mIndices.forEach(index => {
		d3mIndices[index] = true;
	});
	return excludedData.filter(data => d3mIndices[data[D3M_INDEX_FIELD]]).length;
}

export function isRowSelected(selection: RowSelection, d3mIndex: number): boolean {
	if (!selection || selection.d3mIndices.length === 0) {
		return false;
	}
	for (let i=0; i<selection.d3mIndices.length; i++) {
		if (selection.d3mIndices[i] === d3mIndex) {
			return true;
		}
	}
	return false;
}

export function updateTableRowSelection(items: any, selection: RowSelection, context: string) {
	// clear selections
	_.forEach(items, (row, rowNum) => {
		row._rowVariant = null;
	});

	// skip highlighting when the context is the originating table
	if (!selection) {
		return items;
	}

	if (selection.context !== context) {
		return items;
	}
	// add selections
	const d3mIndices = {};
	selection.d3mIndices.forEach(index => {
		d3mIndices[index] = true;
	});
	items.forEach(item => {
		if (d3mIndices[item[D3M_INDEX_FIELD]]) {
			item._rowVariant = 'selected-row';
		}
	});
	return items;
}

export function getSelectedRows(component: Vue, selection: RowSelection): Row[] {
	if (!selection || selection.d3mIndices.length === 0) {
		return [];
	}

	const includedData = dataGetters.getIncludedTableDataItems(component.$store);
	const excludedData = dataGetters.getExcludedTableDataItems(component.$store);

	const d3mIndices = {};
	selection.d3mIndices.forEach(index => {
		d3mIndices[index] = true;
	});

	const rows = [];
	includedData.forEach((data, index) => {
		if (d3mIndices[data[D3M_INDEX_FIELD]]) {
			rows.push({
				index: index,
				included: true,
				cols: _.map(data, (value, key) => {
					return {
						key: key,
						value: value
					};
				})
			});
		}
	});
	excludedData.forEach((data, index) => {
		if (d3mIndices[data[D3M_INDEX_FIELD]]) {
			rows.push({
				index: index,
				included: false,
				cols: _.map(data, (value, key) => {
					return {
						key: key,
						value: value
					};
				})
			});
		}
	});

	return rows;
}

export function addRowSelection(component: Vue, context: string, selection: RowSelection, d3mIndex: number) {
	if (!selection || selection.context !== context) {
		selection = {
			context: context,
			d3mIndices: []
		};
	}
	selection.d3mIndices.push(d3mIndex);
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		row: encodeRowSelection(selection),
	});
	component.$router.push(entry);
}

export function removeRowSelection(component: Vue, context: string, selection: RowSelection, d3mIndex: number) {
	_.remove(selection.d3mIndices, r => {
		return r === d3mIndex;
	});
	if (selection.d3mIndices.length === 0) {
		selection = null;
	}
	const entry = overlayRouteEntry(routeGetters.getRoute(component.$store), {
		row: encodeRowSelection(selection),
	});
	component.$router.push(entry);
}
