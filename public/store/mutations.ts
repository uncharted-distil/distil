import _ from 'lodash';
import Vue from 'vue';
import { MutationTree } from 'vuex';
import { DistilState } from './index';

export const mutations: MutationTree<DistilState> = {
	updateVariableType(state, update) {
		const index = _.findIndex(state.variables, elem => {
			return elem.name === update.field;
		});
		state.variables[index].type = update.type;
	},

	setVariables(state, variables) {
		state.variables = variables;
	},

	setDatasets(state, datasets) {
		state.datasets = datasets;
	},

	setVariableSummaries(state, summaries) {
		state.variableSummaries = summaries;
	},

	updateVariableSummaries(state, histogram) {
		const index = _.findIndex(state.variableSummaries, elem => {
			return elem.name === histogram.name;
		});
		Vue.set(state.variableSummaries, index, histogram);
	},

	setResultsSummaries(state, summaries) {
		state.resultsSummaries = summaries;
	},

	updateResultsSummaries(state, summary) {
		const index = _.findIndex(state.resultsSummaries, r => r.name === summary.name);
		if (index >=  0) {
		  Vue.set(state.resultsSummaries, index, summary);
		} else {
			state.resultsSummaries.push(summary);
		}
	},

	// sets the current filtered data into the store
	setFilteredData(state, filteredData) {
		state.filteredData = filteredData;
	},

	// sets the current selected data into the store
	setSelectedData(state, selectedData) {
		state.selectedData = selectedData;
	},

	// sets the current result data into the store
	setResultData(state, resultData) {
		state.resultData = resultData;
	},

	setWebSocketConnection(state, connection) {
		state.wsConnection = connection;
	},

	// sets the active session in the store as well as in the browser local storage
	setPipelineSession(state, session) {
		state.pipelineSession = session;
		if (!session) {
			window.localStorage.removeItem('pipeline-session-id');
		} else {
			window.localStorage.setItem('pipeline-session-id', session.id);
		}
	},

	// adds a running pipeline or replaces an existing one if the ids match
	addRunningPipeline(state, pipelineData) {
		if (!_.has(state.runningPipelines, pipelineData.requestId)) {
			Vue.set(state.runningPipelines, pipelineData.requestId, {});
		}
		Vue.set(state.runningPipelines[pipelineData.requestId], pipelineData.pipelineId, pipelineData);
	},

	// removes a running pipeline
	removeRunningPipeline(state, args) {
		if (_.has(state.runningPipelines, args.requestId)) {
			// delete the pipeline from the request
			if (_.has(state.runningPipelines[args.requestId], args.pipelineId)) {
				Vue.delete(state.runningPipelines[args.requestId], args.pipelineId);
				// delete the request if empty
				if (_.size(state.runningPipelines[args.requestId]) === 0) {
					Vue.delete(state.runningPipelines, args.requestId);
				}
				return true;
			}
		}
		return false;
	},

	// adds a completed pipeline or replaces an existing one if the ids match
	addCompletedPipeline(state, pipelineData) {
		if (!_.has(state.completedPipelines, pipelineData.requestId)) {
			Vue.set(state.completedPipelines, pipelineData.requestId, {});
		}
		Vue.set(state.completedPipelines[pipelineData.requestId], pipelineData.pipelineId, pipelineData);
	},

	// removes a completed pipeline
	removeCompletedPipeline(state, args) {
		if (_.has(state.runningPipelines, args.requestId)) {
			// delete the pipeline from the request
			if (_.has(state.completedPipelines[args.requestId], args.pipelineId)) {
				// delete the request if empty
				Vue.delete(state.completedPipelines[args.requestId], args.pipelineId);
				if (_.size(state.completedPipelines[args.requestId]) === 0) {
					Vue.delete(state.completedPipelines, args.requestId);
				}
				return true;
			}
		}
		return false;
	},

	highlightFeatureRange(state, highlight) {
		Vue.set(state.highlightedFeatureRanges, highlight.name, {
			from: highlight.from,
			to: highlight.to
		});
		if (state.resultsSummaries) {
			state.resultsSummaries.forEach(summary => {
				Vue.set(state.highlightedFeatureRanges, summary.feature, highlight);
			});
		}
	},

	clearFeatureHighlightRange(state, name) {
		Vue.delete(state.highlightedFeatureRanges, name);
	},

	highlightFeatureValues(state, highlights) {
		Vue.set(state, 'highlightedFeatureValues', highlights);
		if (state.resultsSummaries) {
			state.resultsSummaries.forEach(summary => {
				Vue.set(state.highlightedFeatureValues, summary.name, highlights[summary.feature]);
			});
		}
	},

	clearFeatureHighlightValues(state) {
		Vue.delete(state, 'highlightedFeatureValues');
	},

	addRecentDataset(state, dataset) {
		const datasetsStr = window.localStorage.getItem('recent-datasets');
		const datasets = (datasetsStr) ? datasetsStr.split(',') : [];
		datasets.unshift(dataset);
		window.localStorage.setItem('recent-datasets', datasets.join(','));
	}
};

