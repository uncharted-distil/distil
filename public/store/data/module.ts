import { Module } from 'vuex';
import { state, DataState } from './index';
import { getters } from './getters';
import { actions } from './actions';
import { mutations } from './mutations';
import { getStoreAccessors } from 'vuex-typescript';

export const dataModule: Module<DataState, any> = {
	getters: getters,
	actions: actions,
	mutations: mutations,
	state: state
}

const { commit, read, dispatch } = getStoreAccessors<DataState, any>(null); 

// Typed getters
export const getVariables = read(getters.getVariables);
export const getVariablesMap = read(getters.getVariablesMap);
export const getDatasets = read(getters.getDatasets);
export const getAvailableVariables = read(getters.getAvailableVariables);
export const getAvailableVariablesMap = read(getters.getAvailableVariablesMap);
export const getTrainingVariablesMap = read(getters.getTrainingVariablesMap);
export const getVariableSummaries = read(getters.getVariableSummaries);
export const getResultsSummaries = read(getters.getResultsSummaries);
export const getSelectedFilters = read(getters.getSelectedFilters);
export const getAvailableVariableSummaries = read(getters.getAvailableVariableSummaries);
export const getTrainingVariableSummaries = read(getters.getTrainingVariableSummaries);
export const getTargetVariableSummaries = read(getters.getTargetVariableSummaries);
export const getFilteredData = read(getters.getFilteredData);
export const getFilteredDataItems = read(getters.getFilteredDataItems);
export const getFilteredDataFields = read(getters.getFilteredDataFields);
export const getResultData = read(getters.getResultData);
export const getResultDataItems = read(getters.getResultDataItems);
export const getResultDataFields = read(getters.getResultDataFields);
export const getSelectedData = read(getters.getSelectedData);
export const getSelectedDataItems = read(getters.getSelectedDataItems);
export const getSelectedDataFields = read(getters.getSelectedDataFields);
export const getHighlightedFeatureValues = read(getters.getHighlightedFeatureValues);

// Typed actions

// Typed mutations
