import { FilterParams } from "../../util/filters";

export const SOLUTION_PENDING = "SOLUTION_PENDING";
export const SOLUTION_FITTING = "SOLUTION_FITTING";
export const SOLUTION_SCORING = "SOLUTION_SCORING";
export const SOLUTION_PRODUCING = "SOLUTION_PRODUCING";
export const SOLUTION_COMPLETED = "SOLUTION_COMPLETED";
export const SOLUTION_ERRORED = "SOLUTION_ERRORED";

export const PREDICT_PENDING = "PREDICT_PENDING";
export const PREDICT_RUNNING = "PREDICT_RUNNING";
export const PREDICT_COMPLETED = "PREDICT_COMPLETED";
export const PREDICT_ERRORED = "PREDICT_ERRORED";

export const SOLUTION_REQUEST_PENDING = "REQUEST_PENDING";
export const SOLUTION_REQUEST_RUNNING = "REQUEST_RUNNING";
export const SOLUTION_REQUEST_COMPLETED = "REQUEST_COMPLETED";
export const SOLUTION_REQUEST_ERRORED = "REQUEST_ERRORED";

export const NUM_SOLUTIONS = 3;

export interface Request {
  requestId: string;
  progress: string;
  timestamp: number;
}

// A request to start the process of training, fitting and scoring a model
export interface SolutionRequest extends Request {
  dataset: string;
  feature: string;
  filters: FilterParams;
  features: Feature[];
}

// A request to generate predictions from fitted a model
export interface PredictRequest extends Request {
  fittedSolutionId: string;
  dataset: string;
  feature: string;
}

// A
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
