
export interface AppState {
	isAborted: boolean;
	versionNumber: string;
	versionTimestamp: string;
	isDiscovery: boolean;
}

// shared data model
export const state: AppState = {
	isAborted: false,
	versionNumber: 'unknown',
	versionTimestamp: 'unknown',
	isDiscovery: false
};
