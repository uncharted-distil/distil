export interface Session {
	id: string;
	uuids: string[];
}

export interface DistilState {
	wsConnection: WebSocket;
	pipelineSession: Session;
}

// shared data model
export const state: DistilState = {
	// the underlying websocket connection
	wsConnection: {} as any,

	// the pipeline session id
	pipelineSession: {} as any,
};
