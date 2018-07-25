import _ from 'lodash';
import { Dictionary } from './dict';
import { SolutionState, Solution } from '../store/solutions/index';

export interface NameInfo {
	displayName: string,
	schemaName: string
}

export interface Task {
	displayName: string,
	schemaName: string
};

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

// Gets a task object based on a variable type.
export function getTask(varType: string): Task {
	const lowerType = _.toLower(varType);
	return _.get(TASKS_BY_VARIABLES, lowerType);
}

// CLASSIFICATION_TASK task info
export const CLASSIFICATION_TASK: Task = {
	displayName: 'Classification',
	schemaName: 'classification'
};

// regression task info
export const REGRESSION_TASK: Task = {
	displayName: 'Regression',
	schemaName: 'regression'
};

// variable type to task mappings
const TASKS_BY_VARIABLES: Dictionary<Task> = {
	float:  REGRESSION_TASK,
	latitude:  REGRESSION_TASK,
	longitude:  REGRESSION_TASK,
	integer: REGRESSION_TASK,
	image: CLASSIFICATION_TASK,
	timeseries: CLASSIFICATION_TASK,
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
	unknown: CLASSIFICATION_TASK,
	boolean: CLASSIFICATION_TASK
};
