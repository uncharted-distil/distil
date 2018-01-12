import _ from 'lodash';
import { DataState, Datasets, VariableSummary } from '../store/data/index';
import { Extrema, TargetRow, FieldInfo } from '../store/data/index';
import { PipelineInfo } from '../store/pipelines/index';
import { DistilState } from '../store/store';
import { Dictionary } from './dict';
import { ActionContext } from 'vuex';
import axios from 'axios';
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
	const datasets = window.localStorage.getItem('recent-datasets');
	return (datasets) ? datasets.split(',') : [];
}

// adds a recent dataset to local storage
export function addRecentDataset(dataset: string) {
	const datasetsStr = window.localStorage.getItem('recent-datasets');
	const datasets = (datasetsStr) ? datasetsStr.split(',') : [];
	datasets.unshift(dataset);
	window.localStorage.setItem('recent-datasets', datasets.join(','));
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

export function getPredictedIndex(columns: string[]): number {
	return _.findIndex(columns, isPredicted);
}

export function getErrorIndex(columns: string[]): number {
	return _.findIndex(columns, isError);
}

export function getTargetIndex(columns: string[]): number {
	return _.findIndex(columns, isTarget);
}

export function getTargetCol(target: string): string {
	return target + TARGET_POSTFIX;
}

export function getPredictedCol(target: string): string {
	return target + PREDICTED_POSTFIX;
}

export function getErrorCol(target: string): string {
	return target + ERROR_POSTFIX;
}

export function getVarFromPredicted(decorated: string) {
	return decorated.replace(PREDICTED_POSTFIX, '');
}

export function getVarFromError(decorated: string) {
	return decorated.replace(ERROR_POSTFIX, '');
}

export function getVarFromTarget(decorated: string) {
	return decorated.replace(TARGET_POSTFIX, '');
}

export function getPredictedFacetKey(target: string) {
	return target + PREDICTED_FACET_KEY_POSTFIX;
}

export function getErrorFacetKey(target: string) {
	return target + ERROR_FACET_KEY_POSTFIX;
}

export function getErrorColFromFacetKey(facetKey: string) {
	return facetKey.replace(ERROR_FACET_KEY_POSTFIX, ERROR_POSTFIX);
}

export function getPredictedColFromFacetKey(facetKey: string) {
	return facetKey.replace(PREDICTED_FACET_KEY_POSTFIX, PREDICTED_POSTFIX);
}

export function getTargetColFromFacetKey(facetKey: string) {
	return facetKey + (TARGET_POSTFIX);
}

export function updateSummaries(summary: VariableSummary, summaries: VariableSummary[], matchField: string) {
	const index = _.findIndex(summaries, r => r[matchField] === summary[matchField]);
	if (index >= 0) {
		Vue.set(summaries, index, summary);
	} else {
		summaries.push(summary);
	}
}

export function getSummaries(context: DataContext, endpoint: string, results: PipelineInfo[], nameFunc: (PipelineInfo) => string,
	setFunction: (DataContext, VariableSummary) => void, updateFunction: (DataContext, VariableSummary) => void) {
	// save a placeholder histogram
	const pendingHistograms = _.map(results, r => {
		return {
			name: nameFunc(r),
			feature: '',
			pending: true,
			buckets: [],
			extrema: {} as any,
			pipelineId: r.pipelineId,
			resultId: ''
		};
	});
	setFunction(context, pendingHistograms);

	// fetch the results for each pipeline
	for (var result of results) {
		if (!result.pipeline) {
			// skip
			continue;
		}
		const name = nameFunc(result);
		const feature = result.feature;
		const pipelineId = result.pipelineId;
		const resultId = result.pipeline.resultId;
		axios.get(`${endpoint}/${resultId}`)
			.then(response => {
				// save the histogram data
				const histogram = response.data.histogram;
				if (!histogram) {
					setFunction(context, [
						{
							name: name,
							feature: feature,
							buckets: [],
							extrema: {} as Extrema,
							pipelineId: pipelineId,
							resultId: resultId,
							err: 'No analysis available'
						}
					]);
					return;
				}
				histogram.buckets = histogram.buckets ? histogram.buckets : [];
				histogram.name = name;
				histogram.feature = feature;
				histogram.pipelineId = pipelineId;
				histogram.resultId = resultId;
				updateFunction(context, histogram);
			})
			.catch(error => {
				setFunction(context, [
					{
						name: name,
						feature: feature,
						buckets: [],
						extrema: {} as Extrema,
						pipelineId: pipelineId,
						resultId: resultId,
						err: error
					}
				]);
				return;
			});
	}
}
