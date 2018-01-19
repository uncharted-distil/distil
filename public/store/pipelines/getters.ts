import { PipelineState, PipelineInfo, PIPELINE_RUNNING, PIPELINE_UPDATED, PIPELINE_COMPLETED } from './index';
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
	getRunningPipelines(state: PipelineState): PipelineInfo[] {
		return state.pipelineRequests.filter(pipeline => pipeline.progress === PIPELINE_RUNNING).sort((a, b) => b.timestamp - a.timestamp);
	},

	// Returns a dictionary of dictionaries, where the first key is the pipeline create request ID, and the second
	// key is the pipeline ID.
	getCompletedPipelines(state: PipelineState): PipelineInfo[] {
		return state.pipelineRequests.filter(pipeline => pipeline.progress === PIPELINE_UPDATED || pipeline.progress === PIPELINE_COMPLETED).sort((a, b) => b.timestamp - a.timestamp);
	},

	getPipelines(state: PipelineState): PipelineInfo[] {
		return Array.from(state.pipelineRequests).sort((a, b) => b.timestamp - a.timestamp);
	}
}
