import { PipelineState, PipelineInfo } from './index';
import { Dictionary } from '../data/index';

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
	}
}
