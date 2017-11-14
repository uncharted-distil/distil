import { AppState, Session } from './index';

export const mutations = {
	// sets the active session in the store as well as in the browser local storage
	setPipelineSession(state: AppState, session: Session) {
		state.pipelineSession = session;
		if (!session) {
			window.localStorage.removeItem('pipeline-session-id');
		} else {
			console.log(`Storing session id ${session.id} in localStorage`);
			window.localStorage.setItem('pipeline-session-id', session.id);
		}
	}
};
