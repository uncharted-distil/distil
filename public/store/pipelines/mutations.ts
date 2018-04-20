import _ from 'lodash';
import moment from 'moment';
import Vue from 'vue';
import { PipelineState, PipelineInfo } from './index';

export const mutations = {

	// adds a pipeline request or replaces an existing one if the ids match.
	updatePipelineRequests(state: PipelineState, pipeline: PipelineInfo) {
		const index = _.findIndex(state.pipelineRequests, p => {
			return p.pipelineId === pipeline.pipelineId;
		});
		if (index === -1) {
			state.pipelineRequests.push(pipeline);
		} else {
			if (moment(pipeline.timestamp) >= moment(state.pipelineRequests[index].timestamp)) {
				Vue.set(state.pipelineRequests, index, pipeline);
			}
		}
	},

	clearPipelineRequests(state: PipelineState) {
		state.pipelineRequests = [];
	}
}
