import _ from 'lodash';
import moment from 'moment';
import { Variable } from '../dataset/index';
import { SolutionState, Solution, SolutionRequest, SOLUTION_RUNNING, SOLUTION_ERRORED, SOLUTION_COMPLETED } from './index';
import { Dictionary } from '../../util/dict';
import { Stream } from '../../util/ws';

export function sortRequestsByTimestamp(a: SolutionRequest, b: SolutionRequest): number {
	// descending order
	return moment(b.timestamp).unix() - moment(a.timestamp).unix();
}

function getScoreValue(s: Solution): number {
	if (s.progress === SOLUTION_ERRORED) {
		return -1;
	}
	return (s.scores && s.scores.length > 0) ? (s.scores[0].value * s.scores[0].sortMultiplier) : -1;
}

export function sortSolutionsByScore(a: Solution, b: Solution): number {
	// descending order of score
	return getScoreValue(b) - getScoreValue(a);
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
		return running.sort(sortSolutionsByScore);
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
		return running.sort(sortSolutionsByScore);
	},

	getSolutions(state: SolutionState): Solution[] {
		let solutions = [];
		state.requests.forEach(request => {
			solutions = solutions.concat(request.solutions);
		});
		return solutions.sort(sortSolutionsByScore);
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
		return solutions.sort(sortSolutionsByScore);
	},

	getRelevantSolutionRequests(state: SolutionState, getters: any): SolutionRequest[] {
		const target = getters.getRouteTargetVariable;
		const dataset = getters.getRouteDataset;
		// get only matching dataset / target
		const requests = state.requests.filter(request => {
			return request.dataset === dataset && request.feature === target;
		});
		// sort and return
		requests.sort(sortRequestsByTimestamp);
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
		requests.sort(sortRequestsByTimestamp);
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
		return activeSolution.features.filter(f => f.featureType === 'train').map(f => variables[f.featureName]).filter(v => !!v);
	},

	getActiveSolutionTargetVariable(state: SolutionState, getters: any): Variable[] {
		const target = getters.getRouteTargetVariable;
		const variables = getters.getVariables;
		return variables.filter(variable => variable.colName === target);
	},

	getRequestStreams(state: SolutionState, getters: any): Dictionary<Stream> {
		return state.streams;
	}
};
