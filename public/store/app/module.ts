import { Module } from 'vuex';
import { state, AppState } from './index';
import { getters as moduleGetters } from './getters';
import { actions as moduleActions} from './actions';
import { mutations as moduleMutations } from './mutations';
import { getStoreAccessors } from 'vuex-typescript';

export const appModule: Module<AppState, any> = {
	state: state,
	getters: moduleGetters,
	actions: moduleActions,
	mutations: moduleMutations
}

const { commit, read, dispatch } = getStoreAccessors<AppState, any>(null);

// typed getters
export const getters = {
	getWebSocketConnection: read(moduleGetters.getWebSocketConnection),
	getPipelineSessionID: read(moduleGetters.getPipelineSessionID),
	getPipelineSession: read(moduleGetters.getPipelineSession),
	getRecentDatasets: read(moduleGetters.getRecentDatasets)
}

// typed actions
export const actions = {
	getPipelineSession: dispatch(moduleActions.getPipelineSession),
	endPipelineSession: dispatch(moduleActions.endPipelineSession),
	abort: dispatch(moduleActions.abort),
	exportPipeline: dispatch(moduleActions.exportPipeline),
	addRecentDataset: dispatch(moduleActions.addRecentDataset)
}

// type mutators
export const mutations = {
	setWebSocketConnection: commit(moduleMutations.setWebSocketConnection),
	setPipelineSession: commit(moduleMutations.setPipelineSession),
	addRecentDataset: commit(moduleMutations.addRecentDataset)
}
