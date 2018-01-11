import { Dictionary } from '../../util/dict';
import { FilterParams } from '../../util/filters';

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
	filters: FilterParams;
}

export interface PipelineState {
	sessionID: string;
	runningPipelines: Dictionary<Dictionary<PipelineInfo>>;
	completedPipelines: Dictionary<Dictionary<PipelineInfo>>;
}

export const state: PipelineState = {

	// current pipeline session id
	sessionID: null,

	// running pipeline creation tasks grouped by parent create requestID
	runningPipelines: {} as any,

	// completed pipeline creation tasks grouped by parent create request ID
	completedPipelines: {} as any
}
