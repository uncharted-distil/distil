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
  BaseState,
  LabelViewState,
  PredictViewState,
  ResultViewState,
  SelectViewState,
} from "../state/AppStateWrapper";
import { Variable } from "../../store/dataset";
import { DataExplorerRef } from "../componentTypes";
import { META_TYPES } from "../types";
import { GENERIC_METHODS } from "./functions/generic";
import { LABEL_METHODS, LABEL_COMPUTES } from "./functions/label";
import { RESULT_METHODS, RESULT_COMPUTES } from "./functions/result";
import { SELECT_METHODS, SELECT_COMPUTES } from "./functions/select";
import {
  PREDICTION_COMPUTES,
  PREDICTION_METHODS,
} from "./functions/prediction";
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

export const genericMethods = GENERIC_METHODS;
export const labelMethods = LABEL_METHODS;
export const labelComputes = LABEL_COMPUTES;
export const resultMethods = RESULT_METHODS;
export const resultComputes = RESULT_COMPUTES;
export const selectComputes = SELECT_COMPUTES;
export const selectMethods = SELECT_METHODS;
export const predictionMethods = PREDICTION_METHODS;
export const predictionComputes = PREDICTION_COMPUTES;
