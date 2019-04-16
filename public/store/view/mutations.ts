import Vue from 'vue';
import { Location } from 'vue-router';
import { ViewState, LAST_STATE } from './index';
import localStorage from 'store';

export const mutations = {
	setViewActiveDataset(state: ViewState, dataset: string) {
		state.viewActiveDataset = dataset;
	},

	setViewSelectedTarget(state: ViewState, target: string) {
		state.viewSelectedTarget = target;
	},
};
