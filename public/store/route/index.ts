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

export const ROOT_ROUTE = "/";
// export const HOME_ROUTE = "/home";
export const SEARCH_ROUTE = "/search";
export const JOIN_DATASETS_ROUTE = "/join";
export const GROUPING_ROUTE = "/grouping";
export const SELECT_TARGET_ROUTE = "/select";
export const SELECT_TRAINING_ROUTE = "/create";
export const RESULTS_ROUTE = "/results";
export const APPLY_MODEL_ROUTE = "/results";
export const PREDICTION_ROUTE = "/prediction";
export const EXPORT_SUCCESS_ROUTE = "/export-success";
export const DATA_EXPLORER_ROUTE = "/data-explorer";

export const JOINED_VARS_INSTANCE = "joinedVars";
export const TOP_VARS_INSTANCE = "topVars";
export const BOTTOM_VARS_INSTANCE = "bottomVars";
export const AVAILABLE_TRAINING_VARS_INSTANCE = "availableTrainingVars";
export const AVAILABLE_TARGET_VARS_INSTANCE = "availableTargetVars";
export const TRAINING_VARS_INSTANCE = "trainingVars";
export const TARGET_VAR_INSTANCE = "targetVar";
export const RESULT_TRAINING_VARS_INSTANCE = "resultTrainingVars";
export const RESULT_TARGET_VAR_INSTANCE = "resultTargetVar";
export const DATA_EXPLORER_VAR_INSTANCE = "dataExplorerVar";
export const VAR_MODES_INSTANCE = "varModes";
export const ROUTE_PAGE_SUFFIX = "Page";
export const ROUTE_SEARCH_SUFFIX = "Search";

export const LABEL_FEATURE_INSTANCE = "labelFeatureVars";

export const JOINED_VARS_INSTANCE_PAGE = `${JOINED_VARS_INSTANCE}${ROUTE_PAGE_SUFFIX}`;
export const AVAILABLE_TARGET_VARS_INSTANCE_PAGE = `${AVAILABLE_TARGET_VARS_INSTANCE}${ROUTE_PAGE_SUFFIX}`;
export const AVAILABLE_TRAINING_VARS_INSTANCE_PAGE = `${AVAILABLE_TRAINING_VARS_INSTANCE}${ROUTE_PAGE_SUFFIX}`;
export const TRAINING_VARS_INSTANCE_PAGE = `${TRAINING_VARS_INSTANCE}${ROUTE_PAGE_SUFFIX}`;
export const RESULT_TRAINING_VARS_INSTANCE_PAGE = `${RESULT_TRAINING_VARS_INSTANCE}${ROUTE_PAGE_SUFFIX}`;
export const DATA_EXPLORER_VARS_INSTANCE_PAGE = `${DATA_EXPLORER_VAR_INSTANCE}${ROUTE_PAGE_SUFFIX}`;
export const LABEL_FEATURE_VARS_INSTANCE_PAGE = `${LABEL_FEATURE_INSTANCE}${ROUTE_PAGE_SUFFIX}`;
export const TOP_VARS_INSTANCE_SEARCH = `${TOP_VARS_INSTANCE}${ROUTE_SEARCH_SUFFIX}`;
export const BOTTOM_VARS_INSTANCE_SEARCH = `${BOTTOM_VARS_INSTANCE}${ROUTE_SEARCH_SUFFIX}`;
export const AVAILABLE_TARGET_VARS_INSTANCE_SEARCH = `${AVAILABLE_TARGET_VARS_INSTANCE}${ROUTE_SEARCH_SUFFIX}`;
export const AVAILABLE_TRAINING_VARS_INSTANCE_SEARCH = `${AVAILABLE_TRAINING_VARS_INSTANCE}${ROUTE_SEARCH_SUFFIX}`;
export const TRAINING_VARS_INSTANCE_SEARCH = `${TRAINING_VARS_INSTANCE}${ROUTE_SEARCH_SUFFIX}`;
export const RESULT_TRAINING_VARS_INSTANCE_SEARCH = `${RESULT_TRAINING_VARS_INSTANCE}${ROUTE_SEARCH_SUFFIX}`;
export const DATA_EXPLORER_VARS_INSTANCE_SEARCH = `${DATA_EXPLORER_VAR_INSTANCE}${ROUTE_SEARCH_SUFFIX}`;

export const DATA_SIZE_DEFAULT = 100;
export const DATA_SIZE_REMOTE_SENSING_DEFAULT = 500;
