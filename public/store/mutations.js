export function setVariables(state, variables) {
	state.variables = variables;
}

export function setDatasets(state, datasets) {
	state.datasets = datasets;
}

export function setVariableSummaries(state, summaries) {
	state.variableSummaries = summaries;
}

export function updateVariableSummaries(state, args) {
	state.variableSummaries.splice(args.index, 1);
	state.variableSummaries.splice(args.index, 0, args.histogram);
}

export function setFilteredData(state, filteredData) {
	state.filteredData = filteredData;
}

export function setWebSocketConnection(state, connection) {
	state.wsConnection = connection;
}

export function addWebSocketStream(state, stream) {
	state.wsStreams[stream.id] = stream;
}

export function removeWebSocketStream(state, stream) {
	delete state.wsStreams[stream.id];
}
