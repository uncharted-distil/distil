import { Model, ModelState } from "./index";
import _ from "lodash";

export const getters = {
  getFilteredModels(state: ModelState): Model[] {
    return state.filteredModelIds.map(m => state.models[m]);
  },

  getModels(state: ModelState): Model[] {
    return _.values(state.models);
  }
};
