import { Module } from 'vuex';
import { Route } from 'vue-router';
import { DistilState } from '../index';
import { getters } from './getters';

export const routeModule: Module<Route, DistilState> = {
	getters: getters
};
