import { AppState, UserSession } from './index';

export const mutations = {

	// sets the active session in the store as well as in the browser local storage
	setUserSession(state: AppState, session: UserSession) {
		state.session = session;
	},

	setAborted(state: AppState) {
		state.isAborted = true;
	}
};
