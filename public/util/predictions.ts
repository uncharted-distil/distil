import { Predictions } from "../store/requests";

export function getPredictionsById(
  predictions: Predictions[],
  predictionsId: string
): Predictions {
  const id = predictionsId || "";
  return predictions.find(r => r.requestId === id);
}
