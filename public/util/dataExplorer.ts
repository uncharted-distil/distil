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
  datasetMutations,
  requestActions,
  requestGetters,
  resultGetters,
} from "../store";
import {
  BaseState,
  LabelViewState,
  PredictViewState,
  ResultViewState,
  SelectViewState,
} from "./state/AppStateWrapper";
import store from "../store/store";
import {
  D3M_INDEX_FIELD,
  DataMode,
  TableRow,
  Variable,
  VariableSummary,
} from "../store/dataset";
import { getters as routeGetters } from "../store/route/module";
import { isEmpty, isNil } from "lodash";
import { Solution } from "../store/requests";
import { isFittedSolutionIdSavedAsModel } from "./models";
import { SolutionRequestMsg } from "../store/requests/actions";
import { RouteArgs, varModesToString } from "./routes";
import {
  ActionColumnRef,
  CreateSolutionsFormRef,
  SaveModalRef,
  DataExplorerRef,
  DataView,
} from "./componentTypes";
import {
  addFilterToRoute,
  emptyFilterParamsObject,
  EXCLUDE_FILTER,
  INCLUDE_FILTER,
} from "./filters";
import { cloneFilters, createFiltersFromHighlights } from "./highlights";
import {
  bulkRowSelectionUpdate,
  clearRowSelection,
  createFilterFromRowSelection,
} from "./row";
import { Activity, Feature, SubActivity } from "./userEvents";
import { EI } from "./events";
import { CATEGORICAL_TYPE, META_TYPES } from "./types";
import {
  addOrderBy,
  cloneDatasetUpdateRoute,
  downloadFile,
  LowShotLabels,
  LOW_SHOT_RANK_COLUMN_PREFIX,
  LOW_SHOT_SCORE_COLUMN_PREFIX,
} from "./data";
import { LABEL_FEATURE_INSTANCE } from "../store/route";
import router from "../router/router";

export interface Action {
  name: string;
  icon: string;
  paneId: string;
  count?: number;
  toggle?: boolean;
  variables: (self: DataExplorerRef) => Variable[];
}

export default interface ExplorerConfig {
  // required actions in current state
  actionList: Action[];
  // these actions will be toggled when the state is switched to
  defaultAction: ActionNames[];
  currentPane: string;
  resetConfig: (self: DataExplorerRef) => void;
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
    case ExplorerStateNames.LABEL_VIEW:
      return new LabelViewConfig();
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
    case ExplorerStateNames.LABEL_VIEW:
      return new LabelViewState();
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
  resetConfig(self: DataExplorerRef) {
    return;
  }
  currentPane = ACTION_MAP.get(ActionNames.ALL_VARIABLES).paneId;
  defaultAction = [];
}
export class ResultViewConfig implements ExplorerConfig {
  get actionList(): Action[] {
    const actions = [
      ActionNames.MODEL_VARIABLES,
      ActionNames.TEXT_VARIABLES,
      ActionNames.CATEGORICAL_VARIABLES,
      ActionNames.NUMBER_VARIABLES,
      ActionNames.LOCATION_VARIABLES,
      ActionNames.IMAGE_VARIABLES,
      ActionNames.UNKNOWN_VARIABLES,
      ActionNames.TARGET_VARIABLE,
      ActionNames.OUTCOME_VARIABLES,
    ];
    return actions.map((a) => {
      return ACTION_MAP.get(a);
    });
  }
  resetConfig(self: DataExplorerRef) {
    return;
  }
  currentPane = ACTION_MAP.get(ActionNames.MODEL_VARIABLES).paneId;
  defaultAction = [ActionNames.OUTCOME_VARIABLES];
}
export class PredictViewConfig implements ExplorerConfig {
  get actionList(): Action[] {
    const actions = [
      ActionNames.MODEL_VARIABLES,
      ActionNames.TEXT_VARIABLES,
      ActionNames.CATEGORICAL_VARIABLES,
      ActionNames.NUMBER_VARIABLES,
      ActionNames.LOCATION_VARIABLES,
      ActionNames.IMAGE_VARIABLES,
      ActionNames.UNKNOWN_VARIABLES,
      ActionNames.TARGET_VARIABLE,
      ActionNames.OUTCOME_VARIABLES,
    ];
    return actions.map((a) => {
      return ACTION_MAP.get(a);
    });
  }
  resetConfig(self: DataExplorerRef) {
    return;
  }
  currentPane = ACTION_MAP.get(ActionNames.MODEL_VARIABLES).paneId;
  defaultAction = [ActionNames.OUTCOME_VARIABLES];
}
export class LabelViewConfig implements ExplorerConfig {
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
  resetConfig(self: DataExplorerRef) {
    self.labelName = "";
  }
  currentPane = ACTION_MAP.get(ActionNames.ALL_VARIABLES).paneId;
  defaultAction = [];
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
  MODEL_VARIABLES = "Model Variables",
}

