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

// Activity, sub-activity and feature ID types used in the user event reporting logs

export enum Activity {
  DATA_PREPARATION = "DATA_PREPARATION",
  PROBLEM_DEFINITION = "PROBLEM_DEFINITION",
  MODEL_SELECTION = "MODEL_SELECTION",
  PREDICTION_ANALYSIS = "PREDICTION_ANALYSIS",
}

export enum SubActivity {
  APP_LAUNCH = "APP_LAUNCH",
  DATA_OPEN = "DATA_OPEN",
  DATA_EXPLORATION = "DATA_EXPLORATION",
  DATA_AUGMENTATION = "DATA_AUGMENTATION",
  DATA_TRANSFORMATION = "DATA_TRANSFORMATION",
  PROBLEM_SPECIFICATION = "PROBLEM_SPECIFICATION",
  MODEL_SEARCH = "MODEL_SEARCH",
  MODEL_SUMMARIZATION = "MODEL_SUMMARIZATION",
  MODEL_COMPARISON = "MODEL_COMPARISON",
  MODEL_EXPLANATION = "MODEL_EXPLANATION",
  MODEL_EXPORT = "MODEL_EXPORT",
  IMPORT_INFERENCE = "IMPORT_INFERENCE",
  MODEL_PREDICTIONS = "MODEL_PREDICTIONS",
  MODEL_SAVE = "MODEL_SAVE",
}

export enum Feature {
  SEARCH_DATASETS = "SEARCH_DATASETS",
  SELECT_DATASET = "SELECT_DATASET",
  SELECT_TARGET = "SELECT_TARGET",
  RETYPE_FEATURE = "RETYPE_FEATURE",
  RANK_FEATURES = "RANK_FEATURES",
  GEOCODE_FEATURES = "GEOCODE_FEATURES",
  CLUSTER_DATA = "CLUSTER_DATA",
  JOIN_DATASETS = "JOIN_DATASETS",
  OUTLIER_FEATURES = "OUTLIER_FEATURES",
  ADD_FEATURE = "ADD_FEATURE",
  ADD_ALL_FEATURES = "ADD_ALL_FEATURES",
  REMOVE_FEATURE = "REMOVE_FEATURE",
  REMOVE_ALL_FEATURES = "REMOVE_ALL_FEATURES",
  CHANGE_HIGHLIGHT = "CHANGE_HIGHLIGHT",
  CHANGE_SELECTION = "CHANGE_SELECTION",
  CHANGE_ERROR_THRESHOLD = "CHANGE_ERROR_THRESHOLD",
  FILTER_DATA = "FILTER_DATA",
  UNFILTER_DATA = "UNFILTER_DATA",
  SEARCH_FEATURES = "SEARCH_FEATURES",
  CREATE_MODEL = "CREATE_MODEL",
  SELECT_MODEL = "SELECT_MODEL",
  EXPORT_MODEL = "EXPORT_MODEL",
  SELECT_PREDICTIONS = "SELECT_PREDICTIONS",
}
