import { VariableSummary } from "../store/dataset";
import store from "../store/store";
import { getters as resultGetters } from "../store/results/module";
import { getters as predictionGetters } from "../store/predictions/module";

export function getIDFromKey(key: string): string {
  return key ? key.split(":")[0] : "";
}

export function getSolutionResultSummary(solutionID: string): VariableSummary {
  return resultGetters
    .getPredictedSummaries(store)
    .find((s) => getIDFromKey(s.key) === solutionID);
}

export function getResidualSummary(solutionID: string): VariableSummary {
  return resultGetters
    .getResidualsSummaries(store)
    .find((s) => getIDFromKey(s.key) === solutionID);
}

export function getCorrectnessSummary(solutionID: string): VariableSummary {
  return resultGetters
    .getCorrectnessSummaries(store)
    .find((s) => getIDFromKey(s.key) === solutionID);
  return null;
}

export function getPredictionResultSummary(requestId: string): VariableSummary {
  return predictionGetters
    .getPredictionSummaries(store)
    .find((s) => getIDFromKey(s.key) === requestId);
}

export function getConfidenceSummary(solutionID: string): VariableSummary {
  return resultGetters
    .getConfidenceSummaries(store)
    .find((s) => getIDFromKey(s.key) === solutionID);
  return null;
}
