import _ from 'lodash';
import { Variable } from '../dataset/index';
import { regression, getTask } from '../../util/solutions';
import { SolutionState, SolutionInfo, SolutionRequest, SOLUTION_RUNNING, SOLUTION_COMPLETED } from './index';
import { Dictionary } from '../../util/dict';

function sortRequests(a: SolutionRequest, b: SolutionRequest): number {
	// descending order
	const aOldest = _.maxBy(a.solutions, sol => sol.timestamp) as any;
	const bOldest = _.maxBy(b.solutions, sol => sol.timestamp) as any;
	return bOldest - aOldest;
}

function sortSolutions(a: SolutionInfo, b: SolutionInfo): number {
	// ascending order
	return a.timestamp - b.timestamp;
}

export const getters = {

	// Returns a dictionary of dictionaries, where the first key is the solution create request ID, and the second
	// key is the solution ID.
	getRunningSolutions(state: SolutionState): SolutionInfo[] {
		return state.solutions.filter(solution => solution.progress === SOLUTION_RUNNING).sort(sortSolutions);
	},

	// Returns a dictionary of dictionaries, where the first key is the solution create request ID, and the second
	// key is the solution ID.
	getCompletedSolutions(state: SolutionState): SolutionInfo[] {
		return state.solutions.filter(solution => solution.progress === SOLUTION_COMPLETED).sort(sortSolutions);
	},

	getSolutions(state: SolutionState): SolutionInfo[] {
		return state.solutions.slice().sort(sortSolutions);
	},

	getSolutionsRequests(state: SolutionState): SolutionRequest[] {
		const reqs = {};
		state.solutions.forEach(solution => {
			if (!reqs[solution.requestId]) {
				reqs[solution.requestId] = {
					requestId: solution.requestId,
					dataset: solution.dataset,
					feature: solution.feature,
					// TODO: FIX THIS
					progress: 'UH OH',
					solutions: []
				};
			}
			reqs[solution.requestId].solutions.push(solution);
		});
		return _.map(reqs, req => {
			req.solutions.sort(sortSolutions);
			return req;
		}).sort(sortRequests);
	},

	getSolutionRequestIds(state: SolutionState): string[] {
		const ids = [];
		state.solutions.forEach(solution => {
			if (ids.indexOf(solution.requestId) === -1) {
				ids.push(solution.requestId);
			}
		});
		return ids;
	},

	getActiveSolution(state: SolutionState, getters: any): SolutionInfo {
		const solutionId = getters.getRouteSolutionId;
		return _.find(state.solutions, solution => solution.solutionId === solutionId);
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
	},

	isRegression(state: SolutionState, getters: any): boolean {
		const variables = getters.getVariables;
		const target = getters.getRouteTargetVariable;
		const targetVariable = variables.find(s => s.name === target);
		const task = getTask(targetVariable.type);
		return task.schemaName === regression.schemaName;
	}
}