export const ACTIONS = [
  {
    name: ActionNames.CREATE_NEW_VARIABLE,
    icon: "fa fa-plus",
    paneId: "add",
    variables: (self: DataExplorerRef) => {
      return [];
    },
  },
  {
    name: ActionNames.ALL_VARIABLES,
    icon: "fa fa-database",
    paneId: "available",
    variables: (self: DataExplorerRef) => {
      return self.variables;
    },
  },
  {
    name: ActionNames.TEXT_VARIABLES,
    icon: "fa fa-font",
    paneId: "text",
    variables: function (self: DataExplorerRef) {
      return self.variables.filter((v) => {
        return META_TYPES[this.paneId].includes(v.colType);
      });
    },
  },
  {
    name: ActionNames.CATEGORICAL_VARIABLES,
    icon: "fa fa-align-left",
    paneId: "categorical",
    variables: function (self: DataExplorerRef) {
      return self.variables.filter((v) => {
        return META_TYPES[this.paneId].includes(v.colType);
      });
    },
  },
  {
    name: ActionNames.NUMBER_VARIABLES,
    icon: "fa fa-bar-chart",
    paneId: "number",
    variables: function (self: DataExplorerRef) {
      return self.variables.filter((v) => {
        return META_TYPES[this.paneId].includes(v.colType);
      });
    },
  },
  {
    name: ActionNames.TIME_VARIABLES,
    icon: "fa fa-clock-o",
    paneId: "time",
    variables: function (self: DataExplorerRef) {
      return self.variables.filter((v) => {
        return META_TYPES[this.paneId].includes(v.colType);
      });
    },
  },
  {
    name: ActionNames.LOCATION_VARIABLES,
    icon: "fa fa-map-o",
    paneId: "location",
    variables: function (self: DataExplorerRef) {
      return self.variables.filter((v) => {
        return META_TYPES[this.paneId].includes(v.colType);
      });
    },
  },
  {
    name: ActionNames.IMAGE_VARIABLES,
    icon: "fa fa-image",
    paneId: "image",
    variables: function (self: DataExplorerRef) {
      return self.variables.filter((v) => {
        return META_TYPES[this.paneId].includes(v.colType);
      });
    },
  },
  {
    name: ActionNames.UNKNOWN_VARIABLES,
    icon: "fa fa-question",
    paneId: "unknown",
    variables: function (self: DataExplorerRef) {
      return self.variables.filter((v) => {
        return META_TYPES[this.paneId].includes(v.colType);
      });
    },
  },
  {
    name: ActionNames.TARGET_VARIABLE,
    icon: "fa fa-crosshairs",
    paneId: "target",
    variables: function (self: DataExplorerRef) {
      return self.target ? [self.target] : [];
    },
  },
  {
    name: ActionNames.TRAINING_VARIABLE,
    icon: "fa fa-asterisk",
    paneId: "training",
    variables: function (self: DataExplorerRef) {
      return self.variables.filter((variable) =>
        self.training.includes(variable.key)
      );
    },
  },
  {
    name: ActionNames.MODEL_VARIABLES,
    icon: "fa fa-database",
    paneId: "model",
    variables: function (self: DataExplorerRef) {
      return self.variables;
    },
  },
  {
    name: ActionNames.OUTCOME_VARIABLES,
    icon: "fas fa-poll",
    paneId: "outcome",
    toggle: false,
    variables: function (self: DataExplorerRef) {
      return self.state.getSecondaryVariables();
    },
  },
] as Action[];

export const ACTION_MAP = new Map(
  ACTIONS.map((a) => {
    return [a.name, a];
  })
);
/**************MIXINS********************/
/*This next portion of the file is dedicated to grouping computes/methods into state objects
 and will be used in the explorer component.
The goal here is to move all of the code for the component out of the component file. 
If this file gets too large we can move each state into their own folder. */
export const GENERIC_METHODS = {
  toggleAction: (
    self: DataExplorerRef
  ): ((actionName: ActionNames) => void) => {
    return (actionName: ActionNames) => {
      const actionColumn = (self.$refs[
        "action-column"
      ] as unknown) as ActionColumnRef;
      actionColumn.toggle(ACTION_MAP.get(actionName).paneId);
    };
  },
};
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
      solutionRequestMsg.filters.variables = routeGetters
        .getRouteTrainingVariables(store)
        .split(",")
        .concat(routeGetters.getRouteTargetVariable(store));
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
          modelCreationRef.success();
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
    };
  },
  onExcludeClick: (self: DataExplorerRef): (() => void) => {
    return () => {
      let filter = null;
      if (self.isFilteringHighlights) {
        filter = createFiltersFromHighlights(self.highlights, EXCLUDE_FILTER);
      } else {
        filter = createFilterFromRowSelection(
          self.rowSelection,
          EXCLUDE_FILTER
        );
      }

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
    };
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
  labelSummary: (self: DataExplorerRef): VariableSummary => {
    const label = routeGetters.getRouteLabel(store);
    return self.summaries.find((s) => {
      return s.key === label;
    });
  },
};

