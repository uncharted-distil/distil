/**
 *
 *    Copyright © 2021 Uncharted Software Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */
import {
  appActions,
  datasetActions,
  datasetGetters,
  requestActions,
  requestGetters,
  resultGetters,
} from "../store";
import {
  BaseState,
  PredictViewState,
  ResultViewState,
  SelectViewState,
} from "./state/AppStateWrapper";
import store from "../store/store";
import { DataMode } from "../store/dataset";
import { getters as routeGetters } from "../store/route/module";
import { isEmpty, isNil } from "lodash";
import { Solution } from "../store/requests";
import { isFittedSolutionIdSavedAsModel } from "./models";
import { SolutionRequestMsg } from "../store/requests/actions";
import { overlayRouteEntry, RouteArgs, varModesToString } from "./routes";
import {
  ActionColumnRef,
  CreateSolutionsFormRef,
  SaveModalRef,
  DataExplorerRef,
} from "./componentTypes";
import { addFilterToRoute, EXCLUDE_FILTER, INCLUDE_FILTER } from "./filters";
import { createFiltersFromHighlights } from "./highlights";
import { createFilterFromRowSelection } from "./row";
import { Activity, Feature, SubActivity } from "./userEvents";
import { EI } from "./events";
import { CATEGORICAL_TYPE } from "./types";
import { LowShotLabels } from "./data";

export interface Action {
  name: string;
  icon: string;
  paneId: string;
  count?: number;
  toggle?: boolean;
}

export default interface ExplorerConfig {
  // required actions in current state
  actionList: Action[];
}

// DataExplorer possible state, used in route
export enum ExplorerStateNames {
  SELECT_VIEW = "select",
  RESULT_VIEW = "result",
  PREDICTION_VIEW = "prediction",
  LABEL_VIEW = "label",
}
// getConfigFromName returns a config instance based on supplied enum, throws errors
export function getConfigFromName(state: ExplorerStateNames): ExplorerConfig {
  switch (state) {
    case ExplorerStateNames.SELECT_VIEW:
      return new SelectViewConfig();
    case ExplorerStateNames.RESULT_VIEW:
      return new ResultViewConfig();
    case ExplorerStateNames.PREDICTION_VIEW:
      return new PredictViewConfig();
    default:
      throw Error("Config State not supported");
  }
}
// getStateFromName returns a State instance based on supplied enum, throws errors
export function getStateFromName(state: ExplorerStateNames): BaseState {
  switch (state) {
    case ExplorerStateNames.SELECT_VIEW:
      return new SelectViewState();
    case ExplorerStateNames.RESULT_VIEW:
      return new ResultViewState();
    case ExplorerStateNames.PREDICTION_VIEW:
      return new PredictViewState();
    default:
      throw Error("Config State not supported");
  }
}

export class SelectViewConfig implements ExplorerConfig {
  get actionList(): Action[] {
    const actions = [
      ActionNames.CREATE_NEW_VARIABLE,
      ActionNames.ALL_VARIABLES,
      ActionNames.TEXT_VARIABLES,
      ActionNames.CATEGORICAL_VARIABLES,
      ActionNames.NUMBER_VARIABLES,
      ActionNames.LOCATION_VARIABLES,
      ActionNames.IMAGE_VARIABLES,
      ActionNames.UNKNOWN_VARIABLES,
      ActionNames.TARGET_VARIABLE,
      ActionNames.TRAINING_VARIABLE,
    ];
    return actions.map((a) => {
      return ACTION_MAP.get(a);
    });
  }
}
export class ResultViewConfig implements ExplorerConfig {
  get actionList(): Action[] {
    const actions = [
      ActionNames.ALL_VARIABLES,
      ActionNames.TEXT_VARIABLES,
      ActionNames.CATEGORICAL_VARIABLES,
      ActionNames.NUMBER_VARIABLES,
      ActionNames.LOCATION_VARIABLES,
      ActionNames.IMAGE_VARIABLES,
      ActionNames.UNKNOWN_VARIABLES,
      ActionNames.TARGET_VARIABLE,
      ActionNames.TRAINING_VARIABLE,
      ActionNames.OUTCOME_VARIABLES,
    ];
    return actions.map((a) => {
      return ACTION_MAP.get(a);
    });
  }
}
export class PredictViewConfig implements ExplorerConfig {
  get actionList(): Action[] {
    const actions = [
      ActionNames.ALL_VARIABLES,
      ActionNames.TEXT_VARIABLES,
      ActionNames.CATEGORICAL_VARIABLES,
      ActionNames.NUMBER_VARIABLES,
      ActionNames.LOCATION_VARIABLES,
      ActionNames.IMAGE_VARIABLES,
      ActionNames.UNKNOWN_VARIABLES,
      ActionNames.TARGET_VARIABLE,
      ActionNames.TRAINING_VARIABLE,
      ActionNames.OUTCOME_VARIABLES,
    ];
    return actions.map((a) => {
      return ACTION_MAP.get(a);
    });
  }
}
export enum ActionNames {
  CREATE_NEW_VARIABLE = "Create New Variable",
  ALL_VARIABLES = "All Variables",
  TEXT_VARIABLES = "Text Variables",
  CATEGORICAL_VARIABLES = "Categorical Variables",
  NUMBER_VARIABLES = "Number Variables",
  TIME_VARIABLES = "Time Variables",
  LOCATION_VARIABLES = "Location Variables",
  IMAGE_VARIABLES = "Image Variables",
  UNKNOWN_VARIABLES = "Unknown Variables",
  TARGET_VARIABLE = "Target Variable",
  TRAINING_VARIABLE = "Training Variables",
  OUTCOME_VARIABLES = "Outcome Variables",
}

