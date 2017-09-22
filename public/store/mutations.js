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

export function setResultsSummaries(state, summaries) {
	state.resultsSummaries = summaries;
}

export function updateResultsSummaries(state, summary) {
	const idx = _.findIndex(state.resultsSummaries, r => r.name === summary.name);
	if (idx >=  0) {
		state.resultsSummaries.splice(idx, 1, summary);
	} else {
		state.resultsSummaries.push(summary);
	}
}

function validateData(data) {
	return  !_.isEmpty(data) &&
		!_.isEmpty(data.values) &&
		!_.isEmpty(data.columns);
}

// sets the current filtered data into the store
export function setFilteredData(state, filteredData) {
	state.filteredData = filteredData;
	state.filteredDataItems = [];
	if (validateData(filteredData)) {
		state.filteredDataItems = _.map(filteredData.values, d => {
			const row = {};
			for (const [index, col] of filteredData.columns.entries()) {
				row[col] = d[index];
			}
			return row;
		});
	}
}

// sets the current result data into the store
export function setResultData(state, resultData) {
	state.resultData = resultData;
	state.resultDataItems = [];
	if (validateData(resultData)) {
		state.resultDataItems = _.cloneDeep(state.filteredDataItems);
		// append the result variable data to the baseline variable data
		for (const [i, row] of state.resultDataItems.entries()) {
			for (const [j, colName] of resultData.columns.entries()) {
				const label = `Predicted ${colName}`;
				row[label] = resultData.values[i][j];
				row._target = { truth: colName, predicted: label };
			}
		}
	}
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
	if (!_.has(state.runningPipelines, pipelineData.requestId)) {
		Vue.set(state.runningPipelines, pipelineData.requestId, {});
	}
	Vue.set(state.runningPipelines[pipelineData.requestId], pipelineData.pipelineId, pipelineData);
}

// removes a running pipeline
export function removeRunningPipeline(state, args) {
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
}

// adds a completed pipeline or replaces an existing one if the ids match
export function addCompletedPipeline(state, pipelineData) {
	if (!_.has(state.completedPipelines, pipelineData.requestId)) {
		Vue.set(state.completedPipelines, pipelineData.requestId, {});
	}
	Vue.set(state.completedPipelines[pipelineData.requestId], pipelineData.pipelineId, pipelineData);
}

// removes a completed pipeline
export function removeCompletedPipeline(state, args) {
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
}

export function highlightFeature(state, highlight) {
	state.highlightedFeature = highlight;
}

export function clearFeatureHighlight(state) {
	state.highlightedFeature = null;
}

function isHighlighted(highlightedFeature, row) {
	if (!highlightedFeature) {
		return false;
	}
	return row[highlightedFeature.name] >= highlightedFeature.range.from &&
		row[highlightedFeature.name] <= highlightedFeature.range.to;
}

export function highlightFilteredDataItems(state) {
	const items = state.filteredDataItems;
	const highlightedFeature = state.highlightedFeature;
	items.forEach(item => {
		if (isHighlighted(highlightedFeature, item)) {
			Vue.set(item, '_rowVariant', 'info');
		} else {
			Vue.set(item, '_rowVariant', undefined);
		}
	});
}

export function highlightResultdDataItems(state) {
	const items = state.resultDataItems;
	const highlightedFeature = state.highlightedFeature;
	items.forEach(item => {
		if (isHighlighted(highlightedFeature, item)) {
			Vue.set(item, '_rowVariant', 'info');
		} else {
			Vue.set(item, '_rowVariant', undefined);
		}
	});
}
