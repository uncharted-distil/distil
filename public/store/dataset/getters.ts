import { Variable, Extrema, TimeseriesExtrema, DatasetState, Dataset, VariableSummary, TimeseriesSummary, TableData, TableRow, TableColumn } from './index';
import { Dictionary } from '../../util/dict';
import { getTableDataItems, getTableDataFields } from '../../util/data';

export const getters = {

	getDatasets(state: DatasetState): Dataset[] {
		return state.datasets;
	},

	getFilteredDatasets(state: DatasetState): Dataset[] {
		return state.filteredDatasets;
	},

	getVariables(state: DatasetState, getters: any): Variable[] {
		const timeseriesAnalysis = getters.getRouteTimeseriesAnalysis;
		if (timeseriesAnalysis) {
			// don't return the time var
			return state.variables.filter(v => v.colName !== timeseriesAnalysis);
		}
		return state.variables;
	},

	getPendingRequests(state: DatasetState) {
		return state.pendingRequests;
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

	getVariableRankings(state: DatasetState): Dictionary<Dictionary<number>> {
		return state.variableRankings;
	},

	getVariableSummaries(state: DatasetState): VariableSummary[] {
		return state.includedSet.variableSummaries;
	},

	getIncludedVariableSummaries(state: DatasetState): VariableSummary[] {
		return state.includedSet.variableSummaries;
	},

	getExcludedVariableSummaries(state: DatasetState): VariableSummary[] {
		return state.excludedSet.variableSummaries;
	},

	getTimeseriesAnalysisVariable(state: DatasetState, getters: any): Variable {
		const timeseriesAnalysis = getters.getRouteTimeseriesAnalysis;
		if (timeseriesAnalysis) {
			return getters.getVariablesMap[timeseriesAnalysis];
		}
		return null;
	},

	getTimeseriesAnalysisExtrema(state: DatasetState, getters: any): Extrema {
		const v = getters.getTimeseriesAnalysisVariable;
		if (v) {
			return {
				min: v.min,
				max: v.max
			};
		}
		return null;
	},

	getTimeseriesAnalysisRange(state: DatasetState, getters: any): number {
		const extrema = getters.getTimeseriesAnalysisExtrema;
		if (!extrema) {
			return undefined;
		}
		return extrema.max - extrema.min;
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
		return !!state.includedSet.tableData;
	},

	getIncludedTableData(state: DatasetState): TableData {
		return state.includedSet.tableData;
	},

	getIncludedTableDataNumRows(state: DatasetState): number {
		return state.includedSet.tableData ? state.includedSet.tableData.numRows : 0;
	},

	getIncludedTableDataItems(state: DatasetState, getters: any): TableRow[] {
		return getTableDataItems(state.includedSet.tableData);
	},

	getIncludedTableDataFields(state: DatasetState, getters: any): Dictionary<TableColumn> {
		return getTableDataFields(state.includedSet.tableData);
	},

	hasExcludedTableData(state: DatasetState): boolean {
		return !!state.excludedSet.tableData;
	},

	getExcludedTableData(state: DatasetState): TableData {
		return state.excludedSet.tableData;
	},

	getExcludedTableDataNumRows(state: DatasetState): number {
		return state.excludedSet.tableData ? state.excludedSet.tableData.numRows : 0;
	},

	getExcludedTableDataItems(state: DatasetState, getters: any): TableRow[] {
		return getTableDataItems(state.excludedSet.tableData);
	},

	getExcludedTableDataFields(state: DatasetState, getters: any): Dictionary<TableColumn> {
		return getTableDataFields(state.excludedSet.tableData);
	}
};
