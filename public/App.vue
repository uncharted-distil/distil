<template>
	<div id="distil-app">
		<navigation></navigation>
		<keep-alive include="select-target-view,select-training-view,results-view">
			<router-view class="view"></router-view>
		</keep-alive>
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
import { HOME_ROUTE, SELECT_TARGET_ROUTE, SELECT_TRAINING_ROUTE } from './store/route';
import { createRouteEntry } from './util/routes';

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
				const target = appGetters.getProblemTarget(this.$store);
				console.log(`Task 2: Routing directly to create models view with dataset=\`${dataset}\` and target=\`${target}\``, dataset, target);
				const entry = createRouteEntry(SELECT_TRAINING_ROUTE, {
					dataset: dataset,
					target: target
				});
				this.$router.push(entry);
			}
		});

	}
});
</script>

<style>
</style>
