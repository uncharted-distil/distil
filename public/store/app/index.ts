export interface Session {
	id: string;
	uuids: string[];
}

export interface AppState {
	pipelineSession: Session;
}

// shared data model
export const state: AppState = {
	// the pipeline session id
	pipelineSession: {} as any,
};