export const ACTIONS = [
  { name: ActionNames.CREATE_NEW_VARIABLE, icon: "fa fa-plus", paneId: "add" },
  {
    name: ActionNames.ALL_VARIABLES,
    icon: "fa fa-database",
    paneId: "available",
  },
  { name: ActionNames.TEXT_VARIABLES, icon: "fa fa-font", paneId: "text" },
  {
    name: ActionNames.CATEGORICAL_VARIABLES,
    icon: "fa fa-align-left",
    paneId: "categorical",
  },
  {
    name: ActionNames.NUMBER_VARIABLES,
    icon: "fa fa-bar-chart",
    paneId: "number",
  },
  { name: ActionNames.TIME_VARIABLES, icon: "fa fa-clock-o", paneId: "time" },
  {
    name: ActionNames.LOCATION_VARIABLES,
    icon: "fa fa-map-o",
    paneId: "location",
  },
  { name: ActionNames.IMAGE_VARIABLES, icon: "fa fa-image", paneId: "image" },
  {
    name: ActionNames.UNKNOWN_VARIABLES,
    icon: "fa fa-question",
    paneId: "unknown",
  },
  {
    name: ActionNames.TARGET_VARIABLE,
    icon: "fa fa-crosshairs",
    paneId: "target",
  },
  {
    name: ActionNames.TRAINING_VARIABLE,
    icon: "fa fa-asterisk",
    paneId: "training",
  },
  {
    name: ActionNames.OUTCOME_VARIABLES,
    icon: "fas fa-poll",
    paneId: "outcome",
    toggle: false,
  },
] as Action[];

