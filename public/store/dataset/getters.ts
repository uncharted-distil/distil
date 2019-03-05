import { Variable, TimeseriesExtrema, DatasetState, Dataset, VariableSummary, TableData, TableRow, TableColumn } from './index';
import { Dictionary } from '../../util/dict';
import { getTableDataItems, getTableDataFields } from '../../util/data';
import { Grouping, getGroupings } from '../../util/groupings';

export const getters = {

	getDatasets(state: DatasetState): Dataset[] {
		return state.datasets;
	},

	getFilteredDatasets(state: DatasetState): Dataset[] {
		return state.filteredDatasets;
	},

	getVariables(state: DatasetState): Variable[] {
		const groupings = getGroupings();
		return state.variables.filter(v => {
			// not hidden by a grouping
			return groupings.filter(grouping => grouping.hidden[v.colName]).length === 0;
		});
		// return state.variables;
	},

	getGroupings(state: DatasetState): Grouping[] {
		return getGroupings();
	},

	getVariablesMap(state: DatasetState): Dictionary<Variable> {
		const map = {};
		state.variables.forEach(variable => {
			map[variable.colName] = variable;
			map[variable.colName.toLowerCase()] = variable;
		});
		return map;
	},

	getVariableTypesMap(state: DatasetState): Dictionary<string> {
		const map = {};
		state.variables.forEach(variable => {
			map[variable.colName] = variable.colType;
			map[variable.colName.toLowerCase()] = variable.colType;
		});
		return map;
	},

	getVariableSummaries(state: DatasetState): VariableSummary[] {
		return state.variableSummaries;
	},

	getFiles(state: DatasetState): Dictionary<any> {
		return state.files;
	},

	getTimeseriesExtrema(state: DatasetState): Dictionary<TimeseriesExtrema> {
		return state.timeseriesExtrema;
	},

	getTimeseries(state: DatasetState): Dictionary<Dictionary<number[][]>> {
		return state.timeseries;
	},

	getJoinDatasetsTableData(state: DatasetState): Dictionary<TableData> {
		return state.joinTableData;
	},

	hasIncludedTableData(state: DatasetState): boolean {
		return !!state.includedTableData;
	},

	getIncludedTableData(state: DatasetState): TableData {
		return state.includedTableData;
	},

	getIncludedTableDataNumRows(state: DatasetState): number {
		return state.includedTableData ? state.includedTableData.numRows : 0;
	},

	getIncludedTableDataItems(state: DatasetState, getters: any): TableRow[] {
		return getTableDataItems(state.includedTableData);
	},

	getIncludedTableDataFields(state: DatasetState, getters: any): Dictionary<TableColumn> {
		return getTableDataFields(state.includedTableData);
	},

	hasExcludedTableData(state: DatasetState): boolean {
		return !!state.excludedTableData;
	},

	getExcludedTableData(state: DatasetState): TableData {
		return state.excludedTableData;
	},

	getExcludedTableDataNumRows(state: DatasetState): number {
		return state.excludedTableData ? state.excludedTableData.numRows : 0;
	},

	getExcludedTableDataItems(state: DatasetState, getters: any): TableRow[] {
		return getTableDataItems(state.excludedTableData);
	},

	getExcludedTableDataFields(state: DatasetState, getters: any): Dictionary<TableColumn> {
		return getTableDataFields(state.excludedTableData);
	}
};
