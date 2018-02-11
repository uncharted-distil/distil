import _ from 'lodash';
import { DataState, Datasets, VariableSummary, Data } from '../store/data/index';
import { TargetRow, FieldInfo } from '../store/data/index';
import { PipelineInfo, PIPELINE_UPDATED, PIPELINE_COMPLETED } from '../store/pipelines/index';
import { DistilState } from '../store/store';
import { Dictionary } from './dict';
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

export function createPendingSummary(name: string, label: string, dataset: string, pipelineId?: string): VariableSummary {
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
		pipelineId: pipelineId
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
	pipeline: PipelineInfo,
	nameFunc: (PipelineInfo) => string,
	labelFunc: (PipelineInfo) => string,
	updateFunction: (DataContext, VariableSummary) => void): Promise<any> {

	const name = nameFunc(pipeline);
	const label = labelFunc(pipeline);
	const feature = pipeline.feature;
	const dataset = pipeline.dataset;
	const pipelineId = pipeline.pipelineId;
	const resultId = pipeline.resultId;

	// save a placeholder histogram
	updateFunction(context, createPendingSummary(name, label, dataset, pipelineId));

	// fetch the results for each pipeline
	if (pipeline.progress !== PIPELINE_UPDATED &&
		pipeline.progress !== PIPELINE_COMPLETED) {
		// skip
		return;
	}

	// return promise
	return axios.get(`${endpoint}/${resultId}`)
		.then(response => {
			// save the histogram data
			const histogram = response.data.histogram;
			histogram.name = name;
			histogram.label = label;
			histogram.feature = feature;
			histogram.pipelineId = pipelineId;
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
	pipelines: PipelineInfo[],
	nameFunc: (PipelineInfo) => string,
	labelFunc: (PipelineInfo) => string,
	updateFunction: (DataContext, VariableSummary) => void): Promise<any> {

	// return as singular promise
	const promises = pipelines.map(pipeline => {
		return getSummary(
			context,
			endpoint,
			pipeline,
			nameFunc,
			labelFunc,
			updateFunction);
	});
	return Promise.all(promises);
}
