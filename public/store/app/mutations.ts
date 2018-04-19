import { AppState } from './index';

export const mutations = {

	setAborted(state: AppState) {
		state.isAborted = true;
	},

	setVersionNumber(state: AppState, versionNumber: string) {
		state.versionNumber = versionNumber;
	},

	setVersionTimestamp(state: AppState, versionTimestamp: string) {
		state.versionTimestamp = versionTimestamp;
	},
};
