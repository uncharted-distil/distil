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
	getRouteTrainingVariables: read(moduleGetters.getRouteTrainingVariables),
	getRouteTargetVariable: read(moduleGetters.getRouteTargetVariable),
	getRoutePipelineId: read(moduleGetters.getRoutePipelineId),
	getRouteFilters: read(moduleGetters.getRouteFilters),
	getRouteHighlightRoot: read(moduleGetters.getRouteHighlightRoot),
	getRouteRowSelection: read(moduleGetters.getRouteRowSelection),
	getRouteResidualThresholdMin: read(moduleGetters.getRouteResidualThresholdMin),
	getRouteResidualThresholdMax: read(moduleGetters.getRouteResidualThresholdMax),
	getDecodedFilterParams: read(moduleGetters.getDecodedFilterParams),
	getDecodedHighlightRoot: read(moduleGetters.getDecodedHighlightRoot),
	getDecodedRowSelection: read(moduleGetters.getDecodedRowSelection)

}
