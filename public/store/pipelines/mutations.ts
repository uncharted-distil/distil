import _ from 'lodash';
import Vue from 'vue';
import { PipelineState, PipelineInfo } from './index';
import localStorage from 'store';

export const mutations = {
	// sets the active session in the store as well as in the browser local storage
	setPipelineSessionID(state: PipelineState, sessionID: string) {
		state.sessionID = sessionID;
		if (!sessionID) {
			localStorage.remove('pipeline-session-id');
		} else {
			console.log(`Storing session id ${sessionID} in localStorage`);
			localStorage.set('pipeline-session-id', sessionID);
		}
	},

	// adds a pipeline request or replaces an existing one if the ids match.
	updatePipelineRequest(state: PipelineState, pipelineData: PipelineInfo) {
		const index = _.findIndex(state.pipelineRequests, pipeline => {
			return pipeline.pipelineId === pipelineData.pipelineId;
		});
		if (index === -1) {
			state.pipelineRequests.push(pipelineData);
		} else {
			Vue.set(state.pipelineRequests, index, pipelineData);
		}
	}
}
