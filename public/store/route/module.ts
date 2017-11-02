import { Module } from 'vuex';
import { Route } from 'vue-router';
import { getters as moduleGetters } from './getters';
import { getStoreAccessors } from 'vuex-typescript';

export const routeModule: Module<Route, any> = {
	getters: moduleGetters
};

const { read } = getStoreAccessors<Route, any>(null);

export const getters = {
	getRoute: read(moduleGetters.getRoute),
	getRoutePath: read(moduleGetters.getRoutePath),
	getRouteTerms: read(moduleGetters.getRouteTerms),
	getRouteDataset: read(moduleGetters.getRouteDataset),
	getRouteTrainingVariables: read(moduleGetters.getRouteTrainingVariables),
	getRouteTargetVariable: read(moduleGetters.getRouteTargetVariable),
	getRouteCreateRequestId: read(moduleGetters.getRouteCreateRequestId),
	getRouteResultId: read(moduleGetters.getRouteResultId),
	getRouteFilters: read(moduleGetters.getRouteFilters),
	getRouteResultFilters: read(moduleGetters.getRouteResultFilters),
	getRouteFacetsPage: read(moduleGetters.getRouteFacetsPage),
	getRouteResidualThreshold: read(moduleGetters.getRouteResidualThreshold),
	getFilters: read(moduleGetters.getFilters),
	getResultsFilters: read(moduleGetters.getResultsFilters),
	getTrainingVariables: read(moduleGetters.getTrainingVariables),
	getTargetVariable: read(moduleGetters.getTargetVariable)
}


