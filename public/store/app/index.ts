export interface UserSession {
	//sessionId: string;
}

export interface AppState {
	session: UserSession;
}

// shared data model
export const state: AppState = {
	session: {} as UserSession
};
