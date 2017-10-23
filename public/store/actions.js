import _ from 'lodash';
import axios from 'axios';
import moment from 'moment';
import {encodeQueryParams} from '../util/filters';

// TODO: move this somewhere more appropriate.
const ES_INDEX = 'datasets';
const CREATE_PIPELINES_MSG = 'CREATE_PIPELINES';
const PIPELINE_COMPLETE = 'COMPLETED';
const STREAM_CLOSE = 'STREAM_CLOSE';
const FEATURE_TYPE_TARGET = 'target';


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

export function createResultName(dataset, timestamp, targetFeature) {
	const t = moment(timestamp);
	return `${dataset}: ${targetFeature} at ${t.format('MMMM Do YYYY, h:mm:ss.SS a')}`;
}

// searches dataset descriptions and column names for supplied terms
export function getSession(context, args) {
	const sessionId = args.sessionId;
	return axios.get(`/distil/session/${sessionId}`)
		.then(response => {
			if (response.data.pipelines) {
				response.data.pipelines.forEach((pipeline) => {
					// determine the target feature for this request
					let targetFeature = '';
					pipeline.Features.forEach((feature) => {
						if (feature.FeatureType === FEATURE_TYPE_TARGET) {
							targetFeature = feature.FeatureName;
						}
					});

					pipeline.Results.forEach((res) => {
						// inject the name and pipeline id
						const name = createResultName(pipeline.Dataset, res.CreatedTime, targetFeature);
						res.name = name;

						// add/update the running pipeline info
						if (res.Progress === PIPELINE_COMPLETE) {
							// add the pipeline to complete
							context.commit('addCompletedPipeline', {
								name: res.name,
								feature: targetFeature,
								timestamp: res.CreatedTime,
								requestId: pipeline.RequestID,
								dataset: pipeline.Dataset,
								pipelineId: res.PipelineID,
								pipeline: {
									resultUri: res.ResultUUID,
									output: '',
									scores: res.Scores
								}
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
export function getVariables(context, args) {
	const dataset = args.dataset;
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
export function getVariableSummaries(context, args) {
	const dataset = args.dataset;
	const variables = args.variables;
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
		axios.get(`/distil/variable-summaries/${ES_INDEX}/${dataset}/${variable.name}`)
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
}

// update filtered data based on the  current filter state
export function updateFilteredData(context, args) {
	const dataset = args.dataset;
	const filters = args.filters;
	const queryParams = encodeQueryParams(filters);
	const url = `distil/filtered-data/${ES_INDEX}/${dataset}/inclusive${queryParams}`;
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

// update filtered data based on the  current filter state
export function updateSelectedData(context, args) {
	const dataset = args.dataset;
	const filters = args.filters;
	const queryParams = encodeQueryParams(filters);
	const url = `distil/filtered-data/${ES_INDEX}/${dataset}/exclusive${queryParams}`;
	// request filtered data from server - no data is valid given filter settings
	return axios.get(url)
		.then(response => {
			context.commit('setSelectedData', response.data);
		})
		.catch(error => {
			console.error(error);
			context.commit('setSelectedData', []);
		});
}

// starts a pipeline session.
export function getPipelineSession(context, args) {
	const sessionId = args.sessionId;
	const conn = context.getters.getWebSocketConnection();
	return conn.send({
			type: 'GET_SESSION',
			session: sessionId
		}).then(res => {
			if (sessionId && res.created) {
				console.warn('previous session', sessionId, 'could not be resumed, new session created');
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
export function endPipelineSession(context, args) {
	const sessionId = args.sessionId;
	const conn = context.getters.getWebSocketConnection();
	if (!sessionId) {
		return;
	}
	return conn.send({
			type: 'END_SESSION',
			session: sessionId
		}).then(() => {
			context.commit('setPipelineSession', null);
		}).catch(err => {
			console.warn(err);
		});
}

// issues a pipeline create request
export function createPipelines(context, request) {
	const conn = context.getters.getWebSocketConnection();
	if (!request.sessionId) {
		console.warn('Missing session id');
		return;
	}
	const stream = conn.stream(res => {
		if (_.has(res, STREAM_CLOSE)) {
			stream.close();
			return;
		}
		// inject the name and pipeline id
		const name = createResultName(request.dataset, res.createdTime, request.feature);
		res.name = name;
		res.feature = request.feature;
		// add/update the running pipeline info
		context.commit('addRunningPipeline', res);
		if (res.progress === PIPELINE_COMPLETE) {
			//move the pipeline from running to complete
			context.commit('removeRunningPipeline', {pipelineId: res.pipelineId, requestId: res.requestId});
			context.commit('addCompletedPipeline', {
				name: res.name,
				feature: request.feature,
				timestamp: res.createdTime,
				requestId: res.requestId,
				dataset: res.dataset,
				pipelineId: res.pipelineId,
				pipeline: res.pipeline
			});
		}
	});

	stream.send({
		type: CREATE_PIPELINES_MSG,
		session: request.sessionId,
		index: ES_INDEX,
		dataset: request.dataset,
		feature: request.feature,
		task: request.task,
		metric: request.metric,
		output: request.output,
		maxPipelines: 3,
		filters: request.filters
	});
}

// fetches result summaries for a given pipeline create request
export function getResultsSummaries(context, args) {
	const dataset = args.dataset;
	const requestId = args.requestId;
	const results = context.getters.getPipelineResults(requestId);

	// save a placeholder histogram
	const pendingHistograms = _.map(results, r => {
		return {
			name: r.name,
			pending: true
		};
	});
	context.commit('setResultsSummaries', pendingHistograms);

	// fetch the results for each pipeline
	for (var result of results) {
		const name = result.name;
		const feature = result.feature;
		const pipelineId = result.pipelineId;
		const res = encodeURIComponent(result.pipeline.resultUri);
		axios.get(`/distil/results-summary/${ES_INDEX}/${dataset}/${res}`)
			.then(response => {
				// save the histogram data
				const histogram = response.data.histogram;
				if (!histogram) {
					context.commit('setResultsSummaries', [
						{
							name: name,
							feature: feature,
							pipelineId: pipelineId,
							err: 'No analysis available'
						}
					]);
					return;
				}
				// ensure buckets is not nil
				histogram.buckets = histogram.buckets ? histogram.buckets : [];
				histogram.name = name;
				histogram.feature = feature;
				histogram.pipelineId = pipelineId;
				context.commit('updateResultsSummaries', histogram);
			})
			.catch(error => {
				context.commit('setResultsSummaries', [
					{
						name: name,
						feature: feature,
						pipelineId: pipelineId,
						err: error
					}
				]);
				return;
			});
	}
}

// fetches result data for created pipeline
export function updateResults(context, args) {
	const encodedResultId = encodeURIComponent(args.resultId);
	const filters = args.filters;
	const queryParams = encodeQueryParams(filters);
	return axios.get(`/distil/results/${ES_INDEX}/${args.dataset}/${encodedResultId}/inclusive${queryParams}`)
		.then(response => {
			context.commit('setResultData', response.data);
		})
		.catch(error => {
			console.error(`Failed to fetch results from ${args.resultId} with error ${error}`);
		});
}

export function highlightFeatureRange(context, highlight) {
	context.commit('highlightFeatureRange', highlight);
}

export function clearFeatureHighlightRange(context, varName) {
	context.commit('clearFeatureHighlightRange', varName);
}

export function highlightFeatureValues(context, highlight) {
	context.commit('highlightFeatureValues', highlight);
}

export function clearFeatureHighlightValues(context) {
	context.commit('clearFeatureHighlightValues');
}

export function abort() {
	return axios.get('/distil/abort')
		.then(() => {
			console.log('User initiated session abort');
		})
		.catch(error => {
			console.error(`Failed to abort with error ${error}`);
		});
}

export function exportPipeline(context, args) {
	return axios.get(`/distil/export/${args.sessionId}/${args.pipelineId}`)
		.then(() => {
			console.log(`User exported pipeline ${args.pipelineId}`);
		})
		.catch(error => {
			console.error(`Failed to export with error ${error}`);
		});
}

export function addRecentDataset(context, dataset) {
	context.commit('addRecentDataset', dataset);
}
