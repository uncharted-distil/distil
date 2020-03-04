import { FilterParams } from "../../util/filters";

export const SOLUTION_PENDING = "SOLUTION_PENDING";
export const SOLUTION_FITTING = "SOLUTION_FITTING";
export const SOLUTION_SCORING = "SOLUTION_SCORING";
export const SOLUTION_PRODUCING = "SOLUTION_PRODUCING";
export const SOLUTION_COMPLETED = "SOLUTION_COMPLETED";
export const SOLUTION_ERRORED = "SOLUTION_ERRORED";

export const PREDICTION_PENDING = "PREDICTION_PENDING";
export const PREDICTION_RUNNING = "PREDICTION_RUNNING";
export const PREDICTION_COMPLETED = "PREDICTION_COMPLETED";
export const PREDICTION_ERRORED = "PREDICTION_ERRORED";

export const REQUEST_PENDING = "REQUEST_PENDING";
export const REQUEST_RUNNING = "REQUEST_RUNNING";
export const REQUEST_COMPLETED = "REQUEST_COMPLETED";
export const REQUEST_ERRORED = "REQUEST_ERRORED";

export const NUM_SOLUTIONS = 3;

export interface Request {
  requestId: string;
  progress: string;
  timestamp: number;
}

export interface SolutionRequest extends Request {
  dataset: string;
  feature: string;
  filters: FilterParams;
  features: Feature[];
}

export interface PredictRequest extends Request {
  fittedSolutionId: string;
}

export interface Solution extends SolutionRequest {
  solutionId: string;
  fittedSolutionId: string;
  resultId: string;
  scores: Score[];
  predictedKey: string;
  errorKey: string;
  isBad: boolean;
}

export interface Predictions extends PredictRequest {
  resultId: string;
  predictedKey: string;
  isBad: boolean;
}

export interface Score {
  metric: string;
  label: string;
  value: number;
  sortMultiplier: number;
}

export interface Feature {
  featureName: string;
  featureType: string;
}

export interface RequestState {
  solutionRequests: SolutionRequest[];
  solutions: Solution[];
  predictRequests: PredictRequest[];
  predictions: Predictions[];
}

export const state: RequestState = {
  solutionRequests: [],
  solutions: [],
  predictRequests: [],
  predictions: []
};
