import { Model, ModelState } from "./index";

export const getters = {
  getModels(state: ModelState): Model[] {
    return state.models;
  }
};
