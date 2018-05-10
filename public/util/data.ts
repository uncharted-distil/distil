import _ from 'lodash';
import { DataState, Datasets, VariableSummary, Data, SummaryType } from '../store/data/index';
import { TargetRow, FieldInfo, Variable } from '../store/data/index';
import { SolutionInfo, SOLUTION_COMPLETED } from '../store/solutions/index';
import { DistilState } from '../store/store';
import { Dictionary } from './dict';
import { mutations as dataMutations } from '../store/data/module';
import { Group } from './facets';
import { FilterParams } from './filters';
import { ActionContext } from 'vuex';
import axios from 'axios';
import localStorage from 'store';
import Vue from 'vue';

// Postfixes for special variable names
export const PREDICTED_POSTFIX = '_predicted';
export const TARGET_POSTFIX = '_target';
export const ERROR_POSTFIX = '_error';

export const PREDICTED_FACET_KEY_POSTFIX = ' - predicted';
export const ERROR_FACET_KEY_POSTFIX = ' - error';

export const NUM_PER_PAGE = 10;

export type DataContext = ActionContext<DataState, DistilState>;

// filters datasets by id
export function filterDatasets(ids: string[], datasets: Datasets[]): Datasets[] {
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
			if (isInTrainingSet(col.toLowerCase(), training)) {
				row[col] = val;
			}
		});
		return row;
	});
}

export function removeNonTrainingFields(fields: Dictionary<FieldInfo>, training: Dictionary<boolean>): Dictionary<FieldInfo> {
	const res: Dictionary<FieldInfo> = {};
	_.forIn(fields, (val, col) => {
		if (isInTrainingSet(col.toLowerCase(), training)) {
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

export function createEmptyData(name: string): Data {
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
	context: DataContext,
	endpoint: string,
	solution: SolutionInfo,
	nameFunc: (SolutionInfo) => string,
	labelFunc: (SolutionInfo) => string,
	updateFunction: (DataContext, VariableSummary) => void,
	filters: FilterParams): Promise<any> {

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
	return axios.post(`${endpoint}/${resultId}`, filters ? filters: {})
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
	context: DataContext,
	endpoint: string,
	solutions: SolutionInfo[],
	nameFunc: (SolutionInfo) => string,
	labelFunc: (SolutionInfo) => string,
	updateFunction: (DataContext, VariableSummary) => void,
	filters: FilterParams): Promise<any> {

	// return as singular promise
	const promises = solutions.map(solution => {
		return getSummary(
			context,
			endpoint,
			solution,
			nameFunc,
			labelFunc,
			updateFunction,
			filters);
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


export function updatePredictedHighlightSummary(context: DataContext, summary: VariableSummary) {
	mutatePredictedSummary(context, summary, dataMutations.updatePredictedHighlightSummaries)
}

export function updatePredictedSummary(context: DataContext, summary: VariableSummary) {
	mutatePredictedSummary(context, summary, dataMutations.updatePredictedSummaries)
}

// Collapse categorical result summary data, which is returned as a confusion matrix, into a binary
// correct/incorrect reprsenation prior to applying the mutation.
function mutatePredictedSummary(context: DataContext, summary: VariableSummary, f: (DataContext, VariableSummary) => void) {
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
