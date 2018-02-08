import _ from 'lodash';
import axios from 'axios';
import { PipelineInfo, PipelineState, PIPELINE_UPDATED, PIPELINE_COMPLETED, PIPELINE_ERRORED } from './index';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';
import { mutations } from './module';
import { getWebSocketConnection } from '../../util/ws';
import { FilterParams } from '../../util/filters';

const ES_INDEX = 'datasets';
const CREATE_PIPELINES_MSG = 'CREATE_PIPELINES';
const STREAM_CLOSE = 'STREAM_CLOSE';
const FEATURE_TYPE_TARGET = 'target';

interface PipelineRequest {
	sessionId: string;
	dataset: string;
	feature: string;
	task: string;
	metric: string[];
	filters: FilterParams;
	maxPipelines: number;
}

export type AppContext = ActionContext<PipelineState, DistilState>;

export const actions = {

	// starts a pipeline session.
	startPipelineSession(context: AppContext, args: { sessionId: string }) {
		const sessionId = args.sessionId; // server creates a new session on null/undefined
		const conn = getWebSocketConnection();
		return conn.send({
				type: 'GET_SESSION',
				session: sessionId
			}).then(res => {
				if (sessionId && res.created) {
					console.warn('previous session', sessionId, 'could not be resumed, new session created');
				}
				mutations.setPipelineSessionID(context, res.session);
				mutations.setSessionActivity(context, true);
			}).catch((err: string) => {
				console.warn(err);
			});
	},

	// end a pipeline session.
	endPipelineSession(context: AppContext, args: { sessionId: string }) {
		if (!args.sessionId) {
			console.warn('`sessionId` argument is missing');
			return;
		}
		const sessionId = args.sessionId;
		const conn = getWebSocketConnection();
		return conn.send({
				type: 'END_SESSION',
				session: sessionId
			}).then(() => {
				mutations.setPipelineSessionID(context, null);
				mutations.setSessionActivity(context, false);
			}).catch(err => {
				console.warn(err);
			});
	},

	fetchPipelines(context: AppContext, args: { sessionId: string, dataset?: string, target?: string, pipelineId?: string }) {
		if (!args.sessionId) {
			console.warn('`sessionId` argument is missing');
			return;
		}
		if (!args.dataset) {
			args.dataset = 'null';
		}
		if (!args.target) {
			args.target = 'null';
		}
		if (!args.pipelineId) {
			args.pipelineId = 'null';
		}
		return axios.get(`/distil/session/${args.sessionId}/${args.dataset}/${args.target}/${args.pipelineId}`)
			.then(response => {
				if (!response.data.pipelines) {
					return;
				}
				const pipelines = response.data.pipelines as PipelineInfo[];
				pipelines.forEach(pipeline => {

					let targetFeature = '';
					pipeline.features.forEach(feature => {
						if (feature.featureType === FEATURE_TYPE_TARGET) {
							targetFeature = feature.featureName;
						}
					});

					// update pipeline
					mutations.updatePipelineRequest(context, {
						name: targetFeature,
						feature: targetFeature,
						filters: pipeline.filters,
						features: pipeline.features,
						requestId: pipeline.requestId,
						dataset: pipeline.dataset,
						timestamp: pipeline.timestamp,
						progress: pipeline.progress,
						pipelineId: pipeline.pipelineId,
						resultId: pipeline.resultId,
						scores: pipeline.scores,
						output: ''
					});
				});
			})
			.catch(error => {
				console.error(error);
			});
	},

	createPipelines(context: any, request: PipelineRequest) {
		return new Promise((resolve, reject) => {

			if (!request.sessionId) {
				console.warn('Missing session id');
				reject();
				return;
			}

			const conn = getWebSocketConnection();

			let receivedFirstResponse = false;

			const stream = conn.stream(res => {

				if (_.has(res, STREAM_CLOSE)) {
					stream.close();
					return;
				}

				if (res.error) {
					console.error(res.error);
				}

				res.name = request.feature;
				res.feature = request.feature;

				// NOTE: 'fetchPipeline' must be done first to ensure the
				// resultId is present to fetch summary

				// update pipeline status
				context.dispatch('fetchPipelines', {
					sessionId: request.sessionId,
					dataset: request.dataset,
					target: request.feature,
					pipelineId: res.pipelineId,
				}).then(() => {
					// update summaries
					if (res.progress === PIPELINE_ERRORED ||
						res.progress === PIPELINE_UPDATED ||
						res.progress == PIPELINE_COMPLETED) {

						// if current pipelineId, pull result summaries
						const currentPipelineId = context.getters.getRoutePipelineId;
						if (res.pipelineId === currentPipelineId) {
							context.dispatch('fetchResultSummaries', {
								dataset: request.dataset,
								pipelineId: res.pipelineId,
								variables: context.getters.getVariables
							});
						}

						context.dispatch('fetchPredictedSummary', {
							dataset: request.dataset,
							pipelineId: res.pipelineId
						});
						context.dispatch('fetchResidualsSummary', {
							dataset: request.dataset,
							pipelineId: res.pipelineId
						});
					}
				});

				// resolve promise on first response
				if (!receivedFirstResponse) {
					receivedFirstResponse = true;
					resolve(res);
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
				maxPipelines: request.maxPipelines,
				filters: request.filters
			});
		});
	},
}
