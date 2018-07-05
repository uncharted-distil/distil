import { Module } from 'vuex';
import { HighlightState, state } from './index';
import { getters as moduleGetters } from './getters';
import { actions as moduleActions } from './actions';
import { mutations as moduleMutations } from './mutations';
import { DistilState } from '../store';
import { getStoreAccessors } from 'vuex-typescript';

export const highlightsModule: Module<HighlightState, DistilState> = {
	getters: moduleGetters,
	actions: moduleActions,
	mutations: moduleMutations,
	state: state
}

const { commit, read, dispatch } = getStoreAccessors<HighlightState, DistilState>(null);

// Typed getters
export const getters = {
	// highlights
	getHighlightedSummaries: read(moduleGetters.getHighlightedSummaries)
}

// Typed actions
export const actions = {
	// highlight values
	fetchDataHighlightValues: dispatch(moduleActions.fetchDataHighlightValues),
	fetchResultHighlightValues: dispatch(moduleActions.fetchResultHighlightValues)
}

// Typed mutations
export const mutations = {
	updateHighlightSummaries: commit(moduleMutations.updateHighlightSummaries),
	clearHighlightSummaries: commit(moduleMutations.clearHighlightSummaries)
}
