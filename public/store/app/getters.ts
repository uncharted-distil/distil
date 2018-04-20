import { AppState } from './index';

export const getters = {

	isAborted(state: AppState) {
		return state.isAborted;
	},

	getVersionNumber(state: AppState) {
		return state.versionNumber;
	},

	getVersionTimestamp(state: AppState) {
		return state.versionTimestamp;
	}
};
