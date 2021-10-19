import { DataExplorerRef, SaveModalRef } from "../../componentTypes";
import {
  appActions,
  datasetActions,
  datasetGetters,
  modelActions,
  requestGetters,
  resultGetters,
  viewActions,
} from "../../../store";
import { getters as routeGetters } from "../../../store/route/module";
import store from "../../../store/store";
import { Solution } from "../../../store/requests";
import { ExplorerStateNames } from "..";
import { Activity, Feature, SubActivity } from "../../userEvents";
import {
  DataMode,
  Extrema,
  SummaryMode,
  TaskTypes,
} from "../../../store/dataset";
import { overlayRouteEntry, RouteArgs, varModesToString } from "../../routes";
import { isFittedSolutionIdSavedAsModel } from "../../models";
import { EI, EventList } from "../../events";
import { IMAGE_TYPE, isClusterType } from "../../types";

/**
 * RESULT_COMPUTES contains all of the computes for the result state in the data explorer
 **/
export const RESULT_COMPUTES = {
  /**
   * showResiduals determines if residuals should be displayed. Should only show for regression or forecasting
   */
  showResiduals(): boolean {
    const tasks = routeGetters.getRouteTask(store).split(",");
    return (
      tasks &&
      !!tasks.find(
        (t) => t === TaskTypes.REGRESSION || t === TaskTypes.FORECASTING
      )
    );
  },
  /**
   * solution the current active solution in the result state
   */
  solution(): Solution | null {
    return requestGetters.getActiveSolution(store);
  },
  /**
   *  solutionId the current action solution id
   */
  solutionId(): string | undefined {
    const self = (this as unknown) as DataExplorerRef;
    return self.solution?.solutionId;
  },
  fittedSolutionId(): string | undefined {
    const self = (this as unknown) as DataExplorerRef;
    return self.solution?.fittedSolutionId;
  },
  /**
   * isActiveSolutionSaved determines if the active solution model is saved
   * once saved the model can be used for predictions
   */
  isActiveSolutionSaved(): boolean | undefined {
    const self = (this as unknown) as DataExplorerRef;
    return self.isFittedSolutionIdSavedAsModel(self.fittedSolutionId);
  },
  /**
   * hasWeight checks to see if the table data contains shap values
   * the weights create the blue gradients in the result table cells
   */
  hasWeight(): boolean {
    return resultGetters.hasResultTableDataItemsWeight(store);
  },
  /**
   * residual extrema returns the extrema of the residuals used in the residual scrubber
   */
  residualExtrema(): Extrema {
    return resultGetters.getResidualsExtrema(store);
  },
  /**
   * isSingleSolution turns off the model toggle capabilities in ResultFacets
   * when there is a large amount of solutions it increases queries
   * the result facet will toggle off older solutions
   */
  isSingleSolution(): boolean {
    return routeGetters.isSingleSolution(store);
  },
};

export const RESULT_METHODS = {
  /**
   * onApplyModel is called when a model has been saved
   * and is being applied to a prediction dataset
   * this will transition the data explorer into the prediction state
   */
  async onApplyModel(args: RouteArgs): Promise<void> {
    const self = (this as unknown) as DataExplorerRef;
    const modal = (self.$refs.saveModel as unknown) as SaveModalRef;
    self.updateRoute(args);
    modal.hideSaveForm();
    await self.changeStatesByName(ExplorerStateNames.PREDICTION_VIEW);
  },
  /**
   * isFittedSolutionIdSavedAsModel checks if model has been saved and can be used for predictions
   */
  isFittedSolutionIdSavedAsModel,
  /**
   * fetchSummarSolution gets the information from the back end about the model tests during fitting
   * the summaries are used within the ResultFacets
   */
  async fetchSummarySolution(id: string): Promise<void> {
    viewActions.updateResultSummaries(store, { requestIds: [id] });
  },
  /**
   * onSaveModel sends information to the backend and the current active solution's model
   * will be saved and therefore will be able to be used for predictions
   */
  async onSaveModel(args: EI.RESULT.SaveInfo): Promise<void> {
    const self = (this as unknown) as DataExplorerRef;
    appActions.logUserEvent(store, {
      feature: Feature.EXPORT_MODEL,
      activity: Activity.MODEL_SELECTION,
      subActivity: SubActivity.MODEL_SAVE,
      details: {
        solution: args.solutionId,
        fittedSolution: args.fittedSolution,
      },
    });
    const modal = (self.$refs.saveModel as unknown) as SaveModalRef;
    try {
      const err = await appActions.saveModel(store, {
        fittedSolutionId: self.fittedSolutionId,
        modelName: args.name,
        modelDescription: args.description,
      });
      // should probably change UI based on error
      if (!err) {
        await modelActions.fetchModels(store);
        modal.isSaving = false;
        modal.hideSaveForm();
      }
    } catch (err) {
      modal.isSaving = false;
      console.warn(err);
    }
    return;
  },
};

export const RESULT_EVENT_HANDLERS = {
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
    // fetch the new summaries with the clustering applied
    viewActions.updateResultsSummaries(store);
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
    await viewActions.updateResultsSummaries(store);

    // Update the route to know that the outlier has been applied.
    const entry = overlayRouteEntry(self.$route, { outlier: "1" });
    self.$router.push(entry).catch((err) => console.warn(err));
    return;
  },
};
