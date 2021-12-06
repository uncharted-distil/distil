import {
  datasetActions,
  datasetGetters,
  requestGetters,
  viewActions,
} from "../../../store";
import { DataMode, SummaryMode } from "../../../store/dataset";
import { getters as routeGetters } from "../../../store/route/module";
import store from "../../../store/store";
import { DataExplorerRef } from "../../componentTypes";
import { EventList } from "../../events";
import { overlayRouteEntry, varModesToString } from "../../routes";
import { IMAGE_TYPE, isClusterType } from "../../types";

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

export const PREDICTION_EVENT_HANDLERS = {
  /**
   * This function handles the apply cluster event
   * It updates the variable with a cluster column
   * then updates the variable summary with the new cluster datamode
   */
  [EventList.VARIABLES.APPLY_CLUSTER_EVENT]: async function () {
    const self = (this as unknown) as DataExplorerRef;
    // fetch the var modes map
    const varModesMap = routeGetters.getDecodedVarModes(store);
    const clusterVars = new Set<string>();
    await self.state.fetchVariables();
    // find any grouped vars that are using this cluster data and update their
    // mode to cluster now that data is available
    datasetGetters
      .getGroupings(store)
      .filter((v) => isClusterType(v.colType) && v.datasetName === self.dataset)
      .forEach((v) => {
        if (v.grouping.clusterCol) {
          varModesMap.set(v.key, SummaryMode.Cluster);
          clusterVars.add(v.grouping.clusterCol);
        }
      });

    // find any image variables using this cluster data and update their mode
    datasetGetters
      .getVariables(store)
      .filter((v) => v.colType === IMAGE_TYPE && v.datasetName === self.dataset)
      .forEach((v) => {
        varModesMap.set(v.key, SummaryMode.Cluster);
      });

    // serialize the modes map into a string and add to the route
    // and update to know that the clustering has been applied.
    const varModesStr = varModesToString(varModesMap);
    const entry = overlayRouteEntry(self.$route, {
      varModes: varModesStr,
      dataMode: DataMode.Cluster,
      clustering: "1",
      explore: [
        ...routeGetters.getExploreVariables(store),
        ...clusterVars,
      ].join(","),
    });
    self.$router.push(entry).catch((err) => console.warn(err));
    // fetch the new summaries with the clustering applied
    viewActions.updatePredictionTrainingSummaries(store);
    return;
  },
  /**
   * This handles outlier events for the select state
   * All it does is apply the outlier to the ds
   * then update the variables / variable summaries
   * **/
  [EventList.VARIABLES.APPLY_OUTLIER_EVENT]: async function (
    callback?: Function
  ) {
    const self = (this as unknown) as DataExplorerRef;
    const dataset = self.dataset;
    const success = await datasetActions.applyOutliers(store, dataset);
    if (!success) return;

    // Update the variables, which should now include the outlier variable.
    await datasetActions.fetchVariables(store, {
      dataset,
    });
    await viewActions.updatePredictionTrainingSummaries(store);

    // Update the route to know that the outlier has been applied.
    const entry = overlayRouteEntry(self.$route, { outlier: "1" });
    self.$router.push(entry).catch((err) => console.warn(err));
    if (callback) {
      callback();
    }
    return;
  },
};
