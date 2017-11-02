import { Module } from 'vuex';
import { DistilState } from '../store';
import { state, AppState } from './index';
import { getters as moduleGetters } from './getters';
import { actions as moduleActions} from './actions';
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
	getPipelineSessionID: read(moduleGetters.getPipelineSessionID),
	getPipelineSession: read(moduleGetters.getPipelineSession)
}

// typed actions
export const actions = {
	getPipelineSession: dispatch(moduleActions.getPipelineSession),
	endPipelineSession: dispatch(moduleActions.endPipelineSession),
	abort: dispatch(moduleActions.abort),
	exportPipeline: dispatch(moduleActions.exportPipeline)
}

// type mutators
export const mutations = {
	setPipelineSession: commit(moduleMutations.setPipelineSession)
}
