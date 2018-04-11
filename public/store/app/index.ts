export interface UserSession {
	//sessionId: string;
}

export interface AppState {
	session: UserSession;
	isAborted: boolean;
	versionNumber: string;
	versionTimestamp: string;
}

// shared data model
export const state: AppState = {
	session: {} as UserSession,
	isAborted: false,
	versionNumber: 'unknown',
	versionTimestamp: 'unknown'
};
