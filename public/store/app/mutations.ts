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

	setIsTask1(state: AppState, isTask1: boolean) {
		state.isTask1 = isTask1;
	},

	setIsTask2(state: AppState, isTask2: boolean) {
		state.isTask2 = isTask2;
	},

	setProblemDataset(state: AppState, dataset: string) {
		state.problemDataset = dataset;
	},

	setProblemTarget(state: AppState, target: string) {
		state.problemTarget = target;
	},

	setProblemTaskType(state: AppState, task: string) {
		state.problemTaskType = task;
	},

	setProblemTaskSubType(state: AppState, subtask: string) {
		state.problemTaskSubType = subtask;
	},

	setProblemMetrics(state: AppState, metrics: string[]) {
		state.problemMetrics = metrics;
	},
};
