import _ from 'lodash';
import axios from 'axios';
import {
	decodeFilters,
	encodeQueryParams
} from '../util/filters';

// TODO: move this somewhere more appropriate.
const ES_INDEX = 'datasets';
const CREATE_PIPELINES_MSG = 'CREATE_PIPELINES';
const PIPELINE_COMPLETE = 'COMPLETED';
const STREAM_CLOSE = 'STREAM_CLOSE';


// searches dataset descriptions and column names for supplied terms
export function searchDatasets(context, terms) {
	const params = !_.isEmpty(terms) ? `?search=${terms}` : '';
	return axios.get(`/distil/datasets/${ES_INDEX}${params}`)
		.then(response => {
			context.commit('setDatasets', response.data.datasets);
		})
		.catch(error => {
			console.error(error);
			context.commit('setDatasets', []);
		});
}

// searches dataset descriptions and column names for supplied terms
export function getSession(context) {
	const sessionID = context.getters.getPipelineSessionID();
	return axios.get(`/distil/session/${sessionID}`)
		.then(response => {
			if (response.data.pipelines) {
				response.data.pipelines.forEach((pipeline) => {
					pipeline.Results.forEach((res) => {
						// add/update the running pipeline info
						if (pipeline.Progress === PIPELINE_COMPLETE) {
							//move the pipeline from running to complete
							context.commit('addCompletedPipeline', {
								name: res.name,
								requestId: pipeline.requestId,
								pipelineId: pipeline.pipelineId,
								pipeline: res
							});
						}
					});
				});
			}
		})
		.catch(error => {
			console.error(error);
		});
}

// fetches all variables for a single dataset.
export function getVariables(context, dataset) {
	return axios.get(`/distil/variables/${ES_INDEX}/${dataset}`)
		.then(response => {
			context.commit('setVariables', response.data.variables);
		})
		.catch(error => {
			console.error(error);
			context.commit('setVariables', []);
		});
}

// fetches variable summary data for the given dataset and variables
export function getVariableSummaries(context, datasetName) {
	return context.dispatch('getVariables', datasetName)
		.then(() => {
			const variables = context.getters.getVariables();
			// commit empty place holders
			const histograms = variables.map(variable => {
				return {
					name: variable.name,
					pending: true
				};
			});
			context.commit('setVariableSummaries', histograms);
			// fill them in asynchronously
			variables.forEach((variable, idx) => {
				axios.get(`/distil/variable-summaries/${ES_INDEX}/${datasetName}/${variable.name}`)
					.then(response => {
						// save the variable summary data
						const histogram = response.data.histogram;
						if (!histogram) {
							context.commit('updateVariableSummaries', {
								index: idx,
								histogram: {
									name: variable.name,
									err: 'No analysis available'
								}
							});
							return;
						}
						// ensure buckets is not nil
						//histogram.buckets = histogram.buckets ? histogram.buckets : [];
						context.commit('updateVariableSummaries', {
							index: idx,
							histogram: histogram
						});
					})
					.catch(error => {
						console.error(error);
						context.commit('updateVariableSummaries', {
							index: idx,
							histogram: {
								name: variable.name,
								err: error
							}
						});
					});
			});
		})
		.catch(error => {
			console.error(error);
		});
}

// update filtered data based on the  current filter state
export function updateFilteredData(context, datasetName) {
	const filters = context.getters.getRouteFilters();
	const decoded = decodeFilters(filters);
	const queryParams = encodeQueryParams(decoded);
	const url = `distil/filtered-data/${ES_INDEX}/${datasetName}${queryParams}`;
	// request filtered data from server - no data is valid given filter settings
	return axios.get(url)
		.then(response => {
			context.commit('setFilteredData', response.data);
		})
		.catch(error => {
			console.error(error);
			context.commit('setFilteredData', []);
		});
}

// starts a pipeline session.
export function getPipelineSession(context) {
	const conn = context.getters.getWebSocketConnection();
	const sessionID = context.getters.getPipelineSessionID();
	return conn.send({
		type: 'GET_SESSION',
		session: sessionID
	}).then(res => {
		if (sessionID && res.created) {
			console.warn('previous session', sessionID, 'could not be resumed, new session created');
		}
		context.commit('setPipelineSession', {
			id: res.session,
			uuids: res.uuids
		});
	}).catch(err => {
		console.warn(err);
	});
}

// end a pipeline session.
export function endPipelineSession(context) {
	const conn = context.getters.getWebSocketConnection();
	const sessionID = context.getters.getPipelineSessionID();
	if (!sessionID) {
		return;
	}
	return conn.send({
		type: 'END_SESSION',
		session: sessionID
	}).then(() => {
		context.commit('setPipelineSession', null);
	}).catch(err => {
		console.warn(err);
	});
}

// issues a pipeline create request
export function createPipelines(context, request) {

	const conn = context.getters.getWebSocketConnection();
	const sessionID = context.getters.getPipelineSessionID();
	if (!sessionID) {
		return;
	}
	const stream = conn.stream(res => {

		if (_.has(res, STREAM_CLOSE)) {
			stream.close();
			return;
		}

		// inject the name
		const name = `${context.getters.getRouteDataset()}-${request.feature}-${res.pipelineId.substring(0,8)}`;
		res.name = name;

		// add/update the running pipeline info
		context.commit('addRunningPipeline', res);
		if (res.progress === PIPELINE_COMPLETE) {
			//move the pipeline from running to complete
			context.commit('removeRunningPipeline', {pipelineId: res.pipelineId, requestId: res.requestId});
			context.commit('addCompletedPipeline', {
				name: res.name,
				requestId: res.requestId,
				pipelineId: res.pipelineId,
				pipeline: res.pipeline
			});
		}
	});

	stream.send({
		type: CREATE_PIPELINES_MSG,
		session: sessionID,
		index: ES_INDEX,
		dataset: context.getters.getRouteDataset(),
		feature: request.feature,
		task: request.task,
		metric: request.metric,
		output: request.output,
		maxPipelines: 3,
		filters: decodeFilters(context.getters.getRouteFilters())
	});
}

// fetches results for a given pipeline create request
export function getResultsSummaries(context, args) {
	const results = context.getters.getPipelineResults(args.requestId);

	// save a placeholder histogram
	const pendingHistograms = _.map(results, r => {
		return {
			name: r.name,
			pending: true
		};
	});
	context.commit('setResultsSummaries', pendingHistograms);

	const dataset = context.getters.getRouteDataset();

	// dispatch a request to fetch the results for each pipeline
	for (var result of results) {
		const name = result.name;
		const res = encodeURIComponent(result.pipeline.resultUri);
		axios.get(`/distil/results-summary/${ES_INDEX}/${dataset}/${res}`)
		.then(response => {
			// save the histogram data
			const histogram = response.data.histogram;
			if (!histogram) {
				context.commit('setResultsSummaries', [
					{
						name: name,
						err: 'No analysis available'
					}
				]);
				return;
			}
			// ensure buckets is not nil
			histogram.buckets = histogram.buckets ? histogram.buckets : [];
			histogram.name = name;
			context.commit('updateResultsSummaries', histogram);
		})
		.catch(error => {
			context.commit('setResultsSummaries', [
				{
					name: name,
					err: error
				}
			]);
			return;
		});
	}
}
