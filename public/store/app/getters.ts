import { AppState, StatusPanelState } from './index';

export const getters = {

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
		return state.isTask1;
	},

	isTask2(state: AppState): boolean {
		return state.isTask2;
	},

	getProblemTaskType(state: AppState): string {
		return state.problemTaskType;
	},

	getProblemTaskSubType(state: AppState): string {
		return state.problemTaskSubType;
	},

	getProblemMetrics(state: AppState): string[] {
		return state.problemMetrics;
	},

	getStatusPanelState(state: AppState): StatusPanelState {
		return state.statusPanelState;
	}
};
