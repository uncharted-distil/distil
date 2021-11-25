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
import { DataMode, Variable } from "../../store/dataset";
import { DataExplorerRef } from "../componentTypes";
import { DISTIL_ROLES, META_TYPES } from "../types";
import { GENERIC_COMPUTES, GENERIC_METHODS } from "./functions/generic";
import {
  LABEL_METHODS,
  LABEL_COMPUTES,
  LABEL_EVENT_HANDLERS,
} from "./functions/label";
import {
  RESULT_METHODS,
  RESULT_COMPUTES,
  RESULT_EVENT_HANDLERS,
} from "./functions/result";
import {
  SELECT_METHODS,
  SELECT_COMPUTES,
  SELECT_EVENT_HANDLERS,
} from "./functions/select";
import {
  PREDICTION_COMPUTES,
  PREDICTION_EVENT_HANDLERS,
  PREDICTION_METHODS,
} from "./functions/prediction";
import { getters as routeGetters } from "../../store/route/module";
import store from "../../store/store";
import { Dictionary } from "vue-router/types/router";
import { hasRole } from "../data";
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
  eventHandlers: Dictionary<Function>;
}

// DataExplorer possible state, used in route
export enum ExplorerStateNames {
  SELECT_VIEW = "select",
  RESULT_VIEW = "result",
  PREDICTION_VIEW = "prediction",
  LABEL_VIEW = "label",
}
export enum ExplorerViewComponent {
  TABLE = 0,
  MAP = 1,
  IMAGE_MOSAIC = 2,
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
      ActionNames.COMPUTED_VARIABLES,
      ActionNames.TARGET_VARIABLE,
      ActionNames.TRAINING_VARIABLE,
      ActionNames.EXPORT,
    ];
    return actions.map((a) => {
      return ACTION_MAP.get(a);
    });
  }
  resetConfig(self: DataExplorerRef) {
    self.removeEventHandlers(this.eventHandlers);
  }
  currentPane = ACTION_MAP.get(ActionNames.ALL_VARIABLES).paneId;
  defaultAction = [];
  // These events handlers get binded when the state becomes active and will be removed when the state is left
  eventHandlers = SELECT_EVENT_HANDLERS;
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
      ActionNames.COMPUTED_VARIABLES,
      ActionNames.TARGET_VARIABLE,
      ActionNames.OUTCOME_VARIABLES,
      ActionNames.EXPORT,
    ];
    return actions.map((a) => {
      return ACTION_MAP.get(a);
    });
  }
  resetConfig(self: DataExplorerRef) {
    self.removeEventHandlers(this.eventHandlers);
  }
  currentPane = ACTION_MAP.get(ActionNames.MODEL_VARIABLES).paneId;
  defaultAction = [ActionNames.OUTCOME_VARIABLES];
  // These events handlers get binded when the state becomes active and will be removed when the state is left
  eventHandlers = RESULT_EVENT_HANDLERS;
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
      ActionNames.COMPUTED_VARIABLES,
      ActionNames.TARGET_VARIABLE,
      ActionNames.OUTCOME_VARIABLES,
      ActionNames.EXPORT,
    ];
    return actions.map((a) => {
      return ACTION_MAP.get(a);
    });
  }
  resetConfig(self: DataExplorerRef) {
    self.removeEventHandlers(this.eventHandlers);
  }
  currentPane = ACTION_MAP.get(ActionNames.MODEL_VARIABLES).paneId;
  defaultAction = [ActionNames.OUTCOME_VARIABLES];
  eventHandlers = PREDICTION_EVENT_HANDLERS;
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
      ActionNames.COMPUTED_VARIABLES,
      ActionNames.TARGET_VARIABLE,
      ActionNames.TRAINING_VARIABLE,
      ActionNames.OUTCOME_VARIABLES,
      ActionNames.EXPORT,
    ];
    return actions.map((a) => {
      return ACTION_MAP.get(a);
    });
  }
  resetConfig(self: DataExplorerRef) {
    self.labelName = "";
    self.removeEventHandlers(this.eventHandlers);
  }
  currentPane = ACTION_MAP.get(ActionNames.ALL_VARIABLES).paneId;
  defaultAction = [];
  // These events handlers get binded when the state becomes active and will be removed when the state is left
  eventHandlers = LABEL_EVENT_HANDLERS;
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
  COMPUTED_VARIABLES = "Computed Variables",
  EXPORT = "Export",
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
    name: ActionNames.COMPUTED_VARIABLES,
    icon: "fas fa-microchip",
    paneId: "computed",
    variables: function (self: DataExplorerRef) {
      // unique variable list
      const variables = Array.from(
        new Set([
          ...self.state.getSecondaryVariables(),
          ...self.state.getVariables(),
        ])
      );
      const isCluster = routeGetters.getDataMode(store) === DataMode.Cluster;
      // if clustering is enabled find variable with a clusterColumn
      const result = isCluster
        ? variables.filter((v) => {
            return !!v.grouping?.clusterCol;
          })
        : [];
      // add all the augmented variables
      return result.concat(
        variables.filter((v) => hasRole(v, DISTIL_ROLES.Augmented))
      );
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
    icon: "fas fa-dumbbell fa-sm",
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
  {
    name: ActionNames.EXPORT,
    icon: "fas fa-floppy-o",
    paneId: "export",
    variables: function (self: DataExplorerRef) {
      return [];
    },
  },
] as Action[];

export const ACTION_MAP = new Map(
  ACTIONS.map((a) => {
    return [a.name, a];
  })
);
// bind methods is used to bind the mixins to the data explorer
// the methods and computes require being bound to the data explorer instance
export const bindMethods = (
  obj: Record<string, Function>,
  self: DataExplorerRef
): Record<string, any> => {
  return Object.fromEntries(
    Object.keys(obj).map((k) => [k, obj[k].bind(self)])
  );
};
/**************MIXINS********************/
/*This next portion of the file is dedicated to grouping computes/methods into state objects
 and will be used in the explorer component.
The goal here is to move all of the code for the component out of the component file. 
 */
// genericMethods are the methods used across each state
// most of these methods are UI related
export const genericMethods = GENERIC_METHODS;
// genericComputes are the computes used across each state
export const genericComputes = GENERIC_COMPUTES;
// labelMethods are the methods used strictly in the label state
export const labelMethods = LABEL_METHODS;
// labelEventHandlers are the methods with an event attached to it
export const labelEventHandlers = LABEL_EVENT_HANDLERS;
// labelComputes are the computes used strictly in the label state
export const labelComputes = LABEL_COMPUTES;
export const resultMethods = RESULT_METHODS;
export const resultComputes = RESULT_COMPUTES;
export const selectComputes = SELECT_COMPUTES;
export const selectMethods = SELECT_METHODS;
export const predictionMethods = PREDICTION_METHODS;
export const predictionComputes = PREDICTION_COMPUTES;
