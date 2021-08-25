import { requestGetters, viewActions } from "../../../store";
import { getters as routeGetters } from "../../../store/route/module";
import store from "../../../store/store";

export const PREDICTION_COMPUTES = {
  produceRequestId: (): string => {
    return routeGetters.getRouteProduceRequestId(store);
  },
};
export const PREDICTION_METHODS = {
  fetchSummaryPrediction: async (id: string): Promise<void> => {
    const predictions = requestGetters
      .getRelevantPredictions(store)
      .filter((p) => {
        return p.requestId === id;
      });
    viewActions.updatePredictionSummaries(store, {
      predictions: predictions,
    });
  },
};
