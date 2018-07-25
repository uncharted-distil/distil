
export interface AppState {
	isAborted: boolean;
	versionNumber: string;
	versionTimestamp: string;
	isDiscovery: boolean;
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
	isDiscovery: false,
	problemDataset: 'unknown',
	problemTarget: 'unknown',
	problemTaskType: 'unknown',
	problemTaskSubType: 'unknown',
	problemMetrics: []
};
