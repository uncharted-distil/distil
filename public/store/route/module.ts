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
	getRouteCreateRequestId: read(moduleGetters.getRouteCreateRequestId),
	getRoutePipelinetId: read(moduleGetters.getRoutePipelinetId),
	getRouteFilters: read(moduleGetters.getRouteFilters),
	getRouteResultFilters: read(moduleGetters.getRouteResultFilters),
	getRouteResidualThreshold: read(moduleGetters.getRouteResidualThreshold),
	getDecodedFilters: read(moduleGetters.getDecodedFilters),
	getDecodedResultsFilters: read(moduleGetters.getDecodedResultsFilters),
}


