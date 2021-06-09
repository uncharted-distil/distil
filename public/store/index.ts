import {
  actions as dActions,
  getters as dGetters,
  mutations as dMutations,
} from "./dataset/module";
import {
  actions as pActions,
  getters as pGetters,
  mutations as pMutations,
} from "./predictions/module";
import {
  actions as reqActions,
  getters as reqGetters,
  mutations as reqMutations,
} from "./requests/module";
import {
  actions as rActions,
  getters as rGetters,
  mutations as rMutations,
} from "./results/module";
import {
  actions as vActions,
  getters as vGetters,
  mutations as vMutations,
} from "./view/module";

/**
 * Dataset Store Module Exports
 */
export const datasetGetters = dGetters;
export const datasetActions = dActions;
export const datasetMutations = dMutations;

/**
 * Result Store Module Exports
 */
export const resultGetters = rGetters;
export const resultActions = rActions;
export const resultMutations = rMutations;

/**
 * Prediction Store Module Exports
 */
export const predictionGetters = pGetters;
export const predictionActions = pActions;
export const predictionMutations = pMutations;

/**
 * View Store Module Exports
 */
export const viewGetters = vGetters;
export const viewActions = vActions;
export const viewMutations = vMutations;

/**
 * Request Store Module Exports
 */
export const requestGetters = reqGetters;
export const requestActions = reqActions;
export const requestMutations = reqMutations;
