import _ from 'lodash';
import axios from 'axios';
import moment from 'moment';
import { PipelineState, Score } from './index';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';
import { getWebSocketConnection } from '../../util/ws';

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
	FeatureName: string;
	FeatureType: string;
}

interface Result {
	name: string;
	ResultUUID: string;
	PipelineID: string;
	CreatedTime: number;
	Progress: string;
	Scores: Score[];
}

interface PipelineResponse {
	RequestID: string;
	Dataset: string;
	Features: Feature[];
	Results: Result[];
}


interface PipelineRequest {
	sessionId: string,
	dataset: string,
	feature: string,
	task: string,
	metric: string,
	output: string,
	filters: string
}

export type AppContext = ActionContext<PipelineState, DistilState>;

export const actions = {
	getSession(context: any, args: { sessionId: string }) {
		const sessionId = args.sessionId;
		return axios.get(`/distil/session/${sessionId}`)
		.then(response => {
			if (response.data.pipelines) {
				const pipelineResponse  = response.data.pipelines as PipelineResponse[];
				pipelineResponse.forEach((pipeline) => {
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
									resultId: res.ResultUUID,
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
	},

	createPipelines(context: any, request: PipelineRequest) {
		const conn = getWebSocketConnection();
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
				// move the pipeline from running to complete
				context.commit('removeRunningPipeline', { pipelineId: res.pipelineId, requestId: res.requestId });
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
	},
}
