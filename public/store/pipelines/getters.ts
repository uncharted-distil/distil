import { PipelineState } from './index';

export const getters = {
	getRunningPipelines(state: PipelineState) {
		return state.runningPipelines;
	},

	getCompletedPipelines(state: PipelineState) {
		return state.completedPipelines;
	}
}
