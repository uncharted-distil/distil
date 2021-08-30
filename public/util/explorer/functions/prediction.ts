import { requestGetters, viewActions } from "../../../store";
import { getters as routeGetters } from "../../../store/route/module";
import store from "../../../store/store";

/**
 * PREDICTION_COMPUTES contains all of the computes for the prediction state in the data explorer
 **/
export const PREDICTION_COMPUTES = {
  /**
   * produceRequestId the prediction request id of the current active
   */
  produceRequestId: (): string => {
    return routeGetters.getRouteProduceRequestId(store);
  },
};
export const PREDICTION_METHODS = {
  /**
   * fetches the prediction summaries which is used in the PredictionSummaries
   */
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
