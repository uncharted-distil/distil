import { FilterParams } from '../../util/filters';

export const PIPELINE_PENDING = 'PENDING';
export const PIPELINE_RUNNING = 'RUNNING';
export const PIPELINE_COMPLETED = 'COMPLETED';
export const PIPELINE_ERRORED = 'ERRORED';

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
	resultId: string;
	progress: string;
	output: string;
	scores: Score[];
	timestamp: number;
	dataset: string;
	filters: FilterParams;
	features: PipelineFeature[];
}

export interface PipelineState {
	pipelineRequests: PipelineInfo[];
}

export const state: PipelineState = {
	pipelineRequests: [] as any
}
