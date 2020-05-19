import { AppState, SatelliteBand, StatusPanelState } from "./index";

export const getters = {
  getCurrentSatelliteBand(state: AppState): SatelliteBand {
    return state.currentSatelliteBand;
  },

  getVersionNumber(state: AppState): string {
    return state.versionNumber;
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
  }
};
