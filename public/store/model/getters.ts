import { Model, ModelState } from "./index";
import { isInteger, values } from "lodash";

export const getters = {
  getFilteredModels(state: ModelState): Model[] {
    return state.filteredModelIds.map((m) => state.models[m]);
  },

  getModels(state: ModelState): Model[] {
    return values(state.models);
  },

  getCountOfModels(state: ModelState): number {
    const count = values(state.models).length;
    return isInteger(count) ? count : 0;
  },
};
