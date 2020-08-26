import { ViewState } from "./index";

export const mutations = {
  setFetchParamsCache(state: ViewState, args: { key: string; value: string }) {
    state.fetchParamsCache[args.key] = args.value;
  },
};
