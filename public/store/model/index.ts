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
  models: Model[];
}

export const state: ModelState = {
  models: []
};
