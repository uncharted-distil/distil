import { Module } from "vuex";
import { state, RequestState } from "./index";
import { getters as moduleGetters } from "./getters";
import { actions as moduleActions } from "./actions";
import { mutations as moduleMutations } from "./mutations";
import { DistilState } from "../store";
import { getStoreAccessors } from "vuex-typescript";

export const requestsModule: Module<RequestState, DistilState> = {
  state: state,
  getters: moduleGetters,
  actions: moduleActions,
  mutations: moduleMutations
};

const { commit, read, dispatch } = getStoreAccessors<RequestState, DistilState>(
  null
);

export const getters = {
  getRunningSolutions: read(moduleGetters.getRunningSolutions),
  getCompletedSolutions: read(moduleGetters.getCompletedSolutions),
  getSolutions: read(moduleGetters.getSolutions),
  getRelevantSolutions: read(moduleGetters.getRelevantSolutions),
  getRelevantSearchRequests: read(moduleGetters.getRelevantSearchRequests),
  getRelevantSearchRequestIds: read(moduleGetters.getRelevantSearchRequestIds),
  getActiveSolution: read(moduleGetters.getActiveSolution),
  getActiveSolutionTrainingVariables: read(
    moduleGetters.getActiveSolutionTrainingVariables
  ),
  getActiveSolutionTargetVariable: read(
    moduleGetters.getActiveSolutionTargetVariable
  )
};

export const actions = {
  fetchSearchRequests: dispatch(moduleActions.fetchSearchRequests),
  fetchSearchRequest: dispatch(moduleActions.fetchSearchRequest),
  createSearchRequest: dispatch(moduleActions.createSearchRequest),
  stopSearchRequest: dispatch(moduleActions.stopSearchRequest),
  fetchSolutions: dispatch(moduleActions.fetchSolutions),
  fetchSolution: dispatch(moduleActions.fetchSolution)
};

export const mutations = {
  updateSearchRequests: commit(moduleMutations.updateSearchRequests),
  updateSolutions: commit(moduleMutations.updateSolutions),
  clearSearchRequests: commit(moduleMutations.clearSearchRequests),
  clearSolutions: commit(moduleMutations.clearSolutions)
};
