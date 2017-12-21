import _ from 'lodash';
import { FieldInfo, Variable, Data, DataState, Datasets, VariableSummary, TargetRow, RangeHighlights, ValueHighlights } from './index';
import { Filter, EMPTY_FILTER } from '../../util/filters';
import { TARGET_POSTFIX, PREDICTED_POSTFIX } from '../../util/data';
import { Dictionary } from '../../util/dict';
import { getPredictedIndex, getErrorIndex, getTargetIndex } from '../../util/data';

function validateData(data: Data) {
	return !_.isEmpty(data) &&
		!_.isEmpty(data.values) &&
		!_.isEmpty(data.columns);
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
			// convert fetched result data rows into table data rows
			return _.map(resultData.values, resultRow => {
				const row: Dictionary<any> = {};

				for (const [idx, colValues] of resultRow.entries()) {
					const colName = resultData.columns[idx];
					row[colName] = colValues;
				}

				// display predicted error info
				return <TargetRow>row;
			});
		}
		return <TargetRow[]>[];
	},

	getResultDataFields(state: DataState): Dictionary<FieldInfo> {
		const data = state.resultData;
		if (validateData(data)) {
			// look at first row and figure out the target, predicted, error values
			const predictedIndex = getPredictedIndex(data.columns);
			const targetIndex = getTargetIndex(data.columns);
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
				label: targetName.replace(TARGET_POSTFIX, ''),
				sortable: true,
				type: "",
				suggested: {} as any,
			};
			const predictedName = data.columns[predictedIndex];
			result[data.columns[predictedIndex]] = {
				label: `Predicted ${predictedName.replace(PREDICTED_POSTFIX, '')}`,
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

	getHighlightedFeatureValues(state: DataState): ValueHighlights {
		return state.highlightedFeatureValues;
	},

	getHighlightedFeatureRanges(state: DataState): RangeHighlights {
		return state.highlightedFeatureRanges;
	}
}
