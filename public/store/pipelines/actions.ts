import axios from 'axios';
import { PipelineInfo, PipelineState, PIPELINE_COMPLETED, PIPELINE_ERRORED } from './index';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';
import { mutations } from './module';
import { getWebSocketConnection } from '../../util/ws';
import { FilterParams } from '../../util/filters';
import { regression } from '../../util/pipelines';

const ES_INDEX = 'datasets';
const CREATE_PIPELINES = 'CREATE_PIPELINES';

interface CreatePipelineRequest {
	dataset: string;
	target: string;
	task: string;
	maxPipelines: number;
	metrics: string[];
	filters: FilterParams;
}

export type AppContext = ActionContext<PipelineState, DistilState>;

function updateCurrentPipelineResults(context: any, req: CreatePipelineRequest, res: PipelineInfo) {

	const currentPipelineId = context.getters.getRoutePipelineId;

	// if current pipelineId, pull results
	if (res.pipelineId === currentPipelineId) {
		context.dispatch('fetchResultTableData', {
			dataset: req.dataset,
			pipelineId: res.pipelineId
		});
	}

	// if this is a regression task, pull extrema as a first step
	const isRegression = req.task.toLowerCase() === regression.schemaName.toLowerCase();
	let extremaFetches = [];
	if (isRegression) {
		extremaFetches = [
			context.dispatch('fetchResultExtrema', {
				dataset: req.dataset,
				variable: req.feature,
				pipelineId: res.pipelineId
			}),
			context.dispatch('fetchPredictedExtrema', {
				dataset: req.dataset,
				pipelineId: res.pipelineId
			})
		]
	}

	Promise.all(extremaFetches).then(() => {
		// if current pipelineId, pull result summaries
		if (res.pipelineId === currentPipelineId) {
			context.dispatch('fetchTrainingResultSummaries', {
				dataset: req.dataset,
				pipelineId: res.pipelineId,
				variables: context.getters.getActivePipelineVariables,
				extrema: context.getters.getPredictedExtrema
			});
		}
		context.dispatch('fetchPredictedSummary', {
			dataset: req.dataset,
			pipelineId: res.pipelineId,
			extrema: context.getters.getPredictedExtrema
		});
		context.dispatch('fetchResultHighlightValues', {
			dataset: req.dataset,
			highlightRoot: context.getters.getDecodedHighlightRoot,
			extrema: context.getters.getPredictedExtrema,
			pipelineId: res.pipelineId,
			requestIds: context.getters.getPipelines,
			variables: context.getters.getActivePipelineVariables
		});
	});

	if (isRegression) {
		context.dispatch('fetchResidualsExtrema', {
			dataset: req.dataset,
			pipelineId: res.pipelineId
		}).then(() => {
			context.dispatch('fetchResidualsSummary', {
				dataset: req.dataset,
				pipelineId: res.pipelineId,
				extrema: context.getters.getResidualExtrema
			});
		});
	}
}

function updatePipelineResults(context: any, req: PipelineRequest, res: PipelineInfo) {
	const isRegression = req.task.toLowerCase() === regression.schemaName.toLowerCase();
	let extremaFetches = [];
	if (isRegression) {
		extremaFetches = [
			context.dispatch('fetchTargetResultExtrema', {
				dataset: req.dataset,
				variable: req.feature,
				pipelineId: res.pipelineId
			}),
			context.dispatch('fetchPredictedExtrema', {
				dataset: req.dataset,
				pipelineId: res.pipelineId
			})
		]
	}
	Promise.all(extremaFetches).then(() => {
		context.dispatch('fetchPredictedSummary', {
			dataset: req.dataset,
			pipelineId: res.pipelineId,
			extrema: context.getters.getPredictedExtrema
		});
	});

	if (isRegression) {
		context.dispatch('fetchResidualsExtrema', {
			dataset: req.dataset,
			pipelineId: res.pipelineId
		}).then(() => {
			context.dispatch('fetchResidualsSummary', {
				dataset: req.dataset,
				pipelineId: res.pipelineId,
				extrema: context.getters.getResidualExtrema
			});
		});
	}
}

