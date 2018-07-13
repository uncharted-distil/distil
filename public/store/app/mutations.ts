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

	setIsDiscovery(state: AppState, isDiscovery: boolean) {
		state.isDiscovery = isDiscovery;
	},

	setProblemDataset(state: AppState, dataset: string) {
		state.problemDataset = dataset;
	},

	setProblemTarget(state: AppState, target: string) {
		state.problemTarget = target;
	}
};
