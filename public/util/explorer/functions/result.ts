import { DataExplorerRef, SaveModalRef } from "../../componentTypes";
import {
  appActions,
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
import { Extrema, TaskTypes } from "../../../store/dataset";
import { RouteArgs } from "../../routes";
import { isFittedSolutionIdSavedAsModel } from "../../models";
import { EI } from "../../events";
import { ModalPlugin } from "bootstrap-vue";

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
