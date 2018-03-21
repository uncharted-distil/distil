import { AppState } from './index';

export const getters = {

	getUserSession(state: AppState) {
		return state.session;
	},

	isAborted(state: AppState) {
		return state.isAborted;
	}
};