export const actions = {

	fetchPipeline(context: AppContext, args: { pipelineId?: string }) {
		if (!args.pipelineId) {
			console.warn('`pipelineId` argument is missing');
			return null;
		}

		return axios.get(`/distil/pipelines/null/null/${args.pipelineId}`)
			.then(response => {
				if (!response.data.pipelines) {
					return;
				}
				const pipelines = response.data.pipelines;
				pipelines.forEach(pipeline => {
					// update pipeline
					mutations.updatePipelineRequests(context, {
						name: pipeline.feature,
						feature: pipeline.feature,
						filters: pipeline.filters,
						features: pipeline.features,
						requestId: pipeline.requestId,
						dataset: pipeline.dataset,
						timestamp: pipeline.timestamp,
						progress: pipeline.progress,
						pipelineId: pipeline.pipelineId,
						resultId: pipeline.resultId,
						scores: pipeline.scores
					});
				});
			})
			.catch(error => {
				console.error(error);
			});
	},

	fetchPipelines(context: AppContext, args: { dataset?: string, target?: string, pipelineId?: string }) {
		if (!args.dataset) {
			args.dataset = null;
		}
		if (!args.target) {
			args.target = null;
		}
		if (!args.pipelineId) {
			args.pipelineId = null;
		}

		mutations.clearPipelineRequests(context);

		return axios.get(`/distil/pipelines/${args.dataset}/${args.target}/${args.pipelineId}`)
			.then(response => {
				if (!response.data.pipelines) {
					return;
				}
				const pipelines = response.data.pipelines;
				pipelines.forEach(pipeline => {
					// update pipeline
					mutations.updatePipelineRequests(context, {
						name: pipeline.feature,
						feature: pipeline.feature,
						filters: pipeline.filters,
						features: pipeline.features,
						requestId: pipeline.requestId,
						dataset: pipeline.dataset,
						timestamp: pipeline.timestamp,
						progress: pipeline.progress,
						pipelineId: pipeline.pipelineId,
						resultId: pipeline.resultId,
						scores: pipeline.scores
					});
				});
			})
			.catch(error => {
				console.error(error);
			});
	},

	createPipelines(context: any, request: CreatePipelineRequest) {
		return new Promise((resolve, reject) => {

			const conn = getWebSocketConnection();

			let receivedFirstResponse = false;

			const stream = conn.stream(res => {

				if (res.error) {
					console.error(res.error);
					return;
				}

				res.name = request.target;
				res.feature = request.target;

				// NOTE: 'fetchPipeline' must be done first to ensure the
				// resultId is present to fetch summary

				// update pipeline status
				context.dispatch('fetchPipeline', {
					dataset: request.dataset,
					target: request.target,
					pipelineId: res.pipelineId,
				}).then(() => {
					// update summaries
					if (res.progress === PIPELINE_ERRORED ||
						res.progress === PIPELINE_COMPLETED) {

						// if current pipelineId, pull results
						if (res.pipelineId === context.getters.getRoutePipelineId) {
							// current pipelineId is selected
							updateCurrentPipelineResults(context, request, res);
						} else {
							// current pipelineId is NOT selected
							updatePipelineResults(context, request, res);
						}

					}
				});

				// resolve promise on first response
				if (!receivedFirstResponse) {
					receivedFirstResponse = true;
					resolve(res);
				}

				// close stream on complete
				if (res.progress === PIPELINE_COMPLETED) {
					stream.close();
					return;
				}

			});

			// send create pipelines request
			stream.send({
				type: CREATE_PIPELINES,
				index: ES_INDEX,
				dataset: request.dataset,
				target: request.target,
				task: request.task,
				metrics: request.metrics,
				maxPipelines: request.maxPipelines,
				filters: request.filters
			});
		});
	},
}
