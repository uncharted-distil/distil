import _ from 'lodash';
import axios from 'axios';
import Vue from 'vue';
import localStorage from 'store';
import { Dataset, VariableSummary, SummaryType, TableData, TableRow } from '../store/dataset/index';
import { TargetRow, TableColumn, Variable } from '../store/dataset/index';
import { Solution, SOLUTION_COMPLETED } from '../store/solutions/index';
import { Dictionary } from './dict';
import { mutations as resultMutations } from '../store/results/module';
import { mutations as highlightMutations } from '../store/highlights/module';
import { Group } from './facets';
import { FilterParams } from './filters';
import { formatValue } from '../util/types';

// Postfixes for special variable names
export const PREDICTED_POSTFIX = '_predicted';
export const TARGET_POSTFIX = '_target';
export const ERROR_POSTFIX = '_error';
export const CORRECTNESS_POSTFIX = '_correctness';

export const PREDICTED_FACET_KEY_POSTFIX = ' - predicted';
export const ERROR_FACET_KEY_POSTFIX = ' - error';

export const NUM_PER_PAGE = 10;

// filters datasets by id
export function filterDatasets(ids: string[], datasets: Dataset[]): Dataset[] {
	if (_.isUndefined(ids)) {
		return datasets;
	}
	const idSet = new Set(ids);
	return _.filter(datasets, d => idSet.has(d.name));
}

// fetches datasets from local storage
export function getRecentDatasets(): string[] {
	return localStorage.get('recent-datasets') || [];
}

// adds a recent dataset to local storage
export function addRecentDataset(dataset: string) {
	const datasets = getRecentDatasets();
	if (datasets.indexOf(dataset) === -1) {
		datasets.unshift(dataset);
		localStorage.set('recent-datasets', datasets);
	}
}

export function isInTrainingSet(col: string, training: Dictionary<boolean>) {
	return (isPredicted(col) ||
		isError(col) ||
		isTarget(col) ||
		isHiddenField(col) ||
		training[col]);
}

export function removeNonTrainingItems(items: TargetRow[], training: Dictionary<boolean>):  TargetRow[] {
	return _.map(items, item => {
		const row: TargetRow = <TargetRow>{};
		_.forIn(item, (val, col) => {
			if (isInTrainingSet(col, training)) {
				row[col] = val;
			}
		});
		return row;
	});
}

export function removeNonTrainingFields(fields: Dictionary<TableColumn>, training: Dictionary<boolean>): Dictionary<TableColumn> {
	const res: Dictionary<TableColumn> = {};
	_.forIn(fields, (val, col) => {
		if (isInTrainingSet(col, training)) {
			res[col] = val;
		}
	});
	return res;
}

// Identifies column names as one of the special result types.
// Examples: weight_predicted, weight_error, weight_target

export function isPredicted(col: string): boolean {
	return col.endsWith(PREDICTED_POSTFIX);
}

export function isError(col: string): boolean {
	return col.endsWith(ERROR_POSTFIX);
}

export function isTarget(col: string): boolean {
	return col.endsWith(TARGET_POSTFIX);
}

export function isCorrectness(col: string): boolean {
	return col.endsWith(CORRECTNESS_POSTFIX);
}

export function isHiddenField(col: string): boolean {
	return col.startsWith('_');
}

// Finds the index of a server-side column.

export function getPredictedIndex(columns: string[]): number {
	return _.findIndex(columns, isPredicted);
}

export function getErrorIndex(columns: string[]): number {
	return _.findIndex(columns, isError);
}

export function getTargetIndex(columns: string[]): number {
	return _.findIndex(columns, isTarget);
}

export function getCorrectnessIndex(columns: string[]): number {
	return _.findIndex(columns, isCorrectness);
}

// Converts from variable name to a server-side result column name
// Example: "weight" -> "weight_predicted"

export function getTargetCol(target: string): string {
	return target + TARGET_POSTFIX;
}

export function getPredictedCol(target: string): string {
	return target + PREDICTED_POSTFIX;
}

export function getErrorCol(target: string): string {
	return target + ERROR_POSTFIX;
}

export function getCorrectnessCol(target: string): string {
	return target + CORRECTNESS_POSTFIX;
}

// Converts from a server side result column name to a variable name
// Example: "weight_error" -> "error"

export function getVarFromPredicted(decorated: string) {
	return decorated.replace(PREDICTED_POSTFIX, '');
}

export function getVarFromError(decorated: string) {
	return decorated.replace(ERROR_POSTFIX, '');
}

export function getVarFromTarget(decorated: string) {
	return decorated.replace(TARGET_POSTFIX, '');
}

export function getVarFromCorrectness(decorated: string) {
	return decorated.replace(CORRECTNESS_POSTFIX, '');
}

export function updateSummaries(summary: VariableSummary, summaries: VariableSummary[], matchField: string) {
	// TODO: add and check timestamps to ensure we don't overwrite old data?
	const index = _.findIndex(summaries, r => r[matchField] === summary[matchField]);
	if (index >= 0) {
		Vue.set(summaries, index, summary);
	} else {
		summaries.push(summary);
	}
}

export function filterSummariesByDataset(summaries: VariableSummary[], dataset: string): VariableSummary[] {
	return summaries.filter(summary => {
		return summary.dataset === dataset;
	});
}

