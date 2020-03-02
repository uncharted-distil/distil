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
  getRelevantSolutionRequests: read(moduleGetters.getRelevantSolutionRequests),
  getRelevantSolutionRequestIds: read(
    moduleGetters.getRelevantSolutionRequestIds
  ),
  getActiveSolution: read(moduleGetters.getActiveSolution),
  getActiveSolutionTrainingVariables: read(
    moduleGetters.getActiveSolutionTrainingVariables
  ),
  getActiveSolutionTargetVariable: read(
    moduleGetters.getActiveSolutionTargetVariable
  )
};

export const actions = {
  fetchSolutionRequests: dispatch(moduleActions.fetchSolutionRequests),
  fetchSolutionRequest: dispatch(moduleActions.fetchSolutionRequest),
  createSolutionRequest: dispatch(moduleActions.createSolutionRequest),
  stopSolutionRequest: dispatch(moduleActions.stopSolutionRequest),
  fetchSolutions: dispatch(moduleActions.fetchSolutions),
  fetchSolution: dispatch(moduleActions.fetchSolution)
};

export const mutations = {
  updateSolutionRequests: commit(moduleMutations.updateSolutionRequests),
  updateSolutions: commit(moduleMutations.updateSolutions),
  clearSolutionRequests: commit(moduleMutations.clearSolutionRequests),
  clearSolutions: commit(moduleMutations.clearSolutions)
};
