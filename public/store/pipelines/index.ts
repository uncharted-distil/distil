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
	feature: string;
	pipelineId: string;
	progress: string;
	pipeline?: PipelineOutput;
	timestamp: number;
	dataset: string;
}

export interface PipelineRequestInfo {
	[pipelineId: string]: PipelineInfo;
}

export interface Pipeline {
	[requestId: string]: PipelineRequestInfo;
}

export interface PipelineState {
	runningPipelines: Pipeline;
	completedPipelines: Pipeline;
}

export const state: PipelineState = {
	// running pipeline creation tasks grouped by parent create requestID
	runningPipelines: {} as any,

	// completed pipeline creation tasks grouped by parent create request ID
	completedPipelines: {} as any
}
