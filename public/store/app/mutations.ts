import { AppState, Session } from './index';

export const mutations = {
	// sets the active session in the store as well as in the browser local storage
	setPipelineSession(state: AppState, session: Session) {
		state.pipelineSession = session;
		if (!session) {
			window.localStorage.removeItem('pipeline-session-id');
		} else {
			window.localStorage.setItem('pipeline-session-id', session.id);
		}
	}
};