export function createEmptyTableData(name: string): TableData {
	return {
		name: name,
		numRows: 0,
		columns: [],
		types: [],
		values: []
	};
}

export function createPendingSummary(name: string, label: string, dataset: string, solutionId?: string): VariableSummary {
	return {
		name: name,
		label: label,
		dataset: dataset,
		feature: '',
		pending: true,
		buckets: [],
		extrema: {
			min: NaN,
			max: NaN
		},
		numRows: 0,
		solutionId: solutionId
	};
}

export function createErrorSummary(name: string, label: string, dataset: string, error: any): VariableSummary {
	return {
		name: name,
		label: label,
		dataset: dataset,
		feature: '',
		buckets: [],
		extrema: {
			min: NaN,
			max: NaN
		},
		err: error.response ? error.response.data : error,
		numRows: 0
	};
}

export function getSummary(
	context: any,
	endpoint: string,
	solution: Solution,
	nameFunc: (Solution) => string,
	labelFunc: (Solution) => string,
	updateFunction: (any, VariableSummary) => void,
	filterParams: FilterParams): Promise<any> {

	const name = nameFunc(solution);
	const label = labelFunc(solution);
	const feature = solution.feature;
	const dataset = solution.dataset;
	const solutionId = solution.solutionId;
	const resultId = solution.resultId;

	// save a placeholder histogram
	updateFunction(context, createPendingSummary(name, label, dataset, solutionId));

	// fetch the results for each solution
	if (solution.progress !== SOLUTION_COMPLETED) {
		// skip
		return;
	}

	// return promise
	return axios.post(`${endpoint}/${resultId}`, filterParams ? filterParams: {})
		.then(response => {
			// save the histogram data
			const histogram = response.data.histogram;
			histogram.name = name;
			histogram.label = label;
			histogram.feature = feature;
			histogram.solutionId = solutionId;
			histogram.resultId = resultId;
			updateFunction(context, histogram);
		})
		.catch(error => {
			console.error(error);
			updateFunction(context, createErrorSummary(name, label, dataset, error));
		});
}

export function getSummaries(
	context: any,
	endpoint: string,
	solutions: Solution[],
	nameFunc: (Solution) => string,
	labelFunc: (Solution) => string,
	updateFunction: (any, VariableSummary) => void,
	filterParams: FilterParams): Promise<any> {

	// return as singular promise
	const promises = solutions.map(solution => {
		return getSummary(
			context,
			endpoint,
			solution,
			nameFunc,
			labelFunc,
			updateFunction,
			filterParams);
	});
	return Promise.all(promises);
}

export function filterVariablesByPage(pageIndex: number, numPerPage: number, variables: any[]) {
	if (variables.length > numPerPage) {
		const firstIndex = numPerPage * (pageIndex - 1);
		const lastIndex = Math.min(firstIndex + numPerPage, variables.length);
		return variables.slice(firstIndex, lastIndex);
	}
	return variables;
}

export function sortVariablesByImportance(variables: Variable[]): Variable[] {
	variables.sort((a, b) => {
		return b.importance - a.importance;
	});
	return variables;
}

export function sortGroupsByImportance(groups: Group[], variables: Variable[]): Group[] {
	// create importance lookup map
	const importance: Dictionary<number> = {};
	variables.forEach(variable => {
		importance[variable.name] = variable.importance;
	});
	// sort by importance
	groups.sort((a, b) => {
		return importance[b.key] - importance[a.key];
	});
	return groups;
}

export function updateCorrectnessHighlightSummary(context: any, summary: VariableSummary) {
	mutateCorrectnessSummary(context, summary, highlightMutations.updateCorrectnessHighlightSummaries)
}

export function updateCorrectnessSummary(context: any, summary: VariableSummary) {
	mutateCorrectnessSummary(context, summary, resultMutations.updateCorrectnessSummaries)
}

// Collapse categorical result summary data, which is returned as a confusion matrix, into a binary
// correct/incorrect reprsentation prior to applying the mutation.
function mutateCorrectnessSummary(context: any, summary: VariableSummary, f: (any, VariableSummary) => void) {
	// Only need to collapse categorical result summaries
	if (summary.type !== SummaryType.Categorical) {
		f(context, summary);
		return;
	}

	// total up correct and incorrect
	let correct = 0;
	let incorrect = 0;
	for (const bucket of summary.buckets) {
		for (const subBucket of bucket.buckets) {
			if (subBucket.key === bucket.key) {
				correct += subBucket.count;
			} else {
				incorrect += subBucket.count;
			}
		}
	}
	// create a new summary, replacing the buckets with the collapsed values
	const clonedSummary = _.cloneDeep(summary);
	clonedSummary.buckets = [
		{ key: "Correct", count: correct },
		{ key: "Incorrect", count: incorrect}
	]

	f(context, clonedSummary);
}

export function validateData(data: TableData) {
	return !_.isEmpty(data) &&
		!_.isEmpty(data.values) &&
		!_.isEmpty(data.columns);
}

export function getTableDataItems(data: TableData, typeMap: Dictionary<string>): TableRow[] {
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

export function getResultDataItems(data: TableData, getters: any): TargetRow[] {
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
	return getTableDataItems(data, resultVariableTypeMap) as TargetRow[];
}

export function getResultDataFields(data: TableData): Dictionary<TableColumn> {
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
