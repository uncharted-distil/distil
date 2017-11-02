import _ from 'lodash';
import { PipelineState } from './index';

export const getters = {
	getPipelineResults(state: PipelineState) {
		return (requestId: string) => {
			return _.concat(
				_.values(state.runningPipelines[requestId]),
				_.values(state.completedPipelines[requestId]));
		};
	},

	getRunningPipelines(state: PipelineState) {
		return () => state.runningPipelines;
	},

	getCompletedPipelines(state: PipelineState) {
		return () => state.completedPipelines;
	}
}
