import { Module } from 'vuex';
import { state, SolutionState } from './index';
import { getters as moduleGetters } from './getters';
import { actions as moduleActions } from './actions';
import { mutations as moduleMutations } from './mutations';
import { DistilState } from '../store';
import { getStoreAccessors } from 'vuex-typescript';

export const solutionModule: Module<SolutionState, DistilState> = {
	state: state,
	getters: moduleGetters,
	actions: moduleActions,
	mutations: moduleMutations
}

const { commit, read, dispatch } = getStoreAccessors<SolutionState, DistilState>(null);

export const getters = {
	getRunningSolutions: read(moduleGetters.getRunningSolutions),
	getCompletedSolutions: read(moduleGetters.getCompletedSolutions),
	getSolutions: read(moduleGetters.getSolutions),
	getSolutionsRequests: read(moduleGetters.getSolutionsRequests),
	getSolutionRequestIds: read(moduleGetters.getSolutionRequestIds),
	getActiveSolution: read(moduleGetters.getActiveSolution),
	getActiveSolutionTrainingMap: read(moduleGetters.getActiveSolutionTrainingMap),
	getActiveSolutionVariables: read(moduleGetters.getActiveSolutionVariables),
	isRegression: read(moduleGetters.isRegression)
}

export const actions = {
	fetchSolutions: dispatch(moduleActions.fetchSolutions),
	createSolutions: dispatch(moduleActions.createSolutions),
}

export const mutations = {
	updateSolutionRequests: commit(moduleMutations.updateSolutionRequests),
	clearSolutionRequests: commit(moduleMutations.clearSolutionRequests)
}
