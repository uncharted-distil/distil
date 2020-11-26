import { Module } from "vuex";
import { Route } from "vue-router";
import { getters as moduleGetters } from "./getters";
import { DistilState } from "../store";
import { getStoreAccessors } from "vuex-typescript";

export const routeModule: Module<Route, DistilState> = {
  getters: moduleGetters,
};

const { read } = getStoreAccessors<Route, DistilState>(null);

export const getters = {
  getRoute: read(moduleGetters.getRoute),
  getRoutePath: read(moduleGetters.getRoutePath),
  getRouteTerms: read(moduleGetters.getRouteTerms),
  getRouteDataset: read(moduleGetters.getRouteDataset),
  getRouteInclude: read(moduleGetters.getRouteInclude),
  getRouteJoinDatasets: read(moduleGetters.getRouteJoinDatasets),
  getJoinDatasetColumnA: read(moduleGetters.getJoinDatasetColumnA),
  getJoinDatasetColumnB: read(moduleGetters.getJoinDatasetColumnB),
  getBaseColumnSuggestions: read(moduleGetters.getBaseColumnSuggestions),
  getJoinColumnSuggestions: read(moduleGetters.getJoinColumnSuggestions),
  getJoinAccuracy: read(moduleGetters.getJoinAccuracy),
  getDecodedJoinDatasetsFilterParams: read(
    moduleGetters.getDecodedJoinDatasetsFilterParams
  ),
  getRouteJoinDatasetsHash: read(moduleGetters.getRouteJoinDatasetsHash),
  getJoinDatasetsVariables: read(moduleGetters.getJoinDatasetsVariables),
  getJoinDatasetsVariableSummaries: read(
    moduleGetters.getJoinDatasetsVariableSummaries
  ),
  getRouteTrainingVariables: read(moduleGetters.getRouteTrainingVariables),
  getRouteIsTrainingVariablesRanked: read(
    moduleGetters.getRouteIsTrainingVariablesRanked
  ),
  getRouteIsClusterGenerated: read(moduleGetters.getRouteIsClusterGenerated),
  getDecodedTrainingVariableNames: read(
    moduleGetters.getDecodedTrainingVariableNames
  ),

  getRouteJoinDatasetsVarsPage: read(
    moduleGetters.getRouteJoinDatasetsVarsPage
  ),
  getRouteAvailableTargetVarsPage: read(
    moduleGetters.getRouteAvailableTargetVarsPage
  ),
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
  getRouteJoinDatasetsVarsSearch: read(
    moduleGetters.getRouteJoinDatasetsVarsSearch
  ),
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
  getRouteSolutionId: read(moduleGetters.getRouteSolutionId),
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
  getTrainingVariables: read(moduleGetters.getTrainingVariables),
  getTrainingVariableSummaries: read(
    moduleGetters.getTrainingVariableSummaries
  ),
  getTargetVariable: read(moduleGetters.getTargetVariable),
  getTargetVariableSummaries: read(moduleGetters.getTargetVariableSummaries),
  getAvailableVariables: read(moduleGetters.getAvailableVariables),
  getDecodedHighlight: read(moduleGetters.getDecodedHighlight),
  getDecodedRowSelection: read(moduleGetters.getDecodedRowSelection),
  getDataMode: read(moduleGetters.getDataMode),
  getDecodedVarModes: read(moduleGetters.getDecodedVarModes),
  getActiveSolutionIndex: read(moduleGetters.getActiveSolutionIndex),
  getGeoCenter: read(moduleGetters.getGeoCenter),
  getGeoZoom: read(moduleGetters.getGeoZoom),
  getGroupingType: read(moduleGetters.getGroupingType),
  getRouteTask: read(moduleGetters.getRouteTask),
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
  isPageSelectTarget: read(moduleGetters.isPageSelectTarget),
  isPageSelectTraining: read(moduleGetters.isPageSelectTraining),
};
