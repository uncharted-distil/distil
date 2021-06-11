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

import { Route } from "vue-router";
import { Module } from "vuex";
import { getStoreAccessors } from "vuex-typescript";
import { DistilState } from "../store";
import { getters as moduleGetters } from "./getters";

export const routeModule: Module<Route, DistilState> = {
  getters: moduleGetters,
};

const { read } = getStoreAccessors<Route, DistilState>(null);

export const getters = {
  getRoute: read(moduleGetters.getRoute),
  getRoutePath: read(moduleGetters.getRoutePath),
  getRouteTerms: read(moduleGetters.getRouteTerms),
  getRouteDataset: read(moduleGetters.getRouteDataset),
  getPriorPath: read(moduleGetters.getPriorPath),
  getRouteJoinDatasets: read(moduleGetters.getRouteJoinDatasets),
  getJoinDatasetColumnA: read(moduleGetters.getJoinDatasetColumnA),
  getJoinDatasetColumnB: read(moduleGetters.getJoinDatasetColumnB),
  getBaseColumnSuggestions: read(moduleGetters.getBaseColumnSuggestions),
  getJoinColumnSuggestions: read(moduleGetters.getJoinColumnSuggestions),
  getJoinAccuracy: read(moduleGetters.getJoinAccuracy),
  getDecodedJoinDatasetsFilterParams: read(
    moduleGetters.getDecodedJoinDatasetsFilterParams
  ),
  getDecodedJoinDatasetsHighlight: read(
    moduleGetters.getDecodedJoinDatasetsHighlight
  ),
  getAnnotationHasChanged: read(moduleGetters.getAnnotationHasChanged),
  getRouteJoinDatasetsHash: read(moduleGetters.getRouteJoinDatasetsHash),
  getJoinDatasetsVariables: read(moduleGetters.getJoinDatasetsVariables),
  getExploreVariables: read(moduleGetters.getExploreVariables),
  getRouteTrainingVariables: read(moduleGetters.getRouteTrainingVariables),
  getRouteIsTrainingVariablesRanked: read(
    moduleGetters.getRouteIsTrainingVariablesRanked
  ),
  getRouteIsClusterGenerated: read(moduleGetters.getRouteIsClusterGenerated),
  isOutlierApplied: read(moduleGetters.isOutlierApplied),
  getDecodedTrainingVariableNames: read(
    moduleGetters.getDecodedTrainingVariableNames
  ),
  getRouteLabel: read(moduleGetters.getRouteLabel),
  hasGeoData: read(moduleGetters.hasGeoData),
  getRouteJoinDatasetsVarsPage: read(
    moduleGetters.getRouteJoinDatasetsVarsPage
  ),
  getRouteAvailableTargetVarsPage: read(
    moduleGetters.getRouteAvailableTargetVarsPage
  ),
  getLabelFeaturesVarsPage: read(moduleGetters.getLabelFeaturesVarsPage),
  getRouteAvailableTrainingVarsPage: read(
    moduleGetters.getRouteAvailableTrainingVarsPage
  ),
  getRouteTrainingVarsPage: read(moduleGetters.getRouteTrainingVarsPage),
  getRouteResultTrainingVarsPage: read(
    moduleGetters.getRouteResultTrainingVarsPage
  ),
  getRouteDataExplorerVarsPage: read(
    moduleGetters.getRouteDataExplorerVarsPage
  ),
  getAllRoutePages: read(moduleGetters.getAllRoutePages),
  getRouteTopVarsSearch: read(moduleGetters.getRouteTopVarsSearch),
  getRouteBottomVarsSearch: read(moduleGetters.getRouteBottomVarsSearch),
  getRouteAvailableTargetVarsSearch: read(
    moduleGetters.getRouteAvailableTargetVarsSearch
  ),
  getRouteAvailableTrainingVarsSearch: read(
    moduleGetters.getRouteAvailableTrainingVarsSearch
  ),
  getRouteTrainingVarsSearch: read(moduleGetters.getRouteTrainingVarsSearch),
  getRouteResultTrainingVarsSearch: read(
    moduleGetters.getRouteResultTrainingVarsSearch
  ),
  getRouteDataExplorerVarsSearch: read(
    moduleGetters.getRouteDataExplorerVarsSearch
  ),
  getAllSearchesByRoute: read(moduleGetters.getAllSearchesByRoute),
  getAllSearchesByQueryString: read(moduleGetters.getAllSearchesByQueryString),
  getRouteDataSize: read(moduleGetters.getRouteDataSize),
  getRouteTargetVariable: read(moduleGetters.getRouteTargetVariable),
  getRoutePreviousTarget: read(moduleGetters.getRoutePreviousTarget),
  getRouteSolutionId: read(moduleGetters.getRouteSolutionId),
  getRouteOpenSolutions: read(moduleGetters.getRouteOpenSolutions),
  getRouteFilters: read(moduleGetters.getRouteFilters),
  getRouteHighlight: read(moduleGetters.getRouteHighlight),
  getRouteRowSelection: read(moduleGetters.getRouteRowSelection),
  getRouteProduceRequestId: read(moduleGetters.getRouteProduceRequestId),
  getRouteResidualThresholdMin: read(
    moduleGetters.getRouteResidualThresholdMin
  ),
  getRouteResidualThresholdMax: read(
    moduleGetters.getRouteResidualThresholdMax
  ),
  getDecodedFilters: read(moduleGetters.getDecodedFilters),
  getDecodedSolutionRequestFilterParams: read(
    moduleGetters.getDecodedSolutionRequestFilterParams
  ),
  getJoinPairs: read(moduleGetters.getJoinPairs),
  getJoinInfo: read(moduleGetters.getJoinInfo),
  getTrainingVariables: read(moduleGetters.getTrainingVariables),
  getTrainingVariableSummaries: read(
    moduleGetters.getTrainingVariableSummaries
  ),
  getTargetVariable: read(moduleGetters.getTargetVariable),
  getTargetVariableSummaries: read(moduleGetters.getTargetVariableSummaries),
  getAvailableVariables: read(moduleGetters.getAvailableVariables),
  getDecodedHighlights: read(moduleGetters.getDecodedHighlights),
  getDecodedRowSelection: read(moduleGetters.getDecodedRowSelection),
  getDataMode: read(moduleGetters.getDataMode),
  getDecodedVarModes: read(moduleGetters.getDecodedVarModes),
  getActiveSolutionIndex: read(moduleGetters.getActiveSolutionIndex),
  getGeoCenter: read(moduleGetters.getGeoCenter),
  getGeoZoom: read(moduleGetters.getGeoZoom),
  getGroupingType: read(moduleGetters.getGroupingType),
  getRouteTask: read(moduleGetters.getRouteTask),
  getColorScale: read(moduleGetters.getColorScale),
  getColorScaleVariable: read(moduleGetters.getColorScaleVariable),
  getRouteFittedSolutionId: read(moduleGetters.getRouteFittedSolutionId),
  getRoutePredictionsDataset: read(moduleGetters.getRoutePredictionsDataset),
  isSingleSolution: read(moduleGetters.isSingleSolution),
  isApplyModel: read(moduleGetters.isApplyModel),
  isMultiBandImage: read(moduleGetters.isMultiBandImage),
  isGeoSpatial: read(moduleGetters.isGeoSpatial),
  isTimeseries: read(moduleGetters.isTimeseries),
  getBandCombinationId: read(moduleGetters.getBandCombinationId),
  getModelLimit: read(moduleGetters.getModelLimit),
  getModelTimeLimit: read(moduleGetters.getModelTimeLimit),
  getModelQuality: read(moduleGetters.getModelQuality),
  getModelMetrics: read(moduleGetters.getModelMetrics),
  getImageAttention: read(moduleGetters.getImageAttention),
  getRouteTrainTestSplit: read(moduleGetters.getRouteTrainTestSplit),
  getRouteTimestampSplit: read(moduleGetters.getRouteTimestampSplit),
  isPageSelectTarget: read(moduleGetters.isPageSelectTarget),
  isPageSelectTraining: read(moduleGetters.isPageSelectTraining),
  getRoutePane: read(moduleGetters.getRoutePane),
  hasOrderBy: read(moduleGetters.hasOrderBy),
  getOrderBy: read(moduleGetters.getOrderBy),
  isBinaryClassification: read(moduleGetters.isBinaryClassification),
  getPositiveLabel: read(moduleGetters.getPositiveLabel),
};
