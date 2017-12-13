import _ from 'lodash';
import { PipelineState, PipelineInfo } from './index';
import { Dictionary } from '../../util/dict';

export const getters = {
	// Returns a dictionary of dictionaries, where the first key is the pipeline create request ID, and the second
	// key is the pipeline ID.
	getRunningPipelines(state: PipelineState): Dictionary<Dictionary<PipelineInfo>> {
		return state.runningPipelines;
	},

	// Returns a dictionary of dictionaries, where the first key is the pipeline create request ID, and the second
	// key is the pipeline ID.
	getCompletedPipelines(state: PipelineState): Dictionary<Dictionary<PipelineInfo>> {
		return state.completedPipelines;
	},

	getPipelines(state: PipelineState): Dictionary<Dictionary<PipelineInfo>> {
		const pipelines: Dictionary<Dictionary<PipelineInfo>> = {};
		_.forIn(state.runningPipelines, (requestGroup, requestId) => {
			pipelines[requestId] = requestGroup;
		});
		_.forIn(state.completedPipelines, (requestGroup, requestId) => {
			if (!pipelines[requestId]) {
				pipelines[requestId] = requestGroup;
			} else {
				_.forIn(requestGroup, (pipeline, pipelineId) => {
					pipelines[requestId][pipelineId] = pipeline;
				});
			}
		});
		return pipelines;
	}
}
