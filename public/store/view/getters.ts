import { ViewState } from "./index";

export const getters = {
  getFetchParamsCache(state: ViewState) {
    return state.fetchParamsCache;
  },
};
