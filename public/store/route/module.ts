import { Module } from 'vuex';
import { Route } from 'vue-router';
import { getters } from './getters';

export const routeModule: Module<Route, any> = {
	getters: getters
};
