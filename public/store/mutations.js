import _ from 'lodash';
import Vue from 'vue';

export function setVariables(state, variables) {
	state.variables = variables;
}

export function setDatasets(state, datasets) {
	state.datasets = datasets;
}

export function setVariableSummaries(state, summaries) {
	state.variableSummaries = summaries;
	state.trainingVariables = {};
}

export function updateVariableSummaries(state, args) {
	state.variableSummaries.splice(args.index, 1);
	state.variableSummaries.splice(args.index, 0, args.histogram);
}

// sets the current filtered data into the store
export function setFilteredData(state, filteredData) {
	state.filteredData = filteredData;
}

export function setWebSocketConnection(state, connection) {
	state.wsConnection = connection;
}

// sets the active session in the store as well as in the browser local storage
export function setPipelineSession(state, session) {
	state.pipelineSession = session;
	if (!session) {
		window.localStorage.removeItem('pipeline-session-id');
	} else {
		window.localStorage.setItem('pipeline-session-id', session.id);
	}
}

// adds a running pipeline or replaces an existing one if the ids match
export function addRunningPipeline(state, pipelineData) {
	Vue.set(state.runningPipelines, pipelineData.pipelineId, pipelineData);
}

// removes a running pipeline
export function removeRunningPipeline(state, pipelineId) {
	if (_.has(state.runningPipelines, pipelineId)) {
		Vue.delete(state.runningPipelines, pipelineId);
		return true;
	}
	return false;
}

// adds a completed pipeline or replaces an existing one if the ids match
export function addCompletedPipeline(state, pipelineData) {
	Vue.set(state.completedPipelines, pipelineData.pipelineId, pipelineData);
}

// removes a completed pipeline
export function removeCompletedPipeline(state, pipelineId) {
	if (_.has(state.completedPipelines, pipelineId)) {
		Vue.delete(state.completedPipelines, pipelineId);
		return true;
	}
	return false;
}
