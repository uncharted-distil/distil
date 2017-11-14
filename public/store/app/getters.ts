import { AppState } from './index';

export const getters = {
	getPipelineSessionID(state: AppState) {
		if (!state.pipelineSession.id) {
			const id = window.localStorage.getItem('pipeline-session-id');
			if (id) {
				console.log(`Loading session id ${id} from localStorage`);
			}
			return id;
		}
		return state.pipelineSession.id;
	},

	getPipelineSession(state: AppState) {
		return state.pipelineSession;
	}
};
