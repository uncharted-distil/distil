import _ from 'lodash';
import moment from 'moment';
import { Variable } from '../dataset/index';
import { REGRESSION_TASK, CLASSIFICATION_TASK, getTask } from '../../util/solutions';
import { SolutionState, Solution, SolutionRequest, SOLUTION_RUNNING, SOLUTION_COMPLETED } from './index';
import { Dictionary } from '../../util/dict';
import { Stream } from '../../util/ws';

export function sortRequests(a: SolutionRequest, b: SolutionRequest): number {
	// descending order
	return moment(b.timestamp).unix() - moment(a.timestamp).unix();
}

export function sortSolutions(a: Solution, b: Solution): number {
	// ascending order
	return moment(a.timestamp).unix() - moment(b.timestamp).unix();
}

export const getters = {

	// Returns a dictionary of dictionaries, where the first key is the solution create request ID, and the second
	// key is the solution ID.
	getRunningSolutions(state: SolutionState): Solution[] {
		const running = [];
		state.requests.forEach(request => {
			request.solutions.forEach(solution => {
				if (solution.progress === SOLUTION_RUNNING) {
					running.push(solution);
				}
			});
		});
		return running.sort(sortSolutions);
	},

	// Returns a dictionary of dictionaries, where the first key is the solution create request ID, and the second
	// key is the solution ID.
	getCompletedSolutions(state: SolutionState): Solution[] {
		const running = [];
		state.requests.forEach(request => {
			request.solutions.forEach(solution => {
				if (solution.progress === SOLUTION_COMPLETED) {
					running.push(solution);
				}
			});
		});
		return running.sort(sortSolutions);
	},

	getSolutions(state: SolutionState): Solution[] {
		let solutions = [];
		state.requests.forEach(request => {
			solutions = solutions.concat(request.solutions);
		});
		return solutions.sort(sortSolutions);
	},

	getRelevantSolutions(state: SolutionState, getters: any): Solution[] {
		const target = getters.getRouteTargetVariable;
		const dataset = getters.getRouteDataset;
		const requests = state.requests.filter(request => {
			return request.dataset === dataset && request.feature === target;
		});
		let solutions = [];
		requests.forEach(request => {
			solutions = solutions.concat(request.solutions);
		});
		return solutions.sort(sortSolutions);
	},

	getRelevantSolutionRequests(state: SolutionState, getters: any): SolutionRequest[] {
		const target = getters.getRouteTargetVariable;
		const dataset = getters.getRouteDataset;
		// get only matching dataset / target
		const requests = state.requests.filter(request => {
			return request.dataset === dataset && request.feature === target;
		});
		// sort and return
		requests.sort(sortRequests);
		return requests;
	},

	getRelevantSolutionRequestIds(state: SolutionState, getters: any): string[] {
		const target = getters.getRouteTargetVariable;
		const dataset = getters.getRouteDataset;
		// get only matching dataset / targer
		const requests = state.requests.filter(request => {
			return request.dataset === dataset && request.feature === target;
		});
		// sort and return
		requests.sort(sortRequests);
		return requests.map(r => r.requestId);
	},

	getActiveSolution(state: SolutionState, getters: any): Solution {
		const solutionId = getters.getRouteSolutionId;
		const solutions = getters.getSolutions;
		return _.find(solutions, solution => solution.solutionId === solutionId);
	},

	getActiveSolutionTrainingVariables(state: SolutionState, getters: any): Variable[] {
		const activeSolution = getters.getActiveSolution;
		if (!activeSolution || !activeSolution.features) {
			return [];
		}
		const variables = getters.getVariablesMap;
		return activeSolution.features.filter(f => f.featureType === 'train').map(f => variables[f.featureName]);
	},

	getActiveSolutionTargetVariable(state: SolutionState, getters: any): Variable[] {
		const target = getters.getRouteTargetVariable;
		const variables = getters.getVariables;
		return variables.filter(variable => variable.key === target);
	},

	isRegression(state: SolutionState, getters: any): boolean {
		const variables = getters.getVariables;
		const target = getters.getRouteTargetVariable;
		const targetVariable = variables.find(s => _.toLower(s.key) === _.toLower(target));
		const task = getTask(targetVariable.type);
		return task.schemaName === REGRESSION_TASK.schemaName;
	},

	isClassification(state: SolutionState, getters: any): boolean {
		const variables = getters.getVariables;
		const target = getters.getRouteTargetVariable;
		const targetVariable = variables.find(s => _.toLower(s.key) === _.toLower(target));
		const task = getTask(targetVariable.type);
		return task.schemaName === CLASSIFICATION_TASK.schemaName;
	},

	getRequestStreams(state: SolutionState, getters: any): Dictionary<Stream> {
		return state.streams;
	}
}
