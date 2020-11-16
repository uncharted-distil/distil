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

  getTA2VersionNumber(state: AppState): string {
    return state.ta2Version;
  },

  getAllSystemVersions(state: AppState): string {
    const ta2Version = state.ta2Version;
    const ta3Version = state.versionNumber;
    const ta3Timestamp =
      state.versionTimestamp !== "unset" ? state.versionTimestamp : "";
    return `TA2 version ${ta2Version}\nTA3 version ${ta3Version} ${ta3Timestamp}`.trim();
  },
  getTrainTestSplit(state: AppState): number {
    return state.trainTestSplit;
  },
  getTrainTestSplitTimeSeries(state: AppState): number {
    return state.trainTestSplitTimeSeries;
  },
};
