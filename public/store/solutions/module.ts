import { Module } from "vuex";
import { state, SolutionState } from "./index";
import { getters as moduleGetters } from "./getters";
import { actions as moduleActions } from "./actions";
import { mutations as moduleMutations } from "./mutations";
import { DistilState } from "../store";
import { getStoreAccessors } from "vuex-typescript";

export const solutionModule: Module<SolutionState, DistilState> = {
  state: state,
  getters: moduleGetters,
  actions: moduleActions,
  mutations: moduleMutations
};

const { commit, read, dispatch } = getStoreAccessors<
  SolutionState,
  DistilState
>(null);

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
  createSolutionRequest: dispatch(moduleActions.createSolutionRequest),
  stopSolutionRequest: dispatch(moduleActions.stopSolutionRequest)
};

export const mutations = {
  updateSolutionRequests: commit(moduleMutations.updateSolutionRequests),
  clearSolutionRequests: commit(moduleMutations.clearSolutionRequests)
};
