import { datasetGetters } from "../store";
import {
  BaseState,
  PredictViewState,
  ResultViewState,
  SelectViewState,
} from "./state/AppStateWrapper";
import DataExplorer from "../views/DataExplorer.vue";
import store from "../store/store";

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
export const SELECT_VIEW_COMPUTED = {};
export const SELECT_VIEW_METHODS = {};
export const LABEL_VIEW_COMPUTED = {
  isClone(datasetName: string): boolean | null {
    const datasets = datasetGetters.getDatasets(store);
    const dataset = datasets.find((d) => d.id === datasetName);
    if (!dataset) {
      return null;
    }
    return dataset.clone === undefined ? false : dataset.clone;
  },
};
