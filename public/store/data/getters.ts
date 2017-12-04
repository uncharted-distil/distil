import _ from 'lodash';
import { Variable, Data, DataState, Dictionary, Datasets, VariableSummary, TargetRow } from './index';
import { FilterMap } from '../../util/filters';
import { Range } from './index';

function getTargetIndexFromPredicted(columns: string[], predictedIndex: number) {
	const targetName = columns[predictedIndex].replace('_res', '');
	return _.findIndex(columns, col => col.toLowerCase() === targetName.toLowerCase());
}

function getPredictedIndex(columns: string[]) {
	return _.findIndex(columns, col => col.endsWith('_res'));
}

function getErrorIndex(columns: string[]) {
	return _.findIndex(columns, col => col === 'error');
}

function validateData(data: Data) {
	return !_.isEmpty(data) &&
		!_.isEmpty(data.values) &&
		!_.isEmpty(data.columns);
}

export interface FieldInfo {
	label: string,
	type: string,
	suggested: Dictionary<string>,
	sortable: boolean
}

export const getters = {
	getVariables(state: DataState): Variable[] {
		return state.variables;
	},

	getVariablesMap(state: DataState): { [name: string]: Variable } {
		const map: { [name: string]: Variable } = {};
		state.variables.forEach(variable => {
			map[variable.name.toLowerCase()] = variable;
		});
		return map;
	},

	getDatasets(state: DataState): Datasets[] {
		return state.datasets;
	},

	getAvailableVariablesMap(state: DataState, getters: any): Dictionary<boolean> {
		const available = getters.getAvailableVariables as string[];
		const map: { [name: string]: boolean } = {};
		available.forEach(name => {
			map[name.toLowerCase()] = true;
		});
		return map;
	},

	getTrainingVariablesMap(state: DataState, getters: any): Dictionary<boolean> {
		const training = getters.getRouteTrainingVariables as string;
		if (training) {
			const map: Dictionary<boolean> = {};
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

	getSelectedFilters(state: DataState, getters: any): FilterMap {
		const training = getters.getRouteTrainingVariables as string;
		if (training) {
			const existing = getters.getDecodedFilters as FilterMap;
			const filters: FilterMap = {};

			training.split(',').forEach(variable => {
				if (!existing[variable]) {
					filters[variable] = {
						name: variable,
						enabled: false
					};
				} else {
					filters[variable] = existing[variable];
				}
			});
			return filters;
		}
		return {};
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

	getFilteredDataItems(state: DataState): Dictionary<any>[] {
		if (validateData(state.filteredData)) {
			return _.map(state.filteredData.values, d => {
				const row: { [col: string]: any } = {};
				for (const [index, col] of state.filteredData.columns.entries()) {
					row[col] = d[index];
				}
				return row;
			});
		}
		return [];
	},

	getFilteredDataFields(state: DataState): Dictionary<FieldInfo> {
		const data = state.filteredData;
		if (validateData(data)) {
			const variables = state.variables;
			const types = {};
			const suggested = {};
			variables.forEach(variable => {
				types[variable.name] = variable.type;
				suggested[variable.name] = variable.suggestedTypes;
			});

			const result: Dictionary<FieldInfo> = {} as any;
			for (const col of data.columns) {
				result[col] = {
					label: col,
					type: types[col],
					suggested: suggested[col],
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

	getResultDataItems(state: DataState): TargetRow[] {
		const resultData = state.resultData;
		if (validateData(resultData)) {

			// look at first row and figure out the target, predicted, error values
			const predictedIdx = getPredictedIndex(resultData.columns);
			const targetName = resultData.columns[getTargetIndexFromPredicted(resultData.columns, predictedIdx)];
			const errorIdx = getErrorIndex(resultData.columns);

			// convert fetched result data rows into table data rows
			return _.map(resultData.values, resultRow => {
				const row: Dictionary<any> = {};

				for (const [idx, colValues] of resultRow.entries()) {
					const colName = resultData.columns[idx];
					row[colName] = colValues;
				}

				// display predicted error info
				const targetRow = <TargetRow>row;
				targetRow._target = { truth: targetName, predicted: resultData.columns[predictedIdx] };
				if (errorIdx >= 0) {
					targetRow._target.error = resultData.columns[errorIdx];
				}
				return targetRow;
			});
		}
		return <TargetRow[]>[];
	},

	getResultDataFields(state: DataState): Dictionary<FieldInfo> {
		const data = state.resultData;
		if (validateData(data)) {
			// look at first row and figure out the target, predicted, error values
			const predictedIndex = getPredictedIndex(data.columns);
			const targetIndex = getTargetIndexFromPredicted(data.columns, predictedIndex);
			const errorIndex = getErrorIndex(data.columns);

			const result: Dictionary<FieldInfo> = {} as any
			// assign column names, ignoring target, predicted and error
			for (const [idx, col] of data.columns.entries()) {
				if (idx !== predictedIndex && idx !== targetIndex && idx !== errorIndex) {
					result[col] = {
						label: col,
						sortable: true,
						type: "",
						suggested: {} as any
					};
				}
			}
			// add target, predicted and error at end with customized labels
			const targetName = data.columns[targetIndex];
			result[targetName] = {
				label: targetName,
				sortable: true,
				type: "",
				suggested: {} as any,
			};
			result[data.columns[predictedIndex]] = {
				label: `Predicted ${targetName}`,
				sortable: true,
				type: "",
				suggested: {} as any
			};
			if (errorIndex >= 0) {
				result[data.columns[errorIndex]] = {
					label: 'Error',
					sortable: true,
					type: "",
					suggested: {} as any
				};
			}
			return result;
		}
		return {} as Dictionary<FieldInfo>;
	},

	getSelectedData(state: DataState): Data {
		return state.selectedData;
	},

	getSelectedDataItems(state: DataState): Dictionary<any>[] {
		if (validateData(state.selectedData)) {
			return _.map(state.selectedData.values, d => {
				const row: { [col: string]: any } = {};
				for (const [index, col] of state.selectedData.columns.entries()) {
					row[col] = d[index];
				}
				return row;
			});
		}
		return [];
	},

	getSelectedDataFields(state: DataState): Dictionary<FieldInfo> {
		const data = state.selectedData;
		if (validateData(data)) {
			const variables = state.variables;
			const types = {};
			const suggested: {} = [];
			variables.forEach(variable => {
				types[variable.name] = variable.type;
				suggested[variable.name] = variable.suggestedTypes;
			});

			const result: { [label: string]: FieldInfo } = {};
			for (const col of data.columns) {
				result[col] = {
					label: col,
					type: types[col],
					suggested: suggested[col],
					sortable: true
				};
			}
			return result;
		}
		return {} as Dictionary<FieldInfo>;
	},

	getHighlightedFeatureValues(state: DataState): Dictionary<any> {
		return state.highlightedFeatureValues;
	},

	getHighlightedFeatureRanges(state: DataState): Range {
		return state.highlightedFeatureRanges;
	}
}
