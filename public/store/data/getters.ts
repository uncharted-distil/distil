import _ from 'lodash';
import { FieldInfo, Variable, Data, DataState, Datasets, VariableSummary, TargetRow, Highlights, TableRow } from './index';
import { Filter, EMPTY_FILTER } from '../../util/filters';
import { TARGET_POSTFIX, PREDICTED_POSTFIX } from '../../util/data';
import { Dictionary } from '../../util/dict';
import { getPredictedIndex, getErrorIndex, getTargetIndex } from '../../util/data';
import { formatValue } from '../../util/types';

function validateData(data: Data) {
	return !_.isEmpty(data) &&
		!_.isEmpty(data.values) &&
		!_.isEmpty(data.columns);
}

function getDataItems(data: Data, typeMap: Dictionary<string>): TableRow[] {
	if (validateData(data)) {
		// convert fetched result data rows into table data rows
		return data.values.map((resultRow, rowIndex) => {
			const row = {} as TargetRow;
			resultRow.forEach((colValue, colIndex) => {
				const colName = data.columns[colIndex];
				const colType = typeMap[colName];
				row[colName] = formatValue(colValue, colType);
			});
			row._key = rowIndex;
			return row;
		});
	}
	return [];
}

export const getters = {
	getVariables(state: DataState): Variable[] {
		return state.variables;
	},

	getVariablesMap(state: DataState): Dictionary<Variable> {
		const map = {};
		state.variables.forEach(variable => {
			map[variable.name] = variable;
			map[variable.name.toLowerCase()] = variable;
		});
		return map;
	},

	getVariableTypesMap(state: DataState): Dictionary<string> {
		const map = {};
		state.variables.forEach(variable => {
			map[variable.name] = variable.type;
			map[variable.name.toLowerCase()] = variable.type;
		});
		return map;
	},

	getDatasets(state: DataState): Datasets[] {
		return state.datasets;
	},

	getAvailableVariablesMap(state: DataState, getters: any): Dictionary<boolean> {
		const available = getters.getAvailableVariables as string[];
		const map = {};
		available.forEach(name => {
			map[name] = true;
			map[name.toLowerCase()] = true;
		});
		return map;
	},

	getTrainingVariablesMap(state: DataState, getters: any): Dictionary<boolean> {
		const training = getters.getRouteTrainingVariables as string;
		if (training) {
			const map = {};
			training.split(',').forEach(name => {
				map[name.toLowerCase()] = true;
			});
			return map;
		}
		return {};
	},

	getAvailableVariables(state: DataState, getters: any): string[] {
		const training = getters.getTrainingVariablesMap as Dictionary<string>;
		const target = getters.getRouteTargetVariable as string;
		return state.variables.filter(variable => {
			return (!target || target.toLowerCase() !== variable.name.toLowerCase()) &&
				!training[variable.name.toLowerCase()];
		}).map(v => v.name);
	},

	getVariableSummaries(state: DataState): VariableSummary[] {
		return state.variableSummaries;
	},

	getResultsSummaries(state: DataState): VariableSummary[] {
		return state.resultsSummaries;
	},

	getResidualsSummaries(state: DataState): VariableSummary[] {
		return state.residualSummaries;
	},

	getSelectedFilters(state: DataState, getters: any): Filter[] {

		const existing = getters.getDecodedFilters as Filter[];
		const filters: Filter[] = [];

		// add training filters
		const training = getters.getRouteTrainingVariables as string;
		if (training) {
			training.split(',').forEach(variable => {
				const index = _.findIndex(existing, filter => {
					return filter.name == variable;
				});
				if (index === -1) {
					filters.push({
						name: variable,
						type: EMPTY_FILTER,
						enabled: false
					});
				} else {
					filters.push(existing[index]);
				}
			});
		}

		// add target filter
		const target = getters.getRouteTargetVariable as string;
		if (target) {
			const index = _.findIndex(existing, filter => {
				return filter.name == target;
			});
			if (index === -1) {
				filters.push({
					name: target,
					type: EMPTY_FILTER,
					enabled: false
				});
			} else {
				filters.push(existing[index]);
			}
		}

		return filters;
	},

	getAvailableVariableSummaries(state: DataState, getters: any): VariableSummary[] {
		const available = getters.getAvailableVariablesMap as Dictionary<Variable>;
		return state.variableSummaries.filter(variable => available[variable.name.toLowerCase()]);
	},

	getTrainingVariableSummaries(state: DataState, getters: any): VariableSummary[] {
		const training = getters.getTrainingVariablesMap as Dictionary<Variable>;
		return state.variableSummaries.filter(variable => training[variable.name.toLowerCase()]);
	},

	getTargetVariableSummaries(state: DataState, getters: any): VariableSummary[] {
		const target = getters.getRouteTargetVariable as string;
		if (!target) {
			return [];
		}
		return state.variableSummaries.filter(variable => {
			return target.toLowerCase() === variable.name.toLowerCase();
		});
	},

	getFilteredData(state: DataState): Data {
		return state.filteredData;
	},

	getFilteredDataNumRows(state: DataState): number {
		return state.filteredData ? state.filteredData.numRows : 0;
	},

	getFilteredDataItems(state: DataState, getters: any): Dictionary<any>[] {
		return getDataItems(state.filteredData, getters.getVariableTypesMap);
	},

	getFilteredDataFields(state: DataState): Dictionary<FieldInfo> {
		const data = state.filteredData;
		if (validateData(data)) {
			const variables = state.variables;
			const types = {};
			variables.forEach(variable => {
				types[variable.name] = variable.type;
			});
			const result: Dictionary<FieldInfo> = {} as any;
			for (const col of data.columns) {
				result[col] = {
					label: col,
					type: types[col],
					sortable: true
				};
			}
			return result;
		}
		return {};
	},

	getResultData(state: DataState): Data {
		return state.resultData;
	},

	getResultDataNumRows(state: DataState): number {
		return state.resultData ? state.resultData.numRows : 0;
	},

	getResultDataItems(state: DataState, getters: any): TargetRow[] {
		return getDataItems(state.resultData, getters.getVariableTypesMap) as TargetRow[];
	},

	getResultDataFields(state: DataState): Dictionary<FieldInfo> {
		const data = state.resultData;
		if (validateData(data)) {
			// look at first row and figure out the target, predicted, error values
			const predictedIndex = getPredictedIndex(data.columns);
			const targetIndex = getTargetIndex(data.columns);
			const errorIndex = getErrorIndex(data.columns);

			const result = {}
			// assign column names, ignoring target, predicted and error
			for (const [idx, col] of data.columns.entries()) {
				if (idx !== predictedIndex && idx !== targetIndex && idx !== errorIndex) {
					result[col] = {
						label: col,
						sortable: true,
						type: ""
					};
				}
			}
			// add target, predicted and error at end with customized labels
			const targetName = data.columns[targetIndex];
			result[targetName] = {
				label: targetName.replace(TARGET_POSTFIX, ''),
				sortable: true,
				type: ""
			};
			const predictedName = data.columns[predictedIndex];
			result[data.columns[predictedIndex]] = {
				label: `Predicted ${predictedName.replace(PREDICTED_POSTFIX, '')}`,
				sortable: true,
				type: ""
			};
			if (errorIndex >= 0) {
				result[data.columns[errorIndex]] = {
					label: 'Error',
					sortable: true,
					type: ""
				};
			}
			return result;
		}
		return {};
	},

	getSelectedData(state: DataState): Data {
		return state.selectedData;
	},

	getSelectedDataNumRows(state: DataState): number {
		return state.selectedData ? state.selectedData.numRows : 0;
	},

	getSelectedDataItems(state: DataState, getters: any): TableRow[] {
		return getDataItems(state.selectedData, getters.getVariableTypesMap);
	},

	getSelectedDataFields(state: DataState): Dictionary<FieldInfo> {
		const data = state.selectedData;
		if (validateData(data)) {
			const vmap = getters.getVariableTypesMap;
			const result = {};
			for (const col of data.columns) {
				result[col] = {
					label: col,
					type: vmap[col],
					sortable: true
				};
			}
			return result;
		}
		return {};
	},

	getHighlightedFeatureValues(state: DataState): Highlights {
		return state.highlightedFeatureValues;
	}
}
