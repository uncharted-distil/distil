import { PipelineState, PipelineInfo, PIPELINE_RUNNING, PIPELINE_UPDATED, PIPELINE_COMPLETED } from './index';
import localStorage from 'store';

function sortPipelines(a: PipelineInfo, b: PipelineInfo): number {
	if (a.pipelineId < b.pipelineId) {
		return -1
	}
	if (a.pipelineId > b.pipelineId) {
		return 1
	}
	return 0;
}

export const getters = {

	getPipelineSessionID(state: PipelineState): string {
		if (!state.sessionID) {
			const id = localStorage.get('pipeline-session-id');
			if (id) {
				console.log(`Loading session id ${id} from localStorage`);
			}
			return id;
		}
		return state.sessionID;
	},

	hasActiveSession(state: PipelineState): boolean {
		return state.sessionIsActive;
	},

	// Returns a dictionary of dictionaries, where the first key is the pipeline create request ID, and the second
	// key is the pipeline ID.
	getRunningPipelines(state: PipelineState): PipelineInfo[] {
		return state.pipelineRequests.filter(pipeline => pipeline.progress === PIPELINE_RUNNING).sort(sortPipelines);
	},

	// Returns a dictionary of dictionaries, where the first key is the pipeline create request ID, and the second
	// key is the pipeline ID.
	getCompletedPipelines(state: PipelineState): PipelineInfo[] {
		return state.pipelineRequests.filter(pipeline => pipeline.progress === PIPELINE_UPDATED || pipeline.progress === PIPELINE_COMPLETED).sort(sortPipelines);
	},

	getPipelines(state: PipelineState): PipelineInfo[] {
		return Array.from(state.pipelineRequests).slice().sort(sortPipelines);
	}
}
