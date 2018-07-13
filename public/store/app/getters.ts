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
	},

	isDiscovery(state: AppState) {
		return state.isDiscovery;
	},

	getProblemDataset(state: AppState) {
		return state.problemDataset;
	},

	getProblemTarget(state: AppState) {
		return state.problemTarget;
	}
};
