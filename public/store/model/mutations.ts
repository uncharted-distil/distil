import { Model, ModelState } from "./index";
import _ from "lodash";

export const mutations = {
  // replace the map of saved models
  setModels(state: ModelState, models: Model[]) {
    state.models = _.keyBy(models, (m) => m.fittedSolutionId);
  },

  // replace the list of filtered models (a subset of the full saved model list)
  setFilteredModels(state: ModelState, models: Model[]) {
    if (!models) return;
    state.filteredModelIds = models.map((m) => m.fittedSolutionId);
  },
};
