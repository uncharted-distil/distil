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
}

const { commit, read, dispatch } = getStoreAccessors<AppState, DistilState>(null);

// typed getters
export const getters = {
	isAborted: read(moduleGetters.isAborted),
	getVersionNumber: read(moduleGetters.getVersionNumber),
	getVersionTimestamp: read(moduleGetters.getVersionTimestamp),
	isDiscovery: read(moduleGetters.isDiscovery)
}

// typed actions
export const actions = {
	abort: dispatch(moduleActions.abort),
	exportSolution: dispatch(moduleActions.exportSolution),
	exportProblem: dispatch(moduleActions.exportProblem),
	fetchConfig: dispatch(moduleActions.fetchConfig)
}

// type mutators
export const mutations = {
	setAborted: commit(moduleMutations.setAborted),
	setVersionNumber: commit(moduleMutations.setVersionNumber),
	setVersionTimestamp: commit(moduleMutations.setVersionTimestamp),
	setIsDiscovery: commit(moduleMutations.setIsDiscovery)
}
