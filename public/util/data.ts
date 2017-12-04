import _ from 'lodash';
import { DataState, Datasets, VariableSummary } from '../store/data/index';
import { Extrema } from '../store/data/index';
import { PipelineInfo } from '../store/pipelines/index';
import { DistilState } from '../store/store';
import { ActionContext } from 'vuex';
import axios from 'axios';
import Vue from 'vue';

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
			pipelineId: r.pipelineId
		};
	});
	setFunction(context, pendingHistograms);

	// fetch the results for each pipeline
	for (var result of results) {
		const name = nameFunc(result);
		const feature = result.feature;
		const pipelineId = result.pipelineId;
		const resultId = encodeURIComponent(result.pipeline.resultId);
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
							err: 'No analysis available'
						}
					]);
					return;
				}
				histogram.buckets = histogram.buckets ? histogram.buckets : [];
				histogram.name = name;
				histogram.feature = feature;
				histogram.pipelineId = pipelineId;
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
						err: error
					}
				]);
				return;
			});
	}
}
