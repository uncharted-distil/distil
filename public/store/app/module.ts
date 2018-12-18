import { Module } from 'vuex';
import { DistilState } from '../store';
import { state, AppState } from './index';
import { getters as moduleGetters } from './getters';
import { actions as moduleActions } from './actions';
import { mutations as moduleMutations } from './mutations';
import { getStoreAccessors } from 'vuex-typescript';

export const appModule: Module<AppState, DistilState> = {
	state: state,
	getters: moduleGetters,
	actions: moduleActions,
	mutations: moduleMutations
};

const { commit, read, dispatch } = getStoreAccessors<AppState, DistilState>(null);

// typed getters
export const getters = {
	isAborted: read(moduleGetters.isAborted),
	getVersionNumber: read(moduleGetters.getVersionNumber),
	getVersionTimestamp: read(moduleGetters.getVersionTimestamp),
	isTask1: read(moduleGetters.isTask1),
	isTask2: read(moduleGetters.isTask2),
	getProblemDataset: read(moduleGetters.getProblemDataset),
	getProblemTarget: read(moduleGetters.getProblemTarget),
	getProblemTaskType: read(moduleGetters.getProblemTaskType),
	getProblemTaskSubType: read(moduleGetters.getProblemTaskSubType),
	getProblemMetrics: read(moduleGetters.getProblemMetrics)
};

// typed actions
export const actions = {
	abort: dispatch(moduleActions.abort),
	exportSolution: dispatch(moduleActions.exportSolution),
	exportProblem: dispatch(moduleActions.exportProblem),
	fetchConfig: dispatch(moduleActions.fetchConfig)
};

// type mutators
export const mutations = {
	setAborted: commit(moduleMutations.setAborted),
	setVersionNumber: commit(moduleMutations.setVersionNumber),
	setVersionTimestamp: commit(moduleMutations.setVersionTimestamp),
	setIsTask1: commit(moduleMutations.setIsTask1),
	setIsTask2: commit(moduleMutations.setIsTask2),
	setProblemDataset: commit(moduleMutations.setProblemDataset),
	setProblemTarget: commit(moduleMutations.setProblemTarget),
	setProblemTaskType: commit(moduleMutations.setProblemTaskType),
	setProblemTaskSubType: commit(moduleMutations.setProblemTaskSubType),
	setProblemMetrics: commit(moduleMutations.setProblemMetrics)
};
