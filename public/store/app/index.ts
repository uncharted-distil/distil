import { DatasetPendingRequestType } from "../dataset/index";

export interface AppState {
  currentSatelliteBand: SatelliteBand;
  versionNumber: string;
  versionTimestamp: string;
  problemDataset: string;
  problemTarget: string;
  problemMetrics: string[];
  statusPanelState: StatusPanelState;
}

export interface SatelliteBand {
  r: number;
  g: number;
  b: number;
}

export interface StatusPanelState {
  isOpen: boolean;
  contentType: StatusPanelContentType;
}

// shared data model
export const state: AppState = {
  currentSatelliteBand: { r: 4, g: 3, b: 2 },
  versionNumber: "unknown",
  versionTimestamp: "unknown",
  problemDataset: "unknown",
  problemTarget: "unknown",
  problemMetrics: [],
  statusPanelState: {
    contentType: undefined,
    isOpen: false
  }
};

export type StatusPanelContentType = DatasetPendingRequestType;