export const LABEL_METHODS = {
  updateTask: (self: DataExplorerRef): (() => Promise<void>) => {
    return async () => {
      const taskResponse = await datasetActions.fetchTask(store, {
        dataset: self.dataset,
        targetName: self.labelName,
        variableNames: self.variables.map((v) => v.key),
      });
      const training = routeGetters.getDecodedTrainingVariableNames(store);
      const check = training.length;
      const trainingMap = new Map(
        training.map((t) => {
          return [t, true];
        })
      );
      self.variables.forEach((variable) => {
        if (!trainingMap.has(variable.key)) {
          training.push(variable.key);
        }
      });
      if (check === training.length) {
        return;
      }
      self.updateRoute({
        task: taskResponse.data.task.join(","),
        training: training.join(","),
        label: self.labelName,
      });
      return;
    };
  },
  onLabelSubmit: (self: DataExplorerRef): (() => Promise<void>) => {
    return async () => {
      if (
        self.variables.some((v) => {
          return v.colName === self.labelName;
        })
      ) {
        self.updateRoute({
          label: self.labelName,
        });
        await self.changeStatesByName(ExplorerStateNames.LABEL_VIEW);
        return;
      }
      const entry = await cloneDatasetUpdateRoute();
      // failed to clone
      if (entry === null) {
        return;
      }
      self.$router.push(entry).catch((err) => console.warn(err));
      // add new field
      await datasetActions.addField<string>(store, {
        dataset: self.dataset,
        name: self.labelName,
        fieldType: CATEGORICAL_TYPE,
        defaultValue: LowShotLabels.unlabeled,
        displayName: self.labelName,
      });
      // fetch new dataset with the newly added field
      await self.changeStatesByName(ExplorerStateNames.LABEL_VIEW);
      // update task based on the current training data
      self.updateTask();
    };
  },
  onAnnotationChanged: (
    self: DataExplorerRef
  ): ((label: LowShotLabels) => Promise<void>) => {
    return async (label: LowShotLabels) => {
      const rowSelection = routeGetters.getDecodedRowSelection(store);
      const innerData = new Map<number, unknown>();
      const updateData = rowSelection.d3mIndices.map((i) => {
        innerData.set(i, { LowShotLabel: label });
        return {
          index: i.toString(),
          name: self.labelName,
          value: label,
        };
      });
      if (!updateData.length) {
        return;
      }
      const dataset = routeGetters.getRouteDataset(store);
      datasetMutations.updateAreaOfInterestIncludeInner(store, innerData);
      datasetActions.updateDataset(store, {
        dataset: dataset,
        updateData,
      });
      clearRowSelection(self.$router);
      self.updateRoute({
        annotationHasChanged: true,
      });
      await self.state.fetchData();
      if (self.isRemoteSensing) {
        self.state.fetchMapBaseline();
      }
      return;
    };
  },
  onExport: (self: DataExplorerRef): (() => Promise<void>) => {
    return async () => {
      const highlights = [
        {
          context: LABEL_FEATURE_INSTANCE,
          dataset: self.dataset,
          key: self.labelName,
          value: LowShotLabels.unlabeled,
        },
      ]; // exclude unlabeled from data export
      const filterParams = routeGetters.getDecodedSolutionRequestFilterParams(
        store
      );
      const dataMode = routeGetters.getDataMode(store);
      const file = await datasetActions.extractDataset(store, {
        dataset: self.dataset,
        filterParams,
        highlights,
        include: true,
        mode: EXCLUDE_FILTER,
        dataMode,
      });
      downloadFile(file, self.dataset, ".csv");
      return;
    };
  },
  onSearchSimilar: (self: DataExplorerRef): (() => Promise<void>) => {
    return async () => {
      self.isBusy = true;
      const res = (await requestActions.createQueryRequest(store, {
        datasetId: self.dataset,
        target: self.labelName,
        filters: emptyFilterParamsObject(),
      })) as { success: boolean; error: string };
      if (!res.success) {
        self.$bvToast.toast(res.error, {
          title: "Error",
          autoHideDelay: 5000,
          appendToast: true,
          variant: "danger",
          toaster: "b-toaster-bottom-right",
        });
      }
      const labelScoreName = LOW_SHOT_SCORE_COLUMN_PREFIX + self.labelName;
      addOrderBy(labelScoreName);
      self.isBusy = false;
      await self.state.fetchData();
      self.state.fetchMapBaseline();
      self.updateRoute({
        annotationHasChanged: false,
      });
      const outcome = ACTION_MAP.get(ActionNames.OUTCOME_VARIABLES);
      const open = routeGetters.getToggledActions(store).some((a) => {
        return a === outcome.paneId;
      });
      // open the outcome variable pane to display the new confidence and ranking
      if (!open) {
        self.toggleAction(ActionNames.OUTCOME_VARIABLES);
      }
    };
  },
  onSaveDataset: (
    self: DataExplorerRef
  ): ((saveName: string, retainUnlabeled: boolean) => Promise<void>) => {
    return async (saveName: string, retainUnlabeled: boolean) => {
      self.isBusy = true;
      const labelScoreName = LOW_SHOT_SCORE_COLUMN_PREFIX + self.labelName;
      const labelRankName = LOW_SHOT_RANK_COLUMN_PREFIX + self.labelName;
      const highlightsClear = [
        {
          context: LABEL_FEATURE_INSTANCE,
          dataset: self.dataset,
          key: self.labelName,
          value: LowShotLabels.unlabeled,
        },
      ]; // exclude unlabeled from data export
      const highlights = retainUnlabeled ? null : highlightsClear;
      let filterParams = routeGetters.getDecodedSolutionRequestFilterParams(
        store
      );
      filterParams = cloneFilters(filterParams);
      if (
        self.allVariables.some((v) => {
          return v.key === labelScoreName;
        })
      ) {
        // delete confidence variable when saving
        await datasetActions.deleteVariable(store, {
          dataset: self.dataset,
          key: labelScoreName,
        });
        await datasetActions.deleteVariable(store, {
          dataset: self.dataset,
          key: labelRankName,
        });
      }
      // clear the unlabeled values when saving
      if (retainUnlabeled) {
        await datasetActions.clearVariable(store, {
          dataset: self.dataset,
          key: self.labelName,
          highlights: highlightsClear,
          filterParams: filterParams,
        });
      }
      const dataMode = routeGetters.getDataMode(store);
      await datasetActions.saveDataset(store, {
        dataset: self.dataset,
        datasetNewName: saveName,
        filterParams,
        highlights,
        include: false,
        mode: INCLUDE_FILTER,
        dataMode,
      });
      self.isBusy = false;
      // CHANGE TO SELECT VIEW AFTER DS IS SAVED IN LABEL VIEW
      self.changeStatesByName(ExplorerStateNames.SELECT_VIEW);
      return;
    };
  },
  confidenceGetter: (
    self: DataExplorerRef
  ): ((item: TableRow, idx: number) => number) => {
    return (item: TableRow, idx: number) => {
      if (item[self.labelName].value === LowShotLabels.positive) {
        return 1.0;
      }
      if (item[self.labelName].value === LowShotLabels.negative) {
        return 0;
      }
      const labelScoreName = LOW_SHOT_SCORE_COLUMN_PREFIX + self.labelName;
      // comes back order by confidence so the rank is already engrained in the array
      if (item[labelScoreName]) {
        return 1.0 - idx / self.items.length;
      }
      return undefined;
    };
  },
  onSelectAll: (self: DataExplorerRef): (() => void) => {
    return () => {
      const dataView = (self.$refs.dataView as unknown) as DataView;
      dataView.selectAll();
    };
  },
  onToolSelection: (
    self: DataExplorerRef
  ): ((selection: EI.MAP.SelectionHighlight) => Promise<void>) => {
    return async (selection: EI.MAP.SelectionHighlight) => {
      const filterParams = routeGetters.getDecodedSolutionRequestFilterParams(
        store
      );
      filterParams.size = datasetGetters.getIncludedTableDataNumRows(store);
      // fetch data selected by map tool
      const resp = await datasetActions.fetchTableData(store, {
        dataset: selection.dataset,
        highlights: [selection],
        filterParams: filterParams,
        dataMode: null,
        include: true,
      });
      // find d3mIndex
      const labelIndex = resp.columns.findIndex((c) => {
        return c.key === D3M_INDEX_FIELD;
      });
      // if -1 then something failed
      if (labelIndex === -1) {
        return;
      }
      // map the values
      const indices = resp.values.map((v) => {
        return v[labelIndex].value.toString();
      });
      // update row selection
      const rowSelection = routeGetters.getDecodedRowSelection(store);
      bulkRowSelectionUpdate(router, selection.context, rowSelection, indices);
    };
  },
};
