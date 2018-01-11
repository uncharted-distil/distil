import _ from 'lodash';
import { PipelineState, PipelineInfo } from './index';
import { Dictionary } from '../../util/dict';
import localStorage from 'store';

export const getters = {

	getPipelineSessionID(state: PipelineState) {
		if (!state.sessionID) {
			const id = localStorage.get('pipeline-session-id');
			if (id) {
				console.log(`Loading session id ${id} from localStorage`);
			}
			return id;
		}
		return state.sessionID;
	},

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
