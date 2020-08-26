import { Module } from "vuex";
import { state, ViewState } from "./index";
import { getters as moduleGetters } from "./getters";
import { actions as moduleActions } from "./actions";
import { mutations as moduleMutations } from "./mutations";
import { DistilState } from "../store";
import { getStoreAccessors } from "vuex-typescript";

export const viewModule: Module<ViewState, DistilState> = {
  state: state,
  actions: moduleActions,
  getters: moduleGetters,
  mutations: moduleMutations,
};

const { commit, read, dispatch } = getStoreAccessors<ViewState, DistilState>(
  null
);

export const getters = {
  getFetchParamsCache: read(moduleGetters.getFetchParamsCache),
};

export const mutations = {
  setFetchParamsCache: commit(moduleMutations.setFetchParamsCache),
};

export const actions = {
  fetchHomeData: dispatch(moduleActions.fetchHomeData),
  fetchSearchData: dispatch(moduleActions.fetchSearchData),
  fetchJoinDatasetsData: dispatch(moduleActions.fetchJoinDatasetsData),
  clearJoinDatasetsData: dispatch(moduleActions.clearJoinDatasetsData),
  updateJoinDatasetsData: dispatch(moduleActions.updateJoinDatasetsData),
  fetchSelectTargetData: dispatch(moduleActions.fetchSelectTargetData),
  fetchSelectTrainingData: dispatch(moduleActions.fetchSelectTrainingData),
  updateSelectTrainingData: dispatch(moduleActions.updateSelectTrainingData),
  fetchResultsData: dispatch(moduleActions.fetchResultsData),
  updateResultsSolution: dispatch(moduleActions.updateResultsSolution),
  fetchPredictionsData: dispatch(moduleActions.fetchPredictionsData),
  updatePrediction: dispatch(moduleActions.updatePredictions),
};
