import _ from 'lodash';
import Vue from 'vue';
import { PipelineState, PipelineInfo } from './index';

export const mutations = {
	// adds a running pipeline or replaces an existing one if the ids match
	addRunningPipeline(state: PipelineState, pipelineData: PipelineInfo) {
		if (!_.has(state.runningPipelines, pipelineData.requestId)) {
			Vue.set(state.runningPipelines, pipelineData.requestId, {});
		}
		Vue.set(state.runningPipelines[pipelineData.requestId], pipelineData.pipelineId, pipelineData);
	},

	// removes a running pipeline
	removeRunningPipeline(state: PipelineState, args: { requestId: string, pipelineId: string }) {
		if (_.has(state.runningPipelines, args.requestId)) {
			// delete the pipeline from the request
			if (_.has(state.runningPipelines[args.requestId], args.pipelineId)) {
				Vue.delete(state.runningPipelines[args.requestId], args.pipelineId);
				// delete the request if empty
				if (_.size(state.runningPipelines[args.requestId]) === 0) {
					Vue.delete(state.runningPipelines, args.requestId);
				}
				return true;
			}
		}
		return false;
	},

	// adds a completed pipeline or replaces an existing one if the ids match
	addCompletedPipeline(state: PipelineState, pipelineData: PipelineInfo) {
		if (!_.has(state.completedPipelines, pipelineData.requestId)) {
			Vue.set(state.completedPipelines, pipelineData.requestId, {});
		}
		Vue.set(state.completedPipelines[pipelineData.requestId], pipelineData.pipelineId, pipelineData);
	},

	// removes a completed pipeline
	removeCompletedPipeline(state: PipelineState, args: { requestId: string, pipelineId: string }) {
		if (_.has(state.runningPipelines, args.requestId)) {
			// delete the pipeline from the request
			if (_.has(state.completedPipelines[args.requestId], args.pipelineId)) {
				// delete the request if empty
				Vue.delete(state.completedPipelines[args.requestId], args.pipelineId);
				if (_.size(state.completedPipelines[args.requestId]) === 0) {
					Vue.delete(state.completedPipelines, args.requestId);
				}
				return true;
			}
		}
		return false;
	}
}
