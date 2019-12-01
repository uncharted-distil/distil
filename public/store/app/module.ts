import { Module } from "vuex";
import { DistilState } from "../store";
import { state, AppState } from "./index";
import { getters as moduleGetters } from "./getters";
import { actions as moduleActions } from "./actions";
import { mutations as moduleMutations } from "./mutations";
import { getStoreAccessors } from "vuex-typescript";

export const appModule: Module<AppState, DistilState> = {
  state: state,
  getters: moduleGetters,
  actions: moduleActions,
  mutations: moduleMutations
};

const { commit, read, dispatch } = getStoreAccessors<AppState, DistilState>(
  null
);

// typed getters
export const getters = {
  getVersionNumber: read(moduleGetters.getVersionNumber),
  getVersionTimestamp: read(moduleGetters.getVersionTimestamp),
  isTask1: read(moduleGetters.isTask1),
  isTask2: read(moduleGetters.isTask2),
  getProblemDataset: read(moduleGetters.getProblemDataset),
  getProblemTarget: read(moduleGetters.getProblemTarget),
  getProblemMetrics: read(moduleGetters.getProblemMetrics),
  getStatusPanelState: read(moduleGetters.getStatusPanelState)
};

// typed actions
export const actions = {
  exportSolution: dispatch(moduleActions.exportSolution),
  exportProblem: dispatch(moduleActions.exportProblem),
  fetchConfig: dispatch(moduleActions.fetchConfig),
  openStatusPanelWithContentType: dispatch(
    moduleActions.openStatusPanelWithContentType
  ),
  closeStatusPanel: dispatch(moduleActions.closeStatusPanel),
  logUserEvent: dispatch(moduleActions.logUserEvent)
};

// type mutators
export const mutations = {
  setVersionNumber: commit(moduleMutations.setVersionNumber),
  setVersionTimestamp: commit(moduleMutations.setVersionTimestamp),
  setIsTask1: commit(moduleMutations.setIsTask1),
  setIsTask2: commit(moduleMutations.setIsTask2),
  setProblemDataset: commit(moduleMutations.setProblemDataset),
  setProblemTarget: commit(moduleMutations.setProblemTarget),
  setProblemMetrics: commit(moduleMutations.setProblemMetrics),
  setStatusPanelContentType: commit(moduleMutations.setStatusPanelContentType),
  openStatusPanel: commit(moduleMutations.openStatusPanel),
  closeStatusPanel: commit(moduleMutations.closeStatusPanel)
};
