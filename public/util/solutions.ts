import _ from 'lodash';
import { Dictionary } from './dict';
import { sortSolutionsByScore } from '../store/solutions/getters';
import { getters as solutionGetters } from '../store/solutions/module';
import { SolutionState, Solution } from '../store/solutions/index';
import store from '../store/store';

export interface NameInfo {
	displayName: string;
	schemaName: string;
}

export interface Task {
	displayName: string;
	schemaName: string;
}

export function getSolutionIndex(solutionId: string) {
	const solutions = solutionGetters.getRelevantSolutions(store);
	const index = _.findIndex(solutions, solution => {
		return solution.solutionId === solutionId;
	});
	return solutions.length - index - 1;
}

export function getRequestIndex(requestId: string) {
	const requests = solutionGetters.getRelevantSolutionRequests(store);
	const index = _.findIndex(requests, req => {
		return req.requestId === requestId;
	});
	return requests.length - index - 1;
}

// Utility function to return all solution results associated with a given request ID
export function getSolutionsByRequestIds(state: SolutionState, requestIds: string[]): Solution[] {
	const ids = {};
	requestIds.forEach(id => {
		ids[id] = true;
	});

	let solutions = [];
	const filtered = state.requests.filter(request => ids[request.requestId]);
	filtered.forEach(request => {
		solutions = solutions.concat(request.solutions);
	});
	return solutions;
}

// Returns a specific solution result given a request and its solution id.
export function getSolutionById(state: SolutionState, solutionId: string): Solution {
	if (!solutionId) {
		return null;
	}
	let found = null;
	state.requests.forEach(request => {
		request.solutions.forEach(solution => {
			if (solution.solutionId === solutionId) {
				found = solution;
			}
		});
	});
	return found;
}

export function isTopSolutionByScore(state: SolutionState, requestId: string, solutionId: string, n: number): boolean {
	if (!solutionId) {
		return null;
	}
	const request = _.find(state.requests, req => {
		return req.requestId === requestId;
	});

	const sortedByScore = request.solutions.slice().sort(sortSolutionsByScore).slice(0, n);

	return !!_.find(sortedByScore, sol => {
		return sol.solutionId === solutionId;
	});
}

// Gets a task object based on a variable type.
export function getTask(varType: string): Task {
	const lowerType = _.toLower(varType);
	return _.get(TASKS_BY_VARIABLES, lowerType);
}

// classification task info
export const CLASSIFICATION_TASK: Task = {
	displayName: 'Classification',
	schemaName: 'classification'
};

// regression task info
export const REGRESSION_TASK: Task = {
	displayName: 'Regression',
	schemaName: 'regression'
};

export const TIMESERIES_FORECASTING_TASK: Task = {
	displayName: 'Time Series Forecasting',
	schemaName: 'timeSeriesForecasting'
};

// variable type to task mappings
const TASKS_BY_VARIABLES: Dictionary<Task> = {
	float:  REGRESSION_TASK,
	real:  REGRESSION_TASK,
	latitude:  REGRESSION_TASK,
	longitude:  REGRESSION_TASK,
	integer: REGRESSION_TASK,
	image: CLASSIFICATION_TASK,
	timeseries: TIMESERIES_FORECASTING_TASK,
	categorical: CLASSIFICATION_TASK,
	ordinal: CLASSIFICATION_TASK,
	address: CLASSIFICATION_TASK,
	city: CLASSIFICATION_TASK,
	state: CLASSIFICATION_TASK,
	country: CLASSIFICATION_TASK,
	email: CLASSIFICATION_TASK,
	phone: CLASSIFICATION_TASK,
	postal_code: CLASSIFICATION_TASK,
	uri: CLASSIFICATION_TASK,
	datetime: CLASSIFICATION_TASK,
	text: CLASSIFICATION_TASK,
	string: CLASSIFICATION_TASK,
	unknown: CLASSIFICATION_TASK,
	boolean: CLASSIFICATION_TASK
};
