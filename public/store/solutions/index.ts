import { FilterParams } from '../../util/filters';

export const SOLUTION_PENDING = 'PENDING';
export const SOLUTION_RUNNING = 'RUNNING';
export const SOLUTION_COMPLETED = 'COMPLETED';
export const SOLUTION_ERRORED = 'ERRORED';

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

export interface SolutionState {
	solutionRequests: SolutionInfo[];
}

export const state: SolutionState = {
	solutionRequests: [] as any
}
