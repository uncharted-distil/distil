export interface UserSession {
	//sessionId: string;
}

export interface AppState {
	session: UserSession;
	isAborted: boolean;
}

// shared data model
export const state: AppState = {
	session: {} as UserSession,
	isAborted: false
};
