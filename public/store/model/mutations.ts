import Vue from "vue";
import { Model, ModelState } from "./index";

export const mutations = {
  setModels(state: ModelState, models: Model[]) {
    // individually add models if they do not exist
    if (models) {
      const lookup = {};
      state.models.forEach((m, index) => {
        lookup[m.fittedSolutionId] = index;
      });
      models.forEach(m => {
        const index = lookup[m.fittedSolutionId];
        if (index !== undefined) {
          // update if it already exists
          Vue.set(state.models, index, m);
        } else {
          state.models.push(m);
        }
      });
    } else {
      Vue.set(state, "models", []);
    }
  }
};
