<template>
	<div id="distil-app">
		<navigation></navigation>
		<router-view class="view"></router-view>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import VueRouterSync from 'vuex-router-sync';
import VueObserveVisibility from 'vue-observe-visibility';
import BootstrapVue from 'bootstrap-vue';
import Navigation from './views/Navigation';
import store from './store/store';
import router from './router/router';
import { getters as routeGetters } from './store/route/module';
import { getters as appGetters, actions as appActions } from './store/app/module';
import { getters as datasetGetters, actions as datasetActions } from './store/dataset/module';
import { HOME_ROUTE, SELECT_TARGET_ROUTE, SELECT_TRAINING_ROUTE } from './store/route';
import { createRouteEntry } from './util/routes';
import { CATEGORICAL_TYPE, INTEGER_TYPE } from './util/types';
import { getComposedVariableKey } from './util/data';

import 'font-awesome/css/font-awesome.css';
import 'bootstrap-vue/dist/bootstrap-vue.css';
import './styles/bootstrap-v4beta2-custom.css';
import './styles/main.css';

// DEBUG: this is a mocked graph until we support actual graph data
import './assets/graphs/G1.gml';

Vue.use(BootstrapVue);
Vue.use(VueObserveVisibility);

// sync store and router
VueRouterSync.sync(store, router, { moduleName: 'routeModule' });

// main app component
export default Vue.extend({
	store: store,
	router: router,
	components: {
		Navigation
	},
	beforeMount() {
		// NOTE: eval only code
		appActions.fetchConfig(this.$store).then(() => {

			const path = routeGetters.getRoutePath(store);

			// if dataset / target exist in problem file, immediately route to
			// create models view.
			if (appGetters.isTask1(this.$store) && path === HOME_ROUTE) {
				const dataset = appGetters.getProblemDataset(this.$store);
				console.log(`Task 1: Routing directly to select target view with dataset=\`${dataset}\``, dataset);
				const entry = createRouteEntry(SELECT_TARGET_ROUTE, {
					dataset: dataset
				});
				this.$router.push(entry);
			}

			if (appGetters.isTask2(this.$store) && (path === HOME_ROUTE || path === SELECT_TARGET_ROUTE)) {
				const dataset = appGetters.getProblemDataset(this.$store);
				let target = appGetters.getProblemTarget(this.$store);
				const taskType = appGetters.getProblemTaskType(this.$store);
				let training = [];

				let promise = Promise.resolve();

				// TASK 2 hack
				if (taskType === 'timeSeriesForecasting') {
					promise = datasetActions.fetchVariables(this.$store, {
						dataset: dataset
					}).then(() => {
						let variables = datasetGetters.getVariables(this.$store);

						const ids = variables.filter(v => v.colType === CATEGORICAL_TYPE).map(v => v.colName);
						const idKey = getComposedVariableKey(ids);

						// set the target / training to the grouping properties
						target = idKey;
						training = ids;

						const alreadyComposed = variables.find(v => v.colName === idKey);

						let nextPromise = Promise.resolve();
						if (!alreadyComposed && ids.length > 1) {
							console.log(`Task 2: Composing ids for grouping`, ids.join(', '));
							nextPromise = datasetActions.composeVariables(this.$store, {
								dataset: dataset,
								key: idKey,
								vars: ids
							});
						}
						return nextPromise.then(() => {

							variables = datasetGetters.getVariables(this.$store);
							const existingGrouping = variables.find(v => v.colName === idKey);
							const alreadyGrouped = existingGrouping && !!existingGrouping.grouping;

							if (alreadyGrouped) {
								// grouping already exists
								return;
							}

							const yCol = target;
							const xCol = variables.filter(v => v.colName !== target && v.colType === INTEGER_TYPE).map(v => v.colName)[0];

							const grouping =  {
								type: 'timeseries',
								dataset: dataset,
								idCol: idKey,
								subIds: ids,
								hidden: [ xCol, yCol ],
								properties: {
									xCol: xCol,
									yCol: yCol,
								}
							};

							console.log(`Task 2: Fetching timeseries variables for `, dataset);
							return datasetActions.fetchVariables(this.$store, {
								dataset: dataset
							});
						});
					});
				}
				// TASK 2 hack

				console.log(`Task 2: Routing directly to create models view with dataset=\`${dataset}\` and target=\`${target}\``, dataset, target);
				promise.then(() => {
					const entry = createRouteEntry(SELECT_TRAINING_ROUTE, {
						dataset: dataset,
						target: target,
						training: training.join(',')
					});
					this.$router.push(entry);
				});
			}
		});

	}
});
</script>

<style>
</style>
