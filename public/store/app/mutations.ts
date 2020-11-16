import { AppState, StatusPanelContentType } from "./index";
import Vue from "vue";

export const mutations = {
  setVersionNumber(state: AppState, versionNumber: string) {
    state.versionNumber = versionNumber;
  },

  setHelpURL(state: AppState, helpURL: string) {
    state.helpURL = helpURL;
  },

  setVersionTimestamp(state: AppState, versionTimestamp: string) {
    state.versionTimestamp = versionTimestamp;
  },

  setProblemDataset(state: AppState, dataset: string) {
    state.problemDataset = dataset;
  },

  setProblemTarget(state: AppState, target: string) {
    state.problemTarget = target;
  },

  setProblemMetrics(state: AppState, metrics: string[]) {
    state.problemMetrics = metrics;
  },

  setStatusPanelContentType(
    state: AppState,
    contentType: StatusPanelContentType
  ) {
    Vue.set(state.statusPanelState, "contentType", contentType);
  },

  setTA2VersionNumber(state: AppState, ta2Version: string) {
    state.ta2Version = ta2Version;
  },

  setTrainTestSplit(state: AppState, trainTestSplit: string) {
    state.trainTestSplit = parseFloat(trainTestSplit);
  },

  setTrainTestSplitTimeSeries(
    state: AppState,
    trainTestSplitTimeSeries: string
  ) {
    state.trainTestSplitTimeSeries = parseFloat(trainTestSplitTimeSeries);
  },

  openStatusPanel(state: AppState) {
    Vue.set(state.statusPanelState, "isOpen", true);
  },

  closeStatusPanel(state: AppState) {
    Vue.set(state.statusPanelState, "isOpen", false);
  },
};
