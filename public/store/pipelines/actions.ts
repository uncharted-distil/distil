import _ from 'lodash';
import axios from 'axios';
import { PipelineState, Score } from './index';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';
import { mutations } from './module';
import { getWebSocketConnection } from '../../util/ws';
import { FilterParams } from '../../util/filters';

const ES_INDEX = 'datasets';
const CREATE_PIPELINES_MSG = 'CREATE_PIPELINES';
const STREAM_CLOSE = 'STREAM_CLOSE';
const FEATURE_TYPE_TARGET = 'target';

interface Feature {
	featureName: string;
	featureType: string;
}

interface Result {
	name: string;
	resultId: string;
	pipelineId: string;
	createdTime: number;
	progress: string;
	scores: Score[];
}

interface PipelineResponse {
	requestId: string;
	dataset: string;
	features: Feature[];
	filters: FilterParams;
	results: Result[];
}

interface PipelineRequest {
	sessionId: string;
	dataset: string;
	feature: string;
	task: string;
	metric: string[];
	filters: FilterParams;
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
			}).catch((err: string) => {
				console.warn(err);
			});
	},

	// end a pipeline session.
	endPipelineSession(context: AppContext, args: { sessionId: string }) {
		if (!args.sessionId) {
			console.warn('Missing session id');
			return;
		}
		const sessionId = args.sessionId;
		const conn = getWebSocketConnection();
		return conn.send({
				type: 'END_SESSION',
				session: sessionId
			}).then(() => {
				mutations.setPipelineSessionID(context, null);
			}).catch(err => {
				console.warn(err);
			});
	},

	fetchPipelines(context: AppContext, args: { sessionId: string }) {
		if (!args.sessionId) {
			console.warn('Missing session id');
			return;
		}
		const sessionId = args.sessionId;
		return axios.get(`/distil/session/${sessionId}`)
			.then(response => {
				if (response.data.pipelines) {
					const pipelineResponse = response.data.pipelines as PipelineResponse[];
					pipelineResponse.forEach(pipeline => {

						// determine the target feature for this request
						let targetFeature = '';
						pipeline.features.forEach((feature) => {
							if (feature.featureType === FEATURE_TYPE_TARGET) {
								targetFeature = feature.featureName;
							}
						});

						// for each result
						pipeline.results.forEach(result => {
							// update pipeline
							mutations.updatePipelineRequest(context, {
								name: targetFeature,
								filters: pipeline.filters,
								features: pipeline.features,
								requestId: pipeline.requestId,
								dataset: pipeline.dataset,
								feature: targetFeature,
								timestamp: result.createdTime,
								progress: result.progress,
								pipelineId: result.pipelineId,
								resultId: result.resultId,
								scores: result.scores,
								output: ''
							});
						});
					});
				}
			})
			.catch(error => {
				console.error(error);
			});
	},

	fetchPipeline(context: AppContext, args: { sessionId: string, pipelineId: string }) {
		// TODO: impl this on the backend
		if (!args.sessionId) {
			console.warn('Missing session id');
			return;
		}
		if (!args.pipelineId) {
			console.warn('Missing pipeline id');
			return;
		}
		const sessionId = args.sessionId;
		return axios.get(`/distil/session/${sessionId}`)
			.then(response => {
				if (response.data.pipelines) {
					const pipelineResponse = response.data.pipelines as PipelineResponse[];
					pipelineResponse.forEach(pipeline => {

						// determine the target feature for this request
						let targetFeature = '';
						pipeline.features.forEach((feature) => {
							if (feature.featureType === FEATURE_TYPE_TARGET) {
								targetFeature = feature.featureName;
							}
						});

						// for each result
						pipeline.results.forEach(result => {
							// only update for the provided pipeline id
							if (result.pipelineId === args.pipelineId) {
								// update pipeline
								mutations.updatePipelineRequest(context, {
									name: targetFeature,
									filters: pipeline.filters,
									features: pipeline.features,
									requestId: pipeline.requestId,
									dataset: pipeline.dataset,
									feature: targetFeature,
									timestamp: result.createdTime,
									progress: result.progress,
									pipelineId: result.pipelineId,
									resultId: result.resultId,
									scores: result.scores,
									output: ''
								});
							}
						});
					});
				}
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

				res.name = request.feature;
				res.feature = request.feature;

				// NOTE: 'fetchPipeline' must be done first to ensure the resultId
				// is present to fetch summary

				// update pipeline status
				context.dispatch('fetchPipeline', {
					sessionId: request.sessionId,
					pipelineId: res.pipelineId
				}).then(() => {
					// update summaries
					context.dispatch('fetchResultsSummary', {
						dataset: request.dataset,
						pipelineId: res.pipelineId
					});
					context.dispatch('fetchResidualsSummary', {
						dataset: request.dataset,
						pipelineId: res.pipelineId
					});
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
				maxPipelines: 3,
				filters: request.filters
			});
		});
	},
}
