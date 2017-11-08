import { AppState } from './index';

export const getters = {
	getPipelineSessionID(state: AppState) {
		if (!state.pipelineSession) {
			return window.localStorage.getItem('pipeline-session-id');
		}
		return state.pipelineSession.id;
	},

	getPipelineSession(state: AppState) {
		return state.pipelineSession;
	}
};