export const ACTION_MAP = new Map(
  ACTIONS.map((a) => {
    return [a.name, a];
  })
);
/**************MIXINS********************/
export const SELECT_COMPUTES = {
  isExcludedDisabled: (self: DataExplorerRef): boolean => {
    return !self.isFilteringHighlights && !self.isFilteringSelection;
  },
  training: (self: DataExplorerRef): string[] => {
    return routeGetters.getDecodedTrainingVariableNames(store);
  },
  isCreateModelPossible: (self: DataExplorerRef): boolean => {
    // check that we have some target and training variables.
    return !isNil(self.target) && !isEmpty(self.training);
  },
};
export const SELECT_METHODS = {
  onModelCreation: (
    self: DataExplorerRef
  ): ((msg: SolutionRequestMsg) => void) => {
    return (solutionRequestMsg: SolutionRequestMsg) => {
      requestActions
        .createSolutionRequest(store, solutionRequestMsg)
        .then(async (res: Solution) => {
          const dataMode = routeGetters.getDataMode(store);
          const dataModeDefault = dataMode ? dataMode : DataMode.Default;
          // transition to result screen
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
          modelCreationRef.pending = false;
          await self.changeStatesByName(ExplorerStateNames.RESULT_VIEW);
          const actionColumn = (self.$refs[
            "action-column"
          ] as unknown) as ActionColumnRef;
          actionColumn.toggle(
            ACTION_MAP.get(ActionNames.OUTCOME_VARIABLES).paneId
          );
        })
        .catch((err) => {
          console.error(err);
        });
      return;
    };
  },
  onExcludeClick: (self: DataExplorerRef): (() => void) => {
    let filter = null;
    if (self.isFilteringHighlights) {
      filter = createFiltersFromHighlights(self.highlights, EXCLUDE_FILTER);
    } else {
      filter = createFilterFromRowSelection(self.rowSelection, EXCLUDE_FILTER);
    }

    addFilterToRoute(self.$router, filter);
    self.resetHighlightsOrRow();

    datasetActions.fetchVariableRankings(store, {
      dataset: self.dataset,
      target: self.target.key,
    });

    appActions.logUserEvent(store, {
      feature: Feature.FILTER_DATA,
      activity: Activity.DATA_PREPARATION,
      subActivity: SubActivity.DATA_TRANSFORMATION,
      details: { filter: filter },
    });
    return;
  },
  onReincludeClick: (self: DataExplorerRef): (() => void) => {
    let filter = null;
    if (self.isFilteringHighlights) {
      filter = createFiltersFromHighlights(self.highlights, INCLUDE_FILTER);
    } else {
      filter = createFilterFromRowSelection(self.rowSelection, INCLUDE_FILTER);
    }

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
export const RESULT_COMPUTES = {
  solution: (self: DataExplorerRef): Solution => {
    return requestGetters.getActiveSolution(store);
  },
  solutionId: (self: DataExplorerRef): string => {
    return self.solution?.solutionId;
  },
  fittedSolutionId: (self: DataExplorerRef): string => {
    return self.solution?.fittedSolutionId;
  },
  isActiveSolutionSaved: (self: DataExplorerRef): boolean => {
    return self.isFittedSolutionIdSavedAsModel(self.fittedSolutionId);
  },
  hasWeight: (self: DataExplorerRef): boolean => {
    return resultGetters.hasResultTableDataItemsWeight(store);
  },
};
export const RESULT_METHODS = {
  onApplyModel: (
    self: DataExplorerRef
  ): ((args: RouteArgs) => Promise<void>) => {
    return async (args: RouteArgs) => {
      self.updateRoute(args);
      await self.changeStatesByName(ExplorerStateNames.PREDICTION_VIEW);
    };
  },
  isFittedSolutionIdSavedAsModel: (
    self: DataExplorerRef
  ): ((id: string) => boolean) => {
    return isFittedSolutionIdSavedAsModel;
  },
  onSaveModel: (
    self: DataExplorerRef
  ): ((args: EI.RESULT.SaveInfo) => Promise<void>) => {
    return async (args: EI.RESULT.SaveInfo) => {
      appActions.logUserEvent(store, {
        feature: Feature.EXPORT_MODEL,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_SAVE,
        details: {
          solution: args.solutionId,
          fittedSolution: args.fittedSolution,
        },
      });

      try {
        const err = await appActions.saveModel(store, {
          fittedSolutionId: self.fittedSolutionId,
          modelName: args.name,
          modelDescription: args.description,
        });
        // should probably change UI based on error
        if (!err) {
          const modal = (self.$refs.saveModel as unknown) as SaveModalRef;
          modal.showSuccessModel();
        }
      } catch (err) {
        console.warn(err);
      }
      return;
    };
  },
};
// label view computes
export const LABEL_COMPUTES = {
  isClone: (self: DataExplorerRef): boolean | null => {
    const datasets = datasetGetters.getDatasets(store);
    const dataset = datasets.find((d) => d.id === self.dataset);
    if (!dataset) {
      return null;
    }
    return dataset.clone === undefined ? false : dataset.clone;
  },
  options: (self: DataExplorerRef): { value: string; text: string }[] => {
    return self.variables
      .filter((v) => {
        return v.colType === CATEGORICAL_TYPE;
      })
      .map((v) => {
        return { value: v.colName, text: v.colName };
      });
  },
  labelModalTitle: (self: DataExplorerRef): string => {
    return self.isClone ? "Select Label Feature" : "Label Creation";
  },
};

export const LABEL_METHODS = {
  onLabelSubmit: (self: DataExplorerRef): (() => Promise<void>) => {
    return async () => {
      if (
        self.variables.some((v) => {
          return v.colName === self.labelName;
        })
      ) {
        const entry = overlayRouteEntry(routeGetters.getRoute(store), {
          label: self.labelName,
        });

        self.$router.push(entry).catch((err) => console.warn(err));
        return;
      }
      // add new field
      await datasetActions.addField<string>(store, {
        dataset: self.dataset,
        name: self.labelName,
        fieldType: CATEGORICAL_TYPE,
        defaultValue: LowShotLabels.unlabeled,
        displayName: self.labelName,
      });
      // fetch new dataset with the newly added field
      await self.state.fetchData();
      // update task based on the current training data
      //this.updateRoute();
    };
  },
};
