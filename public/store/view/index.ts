export const LAST_STATE = "__LAST_STATE__";
import { Dictionary } from "../../util/dict";

export interface ViewState {
  fetchParamsCache: Dictionary<string>;
}

export const state: ViewState = {
  fetchParamsCache: {},
};
