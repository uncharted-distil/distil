import { Dictionary } from "vue-router/types/router";

export interface Model {
  modelName: string;
  modelDescription: string;
  filePath: string;
  fittedSolutionId: string;
  datasetId: string;
  datasetName: string;
  target: string;
  variables: string[];
}

export interface ModelState {
  // list of of all saved models, keyed by fitted solution id
  models: Dictionary<Model>;
  // fitted solution id of models that are currently being filtered
  filteredModelIds: string[];
}

export const state: ModelState = {
  models: {},
  filteredModelIds: [],
};
