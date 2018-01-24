import { FilterParams } from '../../util/filters';

export const PIPELINE_SUBMITTED = 'SUBMITTED';
export const PIPELINE_RUNNING = 'RUNNING';
export const PIPELINE_UPDATED = 'UPDATED';
export const PIPELINE_COMPLETED = 'COMPLETED';

export interface Score {
	metric: string;
	value: number;
}

export interface PipelineFeature {
	featureName: string;
	featureType: string;
}

export interface PipelineInfo {
	requestId: string;
	name: string;
	feature: string;
	pipelineId: string;
	progress: string;
	output: string;
	scores: Score[];
	resultId: string;
	timestamp: number;
	dataset: string;
	filters: FilterParams;
	features: PipelineFeature[];
}

export interface PipelineState {
	sessionID: string;
	sessionIsActive: boolean;
	pipelineRequests: PipelineInfo[];
}

export const state: PipelineState = {
	// current pipeline session id
	sessionID: null,
	// if there is an active session
	sessionIsActive: false,
	// pipeline requests
	pipelineRequests: [] as any
}
