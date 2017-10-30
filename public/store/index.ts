export interface Score {
	metric: string;
	value: number;
}

export interface PipelineOutput {
	output: string,
	scores: Score[];
	resultId: string;
}

export interface PipelineInfo {
	requestId: string;
	name: string;
	id: string;
	feature: string;
	pipelineId: string;
	progress: string;
	pipeline?: PipelineOutput;
}

export interface PipelineRequestInfo {
	[pipelineId: string]: PipelineInfo;
}

export interface PipelineState {
	[requestId: string]: PipelineRequestInfo;
}

export interface Session {
	id: string;
	uuids: string[];
}

export interface DistilState {
	runningPipelines: PipelineState;
	completedPipelines: PipelineState;
	wsConnection: WebSocket;
	pipelineSession: Session;
}

// shared data model
export const state: DistilState = {
	// running pipeline creation tasks grouped by parent create requestID
	runningPipelines: {} as any,

	// completed pipeline creation tasks grouped by parent create request ID
	completedPipelines: {} as any,

	// the underlying websocket connection
	wsConnection: {} as any,

	// the pipeline session id
	pipelineSession: {} as any,
};
