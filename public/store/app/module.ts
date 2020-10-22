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
  mutations: moduleMutations,
};

const { commit, read, dispatch } = getStoreAccessors<AppState, DistilState>(
  null
);

// typed getters
export const getters = {
  getVersionNumber: read(moduleGetters.getVersionNumber),
  getHelpURL: read(moduleGetters.getHelpURL),
  getVersionTimestamp: read(moduleGetters.getVersionTimestamp),
  getProblemDataset: read(moduleGetters.getProblemDataset),
  getProblemTarget: read(moduleGetters.getProblemTarget),
  getProblemMetrics: read(moduleGetters.getProblemMetrics),
  getStatusPanelState: read(moduleGetters.getStatusPanelState),
  getTA2VersionNumber: read(moduleGetters.getTA2VersionNumber),
  getAllSystemVersions: read(moduleGetters.getAllSystemVersions),
  getSessionToken: read(moduleGetters.getSessionToken),
  getMapAPIKey: read(moduleGetters.getMapAPIKey),
  getTileRequestURL: read(moduleGetters.getTileRequestURL),
  getSubdomains: read(moduleGetters.getSubdomains),
};

// typed actions
export const actions = {
  saveModel: dispatch(moduleActions.saveModel),
  exportSolution: dispatch(moduleActions.exportSolution),
  exportProblem: dispatch(moduleActions.exportProblem),
  fetchConfig: dispatch(moduleActions.fetchConfig),
  openStatusPanelWithContentType: dispatch(
    moduleActions.openStatusPanelWithContentType
  ),
  closeStatusPanel: dispatch(moduleActions.closeStatusPanel),
  logUserEvent: dispatch(moduleActions.logUserEvent),
};

// type mutators
export const mutations = {
  setVersionNumber: commit(moduleMutations.setVersionNumber),
  setHelpURL: commit(moduleMutations.setHelpURL),
  setVersionTimestamp: commit(moduleMutations.setVersionTimestamp),
  setProblemDataset: commit(moduleMutations.setProblemDataset),
  setProblemTarget: commit(moduleMutations.setProblemTarget),
  setProblemMetrics: commit(moduleMutations.setProblemMetrics),
  setStatusPanelContentType: commit(moduleMutations.setStatusPanelContentType),
  setTA2VersionNumber: commit(moduleMutations.setTA2VersionNumber),
  openStatusPanel: commit(moduleMutations.openStatusPanel),
  closeStatusPanel: commit(moduleMutations.closeStatusPanel),
  setSessionToken: commit(moduleMutations.setSessionToken),
};
