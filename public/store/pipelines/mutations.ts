import _ from 'lodash';
import moment from 'moment';
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

	// flags the session as active or inactive
	setSessionActivity(state: PipelineState, activity: boolean) {
		state.sessionIsActive = activity;
	},

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
