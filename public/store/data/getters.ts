import _ from 'lodash';
import { FieldInfo, Variable, Data, DataState, Datasets, VariableSummary, TargetRow, TableRow, Extrema } from './index';
import { FilterParams, Filter } from '../../util/filters';
import { TARGET_POSTFIX, PREDICTED_POSTFIX, getTargetCol, getVarFromTarget, getPredictedCol, getErrorCol } from '../../util/data';
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

function getResultDataItems(data: Data, getters: any): TargetRow[] {
	if (!data ||
		!data.columns ||
		data.columns.length === 0) {
		return [];
	}

	// Find the target index and name in the result table
	const targetIndex = getTargetIndex(data.columns);
	const targetVarName = getVarFromTarget(data.columns[targetIndex]);

	// Make a copy of the variable type map and add entries for target, predicted and error
	// types.
	const resultVariableTypeMap = _.clone(<Dictionary<string>>getters.getVariableTypesMap);

	const targetVarType = resultVariableTypeMap[targetVarName];
	resultVariableTypeMap[getTargetCol(targetVarName)] = targetVarType;
	resultVariableTypeMap[getPredictedCol(targetVarName)] = targetVarType;
	resultVariableTypeMap[getErrorCol(targetVarName)] = targetVarType;

	// Fetch data items using modified type map
	return getDataItems(data, resultVariableTypeMap) as TargetRow[];
}

function getResultDataFields(data: Data): Dictionary<FieldInfo> {
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

	getTrainingVariables(state: DataState, getters: any): string[] {
		const training = getters.getTrainingVariablesMap as Dictionary<string>;
		return state.variables.filter(variable => {
			return training[variable.name.toLowerCase()];
		}).map(v => v.name);
	},

	getVariableSummaries(state: DataState): VariableSummary[] {
		return state.variableSummaries;
	},

	getResultSummaries(state: DataState): VariableSummary[] {
		return state.resultSummaries;
	},

	getPredictedSummaries(state: DataState): VariableSummary[] {
		return state.predictedSummaries;
	},

	getResidualsSummaries(state: DataState): VariableSummary[] {
		return state.residualSummaries;
	},

	getCorrectnessSummaries(state: DataState): VariableSummary[] {
		return state.correctnessSummaries;
	},

	getFilters(state: DataState, getters: any): Filter[] {
		const filterParams = getters.getDecodedFilterParams;
		return filterParams.filters.slice();
	},

	getSelectedFilterParams(state: DataState, getters: any): FilterParams {
		const filterParams = _.cloneDeep(getters.getDecodedFilterParams);
		// add training filters
		const training = getters.getRouteTrainingVariables as string;
		if (training) {
			filterParams.variables = filterParams.variables.concat(training.split(','));
		}
		// add target filter
		const target = getters.getRouteTargetVariable as string;
		if (target) {
			filterParams.variables.push(target);
		}
		return filterParams;
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

	getResultDataNumRows(state: DataState): number {
		return state.unhighlightedResultData ? state.unhighlightedResultData.numRows : 0;
	},

	hasHighlightedResultData(state: DataState): boolean {
		return !!state.highlightedResultData;
	},

	getHighlightedResultData(state: DataState): Data {
		return state.highlightedResultData;
	},

	getHighlightedResultDataItems(state: DataState, getters: any): TargetRow[] {
		return getResultDataItems(state.highlightedResultData, getters);
	},

	getHighlightedResultDataFields(state: DataState): Dictionary<FieldInfo> {
		return getResultDataFields(state.highlightedResultData);
	},

	hasUnhighlightedResultData(state: DataState): boolean {
		return !!state.unhighlightedResultData;
	},

	getUnhighlightedResultData(state: DataState): Data {
		return state.unhighlightedResultData;
	},

	getUnhighlightedResultDataItems(state: DataState, getters: any): TargetRow[] {
		return getResultDataItems(state.unhighlightedResultData, getters);
	},

	getUnhighlightedResultDataFields(state: DataState): Dictionary<FieldInfo> {
		return getResultDataFields(state.unhighlightedResultData);
	},

	hasSelectedData(state: DataState): boolean {
		return !!state.selectedData;
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

	getSelectedDataFields(state: DataState, getters: any): Dictionary<FieldInfo> {
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

	hasExcludedData(state: DataState): boolean {
		return !!state.excludedData;
	},

	getExcludedData(state: DataState): Data {
		return state.excludedData;
	},

	getExcludedDataNumRows(state: DataState): number {
		return state.excludedData ? state.excludedData.numRows : 0;
	},

	getExcludedDataItems(state: DataState, getters: any): TableRow[] {
		return getDataItems(state.excludedData, getters.getVariableTypesMap);
	},

	getExcludedDataFields(state: DataState, getters: any): Dictionary<FieldInfo> {
		const data = state.excludedData;
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

	getHighlightedSamples(state: DataState): Dictionary<string[]> {
		return state.highlightValues ? state.highlightValues.samples : {};
	},


	getHighlightedSummaries(state: DataState): VariableSummary[] {
		return state.highlightValues ? state.highlightValues.summaries : null;
	},

	getPredictedExtrema(state: DataState): Extrema {
		if (_.isEmpty(state.predictedExtremas)) {
			return {
				min: NaN,
				max: NaN
			};
		}
		const res = { min: Infinity, max: -Infinity };
		_.forIn(state.predictedExtremas, extrema => {
			res.min = Math.min(res.min, extrema.min);
			res.max = Math.max(res.max, extrema.max);
		});
		if (state.resultExtrema) {
			res.min = Math.min(res.min, state.resultExtrema.min);
			res.max = Math.max(res.max, state.resultExtrema.max);
		}
		return res;
	},

	getResidualExtrema(state: DataState): Extrema {
		if (_.isEmpty(state.residualExtremas)) {
			return {
				min: NaN,
				max: NaN
			};
		}
		const res = { min: Infinity, max: -Infinity };
		_.forIn(state.residualExtremas, extrema => {
			res.min = Math.min(res.min, extrema.min);
			res.max = Math.max(res.max, extrema.max);
		});
		return res;
	},

	getImages(state: DataState): Dictionary<any> {
		return state.loadedImages;
	}
}
