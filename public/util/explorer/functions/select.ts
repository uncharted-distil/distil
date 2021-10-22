import { CreateSolutionsFormRef, DataExplorerRef } from "../../componentTypes";
import { isEmpty, isNil } from "lodash";
import {
  appActions,
  datasetActions,
  datasetGetters,
  requestActions,
  viewActions,
} from "../../../store";
import { getters as routeGetters } from "../../../store/route/module";
import store from "../../../store/store";
import { SolutionRequestMsg } from "../../../store/requests/actions";
import { Solution } from "../../../store/requests";
import { DataMode, SummaryMode } from "../../../store/dataset";
import { overlayRouteEntry, varModesToString } from "../../routes";
import { ExplorerStateNames } from "..";
import { createFiltersFromHighlights } from "../../highlights";
import { addFilterToRoute, EXCLUDE_FILTER } from "../../filters";
import { Activity, Feature, SubActivity } from "../../userEvents";
import { EventList } from "../../events";
import { IMAGE_TYPE, isClusterType } from "../../types";
import { $enum } from "ts-enum-util";

/**
 * SELECT_COMPUTES contains all of the computes for the select state in the data explorer
 **/
export const SELECT_COMPUTES = {
  /**
   * checks to see if the UI surrounding include/exclude table views is enabled
   * this should only be true for the select state
   */
  isExcludeDisabled(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    if (!self) {
      return false;
    }
    return !self.isFilteringHighlights && !self.isFilteringSelection;
  },
  /**
   * returns the variable keys that are currently selected as training features does not include target
   */
  training: (): string[] => {
    return routeGetters.getDecodedTrainingVariableNames(store);
  },
  /**
   * returns true if a target variable is selected and at least 1 training feature has been selected
   * this enables the create model button
   */
  isCreateModelPossible(): boolean {
    const self = (this as unknown) as DataExplorerRef;
    if (!self) {
      return false;
    }
    // check that we have some target and training variables.
    return !isNil(self.target) && !isEmpty(self.training);
  },
};
/**
 * SELECT_METHODS contains the methods used in the select state in the data explorer
 */
export const SELECT_METHODS = {
  /**
   * onModelCreation starts the process for fitting a model
   */
  onModelCreation(msg: SolutionRequestMsg): void {
    const self = (this as unknown) as DataExplorerRef;
    msg.filters.variables = routeGetters
      .getRouteTrainingVariables(store)
      .split(",")
      .concat(routeGetters.getRouteTargetVariable(store));
    requestActions
      .createSolutionRequest(store, msg)
      .then(async (res: Solution) => {
        const dataMode = routeGetters.getDataMode(store);
        const dataModeDefault = dataMode ? dataMode : DataMode.Default;
        // update route with the result params
        self.updateRoute({
          dataset: routeGetters.getRouteDataset(store),
          target: routeGetters.getRouteTargetVariable(store),
          solutionId: res.solutionId,
          task: routeGetters.getRouteTask(store),
          dataMode: dataModeDefault,
          varModes: varModesToString(routeGetters.getDecodedVarModes(store)),
          modelLimit: routeGetters.getModelLimit(store),
          modelTimeLimit: routeGetters.getModelTimeLimit(store),
          modelQuality: routeGetters.getModelQuality(store),
        });
        const modelCreationRef = (self.$refs[
          "model-creation-form"
        ] as unknown) as CreateSolutionsFormRef;
        // model creation is successful stop spinner
        modelCreationRef.success();
        // change data explorer state to RESULT_VIEW
        await self.changeStatesByName(ExplorerStateNames.RESULT_VIEW);
      })
      .catch((err) => {
        const modelCreationRef = (self.$refs[
          "model-creation-form"
        ] as unknown) as CreateSolutionsFormRef;
        modelCreationRef.fail(err);
        console.error(err);
      });
    return;
  },
  /**
   * onExludeClick create a filter from the highlights
   */
  onExcludeClick(): void {
    const self = (this as unknown) as DataExplorerRef;
    // check is highlights exist
    if (!self.isFilteringHighlights) {
      return;
    }
    const filter = createFiltersFromHighlights(self.highlights, EXCLUDE_FILTER);
    // add new filter to route
    addFilterToRoute(self.$router, filter);
    self.resetHighlightsOrRow();
    if (self.target) {
      datasetActions.fetchVariableRankings(store, {
        dataset: self.dataset,
        target: self.target.key,
      });
    }

    appActions.logUserEvent(store, {
      feature: Feature.FILTER_DATA,
      activity: Activity.DATA_PREPARATION,
      subActivity: SubActivity.DATA_TRANSFORMATION,
      details: { filter: filter },
    });
    return;
  },
};

export const SELECT_EVENT_HANDLERS = {
  /**
   * This function handles the apply cluster event
   * It updates the variable with a cluster column
   * then updates the variable summary with the new cluster datamode
   */
  [EventList.VARIABLES.APPLY_CLUSTER_EVENT]: function () {
    const self = (this as unknown) as DataExplorerRef;
    // fetch the var modes map
    const varModesMap = routeGetters.getDecodedVarModes(store);
    const clusterVars = new Set<string>();
    // find any grouped vars that are using this cluster data and update their
    // mode to cluster now that data is available
    datasetGetters
      .getGroupings(store)
      .filter((v) => isClusterType(v.colType))
      .forEach((v) => {
        varModesMap.set(v.key, SummaryMode.Cluster);
        clusterVars.add(v.grouping.clusterCol);
      });

    // find any image variables using this cluster data and update their mode
    datasetGetters
      .getVariables(store)
      .filter((v) => v.colType === IMAGE_TYPE)
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
    });
    self.$router.push(entry).catch((err) => console.warn(err));

    // update variables
    // pull the updated dataset, vars, and summaries
    const filterParams = routeGetters.getDecodedSolutionRequestFilterParams(
      store
    );
    const highlights = routeGetters.getDecodedHighlights(store);
    for (const [k, v] of varModesMap) {
      datasetActions.fetchVariableSummary(store, {
        dataset: self.dataset,
        variable: k,
        highlights: highlights,
        filterParams: filterParams,
        include: true,
        dataMode: DataMode.Cluster,
        mode: $enum(SummaryMode).asValueOrDefault(v, SummaryMode.Default),
        handleMutation: true,
      });
    }

    return;
  },
  /**
   * This handles outlier events for the select state
   * All it does is apply the outlier to the ds
   * then update the variables / variable summaries
   * **/
  [EventList.VARIABLES.APPLY_OUTLIER_EVENT]: async function () {
    const self = (this as unknown) as DataExplorerRef;
    const dataset = self.dataset;
    const success = await datasetActions.applyOutliers(store, dataset);
    if (!success) return;

    // Update the variables, which should now include the outlier variable.
    await datasetActions.fetchVariables(store, {
      dataset,
    });
    await viewActions.updateVariableSummaries(store);

    // Update the route to know that the outlier has been applied.
    const entry = overlayRouteEntry(self.$route, { outlier: "1" });
    self.$router.push(entry).catch((err) => console.warn(err));
    return;
  },
};
