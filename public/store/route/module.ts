import { Module } from 'vuex';
import { Route } from 'vue-router';
import { getters as moduleGetters } from './getters';
import { DistilState } from '../store';
import { getStoreAccessors } from 'vuex-typescript';

export const routeModule: Module<Route, DistilState> = {
	getters: moduleGetters
};

const { read } = getStoreAccessors<Route, DistilState>(null);

export const getters = {
	getRoute: read(moduleGetters.getRoute),
	getRoutePath: read(moduleGetters.getRoutePath),
	getRouteTerms: read(moduleGetters.getRouteTerms),
	getRouteDataset: read(moduleGetters.getRouteDataset),
	getRouteJoinDatasets: read(moduleGetters.getRouteJoinDatasets),
	getJoinDatasetColumnA: read(moduleGetters.getJoinDatasetColumnA),
	getJoinDatasetColumnB: read(moduleGetters.getJoinDatasetColumnB),
	getJoinAccuracy: read(moduleGetters.getJoinAccuracy),
	getDecodedJoinDatasetsFilterParams: read(moduleGetters.getDecodedJoinDatasetsFilterParams),
	getRouteJoinDatasetsHash: read(moduleGetters.getRouteJoinDatasetsHash),
	getRouteJoinDatasetsVarsParge: read(moduleGetters.getRouteJoinDatasetsVarsParge),
	getJoinDatasetsVariables: read(moduleGetters.getJoinDatasetsVariables),
	getJoinDatasetsVariableSummaries: read(moduleGetters.getJoinDatasetsVariableSummaries),
	getRouteTrainingVariables: read(moduleGetters.getRouteTrainingVariables),
	getRouteAvailableTrainingVarsPage: read(moduleGetters.getRouteAvailableTrainingVarsPage),
	getRouteTrainingVarsPage: read(moduleGetters.getRouteTrainingVarsPage),
	getRouteResultTrainingVarsPage: read(moduleGetters.getRouteResultTrainingVarsPage),
	getRouteTargetVariable: read(moduleGetters.getRouteTargetVariable),
	getRouteSolutionId: read(moduleGetters.getRouteSolutionId),
	getRouteFilters: read(moduleGetters.getRouteFilters),
	getRouteHighlightRoot: read(moduleGetters.getRouteHighlightRoot),
	getRouteRowSelection: read(moduleGetters.getRouteRowSelection),
	getRouteResidualThresholdMin: read(moduleGetters.getRouteResidualThresholdMin),
	getRouteResidualThresholdMax: read(moduleGetters.getRouteResidualThresholdMax),
	getDecodedFilters: read(moduleGetters.getDecodedFilters),
	getDecodedFilterParams: read(moduleGetters.getDecodedFilterParams),
	getTrainingVariables: read(moduleGetters.getTrainingVariables),
	getTrainingVariableSummaries: read(moduleGetters.getTrainingVariableSummaries),
	getTargetVariable: read(moduleGetters.getTargetVariable),
	getTargetVariableSummaries: read(moduleGetters.getTargetVariableSummaries),
	getAvailableVariables: read(moduleGetters.getAvailableVariables),
	getAvailableVariableSummaries: read(moduleGetters.getAvailableVariableSummaries),
	getDecodedHighlightRoot: read(moduleGetters.getDecodedHighlightRoot),
	getDecodedRowSelection: read(moduleGetters.getDecodedRowSelection),
	getActiveSolutionIndex: read(moduleGetters.getActiveSolutionIndex),
	getGeoCenter: read(moduleGetters.getGeoCenter),
	getGeoZoom: read(moduleGetters.getGeoZoom)
};
