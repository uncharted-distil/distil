import { FilterParams } from '../../util/filters';
import { Stream } from '../../util/ws';
import { Dictionary } from '../../util/dict';

export const SOLUTION_PENDING = 'SOLUTION_PENDING';
export const SOLUTION_RUNNING = 'SOLUTION_RUNNING';
export const SOLUTION_COMPLETED = 'SOLUTION_COMPLETED';
export const SOLUTION_ERRORED = 'SOLUTION_ERRORED';

export const REQUEST_PENDING = 'REQUEST_PENDING';
export const REQUEST_RUNNING = 'REQUEST_RUNNING';
export const REQUEST_COMPLETED = 'REQUEST_COMPLETED';
export const REQUEST_ERRORED = 'REQUEST_ERRORED';

export const NUM_SOLUTIONS = 3;

export interface Score {
	metric: string;
	label: string;
	value: number;
	sortMultiplier: number;
}

export interface SolutionFeature {
	featureName: string;
	featureType: string;
}

export interface Solution {
	requestId: string;
	feature: string;
	solutionId: string;
	resultId: string;
	progress: string;
	scores: Score[];
	timestamp: number;
	dataset: string;
	filters: FilterParams;
	features: SolutionFeature[];
	predictedKey: string;
	errorKey: string;
}

export interface SolutionRequest {
	requestId: string;
	dataset: string;
	feature: string;
	progress: string;
	solutions: Solution[];
	timestamp: number;
}

export interface SolutionState {
	requests: SolutionRequest[];
	streams: Dictionary<Stream>;
}

export const state: SolutionState = {
	requests: [],
	streams: {}
}
