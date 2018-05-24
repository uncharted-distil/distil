import { FilterParams } from '../../util/filters';

export const SOLUTION_PENDING = 'SOLUTION_PENDING';
export const SOLUTION_RUNNING = 'SOLUTION_RUNNING';
export const SOLUTION_COMPLETED = 'SOLUTION_COMPLETED';
export const SOLUTION_ERRORED = 'SOLUTION_ERRORED';

export const REQUEST_PENDING = 'REQUEST_PENDING';
export const REQUEST_RUNNING = 'REQUEST_RUNNING';
export const REQUEST_COMPLETED = 'REQUEST_COMPLETED';
export const REQUEST_ERRORED = 'REQUEST_ERRORED';

export interface Score {
	metric: string;
	value: number;
}

export interface SolutionFeature {
	featureName: string;
	featureType: string;
}

export interface SolutionInfo {
	requestId: string;
	name: string;
	feature: string;
	solutionId: string;
	resultId: string;
	progress: string;
	scores: Score[];
	timestamp: number;
	dataset: string;
	filters: FilterParams;
	features: SolutionFeature[];
}

export interface SolutionRequest {
	requestId: string;
	dataset: string;
	feature: string;
	progress: string;
	solutions: SolutionInfo[];
}

export interface SolutionState {
	solutions: SolutionInfo[];
}

export const state: SolutionState = {
	solutions: []
}
