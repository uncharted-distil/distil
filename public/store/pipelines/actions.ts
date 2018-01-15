import _ from 'lodash';
import axios from 'axios';
import moment from 'moment';
import { PipelineState, Score } from './index';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';
import { mutations } from './module';
import { getWebSocketConnection } from '../../util/ws';
import { FilterParams } from '../../util/filters';

// TODO: move this somewhere more appropriate.
const ES_INDEX = 'datasets';
const CREATE_PIPELINES_MSG = 'CREATE_PIPELINES';
const STREAM_CLOSE = 'STREAM_CLOSE';
const PIPELINE_COMPLETE = 'COMPLETED';
const FEATURE_TYPE_TARGET = 'target';

function createResultName(dataset: string, timestamp: number, targetFeature: string) {
	const t = moment(timestamp);
	return `${dataset}: ${targetFeature} at ${t.format('MMMM Do YYYY, h:mm:ss.SS a')}`;
}

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
	getSession(context: AppContext, args: { sessionId: string }) {
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

						pipeline.results.forEach((res) => {
							// inject the name and pipeline id
							const name = createResultName(pipeline.dataset, res.createdTime, targetFeature);
							res.name = name;

							// add/update the running pipeline info
							if (res.progress === PIPELINE_COMPLETE) {
								// add the pipeline to complete
								mutations.addCompletedPipeline(context, {
									name: res.name,
									feature: targetFeature,
									timestamp: res.createdTime,
									progress: res.progress,
									requestId: pipeline.requestId,
									dataset: pipeline.dataset,
									pipelineId: res.pipelineId,
									pipeline: {
										resultId: res.resultId,
										output: '',
										scores: res.scores
									},
									filters: pipeline.filters
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
			const conn = getWebSocketConnection();
			if (!request.sessionId) {
				console.warn('Missing session id');
				reject();
				return;
			}

			let receivedFirstResponse = false;

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
				mutations.addRunningPipeline(context, res);

				// update summaries
				context.dispatch('getResultsSummaries', {
					dataset: request.dataset,
					requestId: res.requestId
				});
				context.dispatch('getResidualsSummaries', {
					dataset: request.dataset,
					requestId: res.requestId
				});

				// resolve promise on first response
				if (!receivedFirstResponse) {
					receivedFirstResponse = true;
					resolve(res);
				}

				if (res.progress === PIPELINE_COMPLETE) {
					// move the pipeline from running to complete
					mutations.removeRunningPipeline(context, {
						pipelineId: res.pipelineId,
						requestId: res.requestId
					});
					mutations.addCompletedPipeline(context, {
						name: res.name,
						feature: request.feature,
						progress: res.progress,
						timestamp: res.createdTime,
						requestId: res.requestId,
						dataset: res.dataset,
						pipelineId: res.pipelineId,
						pipeline: res.pipeline,
						filters: res.filters,
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
				maxPipelines: 1,
				filters: request.filters
			});
		});
	},
}
