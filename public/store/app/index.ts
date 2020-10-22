import { DatasetPendingRequestType } from "../dataset/index";

export interface AppState {
  versionNumber: string;
  helpURL: string;
  versionTimestamp: string;
  problemDataset: string;
  problemTarget: string;
  problemMetrics: string[];
  statusPanelState: StatusPanelState;
  ta2Version: string;
  sessionToken: string;
  mapAPIKey: string;
  tileRequestURL: string;
  subdomains: string;
}

export interface StatusPanelState {
  isOpen: boolean;
  contentType: StatusPanelContentType;
}

// shared data model
export const state: AppState = {
  versionNumber: "unknown",
  versionTimestamp: "unknown",
  problemDataset: "unknown",
  problemTarget: "unknown",
  helpURL: "",
  problemMetrics: [],
  statusPanelState: {
    contentType: undefined,
    isOpen: false,
  },
  ta2Version: "unknown",
  sessionToken: "",
  mapAPIKey: "",
  tileRequestURL: "",
  subdomains: "",
};

export type StatusPanelContentType = DatasetPendingRequestType;
