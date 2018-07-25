
export interface AppState {
	isAborted: boolean;
	versionNumber: string;
	versionTimestamp: string;
	isTask1: boolean;
	isTask2: boolean;
	problemDataset: string;
	problemTarget: string;
	problemTaskType: string;
	problemTaskSubType: string;
	problemMetrics: string[];
}

// shared data model
export const state: AppState = {
	isAborted: false,
	versionNumber: 'unknown',
	versionTimestamp: 'unknown',
	isTask1: false,
	isTask2: false,
	problemDataset: 'unknown',
	problemTarget: 'unknown',
	problemTaskType: 'unknown',
	problemTaskSubType: 'unknown',
	problemMetrics: []
};
