import { CreateSolutionsFormRef, DataExplorerRef } from "../../componentTypes";
import { isEmpty, isNil } from "lodash";
import { appActions, datasetActions, requestActions } from "../../../store";
import { getters as routeGetters } from "../../../store/route/module";
import store from "../../../store/store";
import { SolutionRequestMsg } from "../../../store/requests/actions";
import { Solution } from "../../../store/requests";
import { DataMode } from "../../../store/dataset";
import { varModesToString } from "../../routes";
import { ExplorerStateNames } from "..";
import { createFiltersFromHighlights } from "../../highlights";
import {
  addFilterToRoute,
  EXCLUDE_FILTER,
  INCLUDE_FILTER,
} from "../../filters";
import { Activity, Feature, SubActivity } from "../../userEvents";
/**
 * SELECT_COMPUTES contains all of the computes for the select state in the data explorer
 **/
export const SELECT_COMPUTES = {
  // checks to see if the UI surrounding include/exclude table views is enabled
  // this should only be true for the select state
  isExcludeDisabled: (): boolean => {
    const self = this as DataExplorerRef;
    if (!self) {
      return false;
    }
    return !self.isFilteringHighlights && !self.isFilteringSelection;
  },
  // returns the variable keys that are currently selected as training features does not include target
  training: (): string[] => {
    return routeGetters.getDecodedTrainingVariableNames(store);
  },
  // returns true if a target variable is selected and at least 1 training feature has been selected
  // this enables the create model button
  isCreateModelPossible: (): boolean => {
    const self = this as DataExplorerRef;
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
  //onModelCreation starts the process for fitting a model
  onModelCreation: (msg: SolutionRequestMsg): void => {
    const self = this as DataExplorerRef;
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
  // onExludeClick create a filter from the highlights
  onExcludeClick: (): void => {
    const self = this as DataExplorerRef;
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
  // TODO: this needs to be removed
  onReincludeClick: (): void => {
    const self = this as DataExplorerRef;
    if (!self.isFilteringHighlights) {
      return;
    }
    const filter = createFiltersFromHighlights(self.highlights, INCLUDE_FILTER);

    addFilterToRoute(self.$router, filter);
    self.resetHighlightsOrRow();

    datasetActions.fetchVariableRankings(store, {
      dataset: self.dataset,
      target: self.target.key,
    });

    appActions.logUserEvent(store, {
      feature: Feature.UNFILTER_DATA,
      activity: Activity.DATA_PREPARATION,
      subActivity: SubActivity.DATA_TRANSFORMATION,
      details: { filter: filter },
    });
    return;
  },
};
