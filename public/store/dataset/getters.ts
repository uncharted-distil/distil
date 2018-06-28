import { Variable, DatasetState, Dataset, VariableSummary, TableData, TableRow, TableColumn, D3M_INDEX_FIELD } from './index';
import { Dictionary } from '../../util/dict';
import { getTableDataItems, validateData } from '../../util/data';

export const getters = {

	getDatasets(state: DatasetState): Dataset[] {
		return state.datasets;
	},

	getVariables(state: DatasetState): Variable[] {
		return state.variables;
	},

	getVariablesMap(state: DatasetState): Dictionary<Variable> {
		const map = {};
		state.variables.forEach(variable => {
			map[variable.name] = variable;
			map[variable.name.toLowerCase()] = variable;
		});
		return map;
	},

	getVariableTypesMap(state: DatasetState): Dictionary<string> {
		const map = {};
		state.variables.forEach(variable => {
			map[variable.name] = variable.type;
			map[variable.name.toLowerCase()] = variable.type;
		});
		return map;
	},

	getVariableSummaries(state: DatasetState): VariableSummary[] {
		return state.variableSummaries;
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
		return getTableDataItems(state.includedTableData, getters.getVariableTypesMap);
	},

	getIncludedTableDataFields(state: DatasetState, getters: any): Dictionary<TableColumn> {
		const data = state.includedTableData;
		if (validateData(data)) {
			const vmap = getters.getVariableTypesMap;
			const result = {};
			for (const col of data.columns) {
				if (col !== D3M_INDEX_FIELD) {
					result[col] = {
						label: col,
						type: vmap[col],
						sortable: true
					};
				}
			}
			return result;
		}
		return {};
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
		return getTableDataItems(state.excludedTableData, getters.getVariableTypesMap);
	},

	getExcludedTableDataFields(state: DatasetState, getters: any): Dictionary<TableColumn> {
		const data = state.excludedTableData;
		if (validateData(data)) {
			const vmap = getters.getVariableTypesMap;
			const result = {};
			for (const col of data.columns) {
				if (col !== D3M_INDEX_FIELD) {
					result[col] = {
						label: col,
						type: vmap[col],
						sortable: true
					};
				}
			}
			return result;
		}
		return {};
	}
}
