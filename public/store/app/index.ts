import { DatasetPendingRequestType } from "../dataset/index";

export interface AppState {
  versionNumber: string;
  versionTimestamp: string;
  isTask1: boolean;
  isTask2: boolean;
  problemDataset: string;
  problemTarget: string;
  problemMetrics: string[];
  statusPanelState: StatusPanelState;
}

export interface StatusPanelState {
  isOpen: boolean;
  contentType: StatusPanelContentType;
}

// shared data model
export const state: AppState = {
  versionNumber: "unknown",
  versionTimestamp: "unknown",
  isTask1: false,
  isTask2: false,
  problemDataset: "unknown",
  problemTarget: "unknown",
  problemMetrics: [],
  statusPanelState: {
    contentType: undefined,
    isOpen: false
  }
};

export type StatusPanelContentType = DatasetPendingRequestType;
