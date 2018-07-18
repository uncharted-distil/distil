import { AppState } from './index';

export const getters = {

	isAborted(state: AppState): boolean {
		return state.isAborted;
	},

	getVersionNumber(state: AppState): string {
		return state.versionNumber;
	},

	getVersionTimestamp(state: AppState): string {
		return state.versionTimestamp;
	},

	getProblemDataset(state: AppState): string {
		return state.problemDataset;
	},

	getProblemTarget(state: AppState): string {
		return state.problemTarget;
	},

	isTask1(state: AppState): boolean {
		return state.isDiscovery;
	},

	isTask2(state: AppState): boolean {
		return state.problemTarget !== 'unknown' && state.problemDataset !== 'unknown';
	},
};
