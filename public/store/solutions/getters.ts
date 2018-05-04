import _ from 'lodash';
import { Variable } from '../data/index';
import { SolutionState, SolutionInfo, SOLUTION_RUNNING, SOLUTION_COMPLETED } from './index';
import { Dictionary } from '../../util/dict';

function sortSolutions(a: SolutionInfo, b: SolutionInfo): number {
	if (a.solutionId < b.solutionId) {
		return -1
	}
	if (a.solutionId > b.solutionId) {
		return 1
	}
	return 0;
}

export const getters = {

	// Returns a dictionary of dictionaries, where the first key is the solution create request ID, and the second
	// key is the solution ID.
	getRunningSolutions(state: SolutionState): SolutionInfo[] {
		return state.solutionRequests.filter(solution => solution.progress === SOLUTION_RUNNING).sort(sortSolutions);
	},

	// Returns a dictionary of dictionaries, where the first key is the solution create request ID, and the second
	// key is the solution ID.
	getCompletedSolutions(state: SolutionState): SolutionInfo[] {
		return state.solutionRequests.filter(solution => solution.progress === SOLUTION_COMPLETED).sort(sortSolutions);
	},

	getSolutions(state: SolutionState): SolutionInfo[] {
		return Array.from(state.solutionRequests).slice().sort(sortSolutions);
	},

	getSolutionRequestIds(state: SolutionState): string[] {
		const ids = [];
		state.solutionRequests.forEach(solution => {
			if (ids.indexOf(solution.requestId) === -1) {
				ids.push(solution.requestId);
			}
		});
		return ids;
	},

	getActiveSolution(state: SolutionState, getters: any): SolutionInfo {
		const solutionId = getters.getRouteSolutionId;
		return _.find(state.solutionRequests, solution => solution.solutionId === solutionId);
	},

	getActiveSolutionTrainingMap(state: SolutionState, getters: any): Dictionary<boolean> {
		const activeSolution = getters.getActiveSolution;
		if (!activeSolution || !activeSolution.features) {
			return {};
		}
		const training = activeSolution.features.filter(f => f.featureType === 'train').map(f => f.featureName);
		const trainingMap = {};
		training.forEach(t => {
			trainingMap[t] = true;
		});
		return trainingMap;
	},

	getActiveSolutionVariables(state: SolutionState, getters: any): Variable[] {
		const trainingMap = getters.getActiveSolutionTrainingMap;
		const target = getters.getRouteTargetVariable;
		const variables = getters.getVariables;
		return variables.filter(variable => trainingMap[variable.name] || variable.name === target);
	}


}
