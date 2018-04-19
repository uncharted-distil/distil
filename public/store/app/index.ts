
export interface AppState {
	isAborted: boolean;
	versionNumber: string;
	versionTimestamp: string;
}

// shared data model
export const state: AppState = {
	isAborted: false,
	versionNumber: 'unknown',
	versionTimestamp: 'unknown'
};
