import { Module } from "vuex";
import { Route } from "vue-router";
import { getters as moduleGetters } from "./getters";
import { DistilState } from "../store";
import { getStoreAccessors } from "vuex-typescript";
import { modelModule } from "../model/module";

export const routeModule: Module<Route, DistilState> = {
  getters: moduleGetters
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
  getRouteJoinDatasetsVarsParge: read(
    moduleGetters.getRouteJoinDatasetsVarsParge
  ),
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
  getAvailableVariableSummaries: read(
    moduleGetters.getAvailableVariableSummaries
  ),
  getDecodedHighlight: read(moduleGetters.getDecodedHighlight),
  getDecodedRowSelection: read(moduleGetters.getDecodedRowSelection),
  getDecodedVarModes: read(moduleGetters.getDecodedVarModes),
  getActiveSolutionIndex: read(moduleGetters.getActiveSolutionIndex),
  getGeoCenter: read(moduleGetters.getGeoCenter),
  getGeoZoom: read(moduleGetters.getGeoZoom),
  getGroupingType: read(moduleGetters.getGroupingType),
  getRouteTask: read(moduleGetters.getRouteTask),
  getRouteFittedSolutionID: read(moduleGetters.getRouteFittedSolutionId),
  getRoutePredictionsDataset: read(moduleGetters.getRoutePredictionsDataset),
  isSingleSolution: read(moduleGetters.isSingleSolution),
  isApplyModel: read(moduleGetters.isApplyModel),
  isRemoteSensing: read(moduleGetters.isRemoteSensing),
  isTimeseries: read(moduleGetters.isTimeseries),
  getBandCombinationId: read(moduleGetters.getBandCombinationId)
};
