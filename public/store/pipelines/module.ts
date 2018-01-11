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
	getPipelineSessionID: read(moduleGetters.getPipelineSessionID),
	getRunningPipelines: read(moduleGetters.getRunningPipelines),
	getCompletedPipelines: read(moduleGetters.getCompletedPipelines),
	getPipelines: read(moduleGetters.getPipelines)
}

export const actions = {
	startPipelineSession: dispatch(moduleActions.startPipelineSession),
	endPipelineSession: dispatch(moduleActions.endPipelineSession),
	getSessionSummary: dispatch(moduleActions.getSessionSummary),
	createPipelines: dispatch(moduleActions.createPipelines),
}

export const mutations = {
	setPipelineSessionID: commit(moduleMutations.setPipelineSessionID),
	addRunningPipeline: commit(moduleMutations.addRunningPipeline),
	removeRunningPipeline: commit(moduleMutations.removeRunningPipeline),
	addCompletedPipeline: commit(moduleMutations.addCompletedPipeline),
	removeCompletedPipeline: commit(moduleMutations.removeCompletedPipeline)
}
