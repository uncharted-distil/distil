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
	hasActiveSession: read(moduleGetters.hasActiveSession),
	getRunningPipelines: read(moduleGetters.getRunningPipelines),
	getCompletedPipelines: read(moduleGetters.getCompletedPipelines),
	getPipelines: read(moduleGetters.getPipelines),
	getPipelineRequestIds: read(moduleGetters.getPipelineRequestIds),
	getActivePipeline: read(moduleGetters.getActivePipeline),
	getActivePipelineTrainingMap: read(moduleGetters.getActivePipelineTrainingMap),
	getActivePipelineVariables: read(moduleGetters.getActivePipelineVariables),
}

export const actions = {
	startPipelineSession: dispatch(moduleActions.startPipelineSession),
	endPipelineSession: dispatch(moduleActions.endPipelineSession),
	fetchPipelines: dispatch(moduleActions.fetchPipelines),
	createPipelines: dispatch(moduleActions.createPipelines),
}

export const mutations = {
	setPipelineSessionID: commit(moduleMutations.setPipelineSessionID),
	setSessionActivity: commit(moduleMutations.setSessionActivity),
	updatePipelineRequests: commit(moduleMutations.updatePipelineRequests),
	clearPipelineRequests: commit(moduleMutations.clearPipelineRequests)
}
