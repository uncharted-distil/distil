import { Module } from 'vuex';
import { state, PipelineState } from './index';
import { getters as moduleGetters } from './getters';
import { actions as moduleActions } from './actions';
import { mutations as moduleMutations } from './mutations';
import { DistilState } from '../store';
import { getStoreAccessors } from 'vuex-typescript';

export const pipelineModule: Module<PipelineState, DistilState> = {
	state: state,
	getters: moduleGetters,
	actions: moduleActions,
	mutations: moduleMutations
}

const { commit, read, dispatch } = getStoreAccessors<PipelineState, DistilState>(null);

export const getters = {
	getRunningPipelines: read(moduleGetters.getRunningPipelines),
	getCompletedPipelines: read(moduleGetters.getCompletedPipelines)
}

export const actions = {
	getSession: dispatch(moduleActions.getSession),
	createPipelines: dispatch(moduleActions.createPipelines),

}

export const mutations = {
	addRunningPipeline: commit(moduleMutations.addRunningPipeline),
	removeRunningPipeline: commit(moduleMutations.removeRunningPipeline),
	addCompletedPipeline: commit(moduleMutations.addCompletedPipeline),
	removeCompletedPipeline: commit(moduleMutations.removeCompletedPipeline)
}
