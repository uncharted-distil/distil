import { AppState, StatusPanelState } from "./index";

export const getters = {
  getVersionNumber(state: AppState): string {
    return state.versionNumber;
  },

  getHelpURL(state: AppState): string {
    return state.helpURL;
  },

  getVersionTimestamp(state: AppState): string {
    return state.versionTimestamp;
  },

  getProblemDataset(state: AppState): string {
    return state.problemDataset;
  },

  getProblemTarget(state: AppState): string {
    return state.problemTarget;
  },

  getProblemMetrics(state: AppState): string[] {
    return state.problemMetrics;
  },

  getStatusPanelState(state: AppState): StatusPanelState {
    return state.statusPanelState;
  },
};
