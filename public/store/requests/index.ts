import { FilterParams } from "../../util/filters";

export enum SolutionStatus {
  SOLUTION_PENDING = "SOLUTION_PENDING",
  SOLUTION_FITTING = "SOLUTION_FITTING",
  SOLUTION_SCORING = "SOLUTION_SCORING",
  SOLUTION_PRODUCING = "SOLUTION_PRODUCING",
  SOLUTION_COMPLETED = "SOLUTION_COMPLETED",
  SOLUTION_ERRORED = "SOLUTION_ERRORED",
  SOLUTION_CANCELLED = "SOLUTION_CANCELLED",
}

export enum PredictStatus {
  PREDICT_PENDING = "PREDICT_PENDING",
  PREDICT_RUNNING = "PREDICT_RUNNING",
  PREDICT_COMPLETED = "PREDICT_COMPLETED",
  PREDICT_ERRORED = "PREDICT_ERRORED",
}

export enum QueryStatus {
  QUERY_PENDING = "QUERY_PENDING",
  QUERY_RUNNING = "QUERY_RUNNING",
  QUERY_COMPLETED = "QUERY_COMPLETED",
  QUERY_ERRORED = "QUERY_ERRORED",
}

export enum SolutionRequestStatus {
  SOLUTION_REQUEST_PENDING = "REQUEST_PENDING",
  SOLUTION_REQUEST_RUNNING = "REQUEST_RUNNING",
  SOLUTION_REQUEST_COMPLETED = "REQUEST_COMPLETED",
  SOLUTION_REQUEST_ERRORED = "REQUEST_ERRORED",
}

export const NUM_SOLUTIONS = 3;

export interface Request {
  requestId: string;
  dataset: string;
  feature: string;
  features: Feature[];
  timestamp: number;
}

// A request to start the process of training, fitting and scoring a model
export interface SolutionRequest extends Request {
  progress: SolutionStatus;
  filters: FilterParams;
}

export interface Solution extends SolutionRequest {
  solutionId: string;
  fittedSolutionId: string;
  resultId: string;
  scores: Score[];
  predictedKey: string;
  errorKey: string;
  confidenceKey: string;
  isBad: boolean;
  featureLabel: string;
}

export interface Predictions extends Request {
  progress: PredictStatus;
  fittedSolutionId: string;
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

export enum ModelQuality {
  SPEED = "speed",
  HIGHER_QUALITY = "quality",
}

export interface RequestState {
  solutionRequests: SolutionRequest[];
  solutions: Solution[];
  predictions: Predictions[];
}

export const state: RequestState = {
  solutionRequests: [],
  solutions: [],
  predictions: [],
};
