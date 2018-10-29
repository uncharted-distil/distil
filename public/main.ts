import Vue from 'vue';
import VueRouter from 'vue-router';
import VueRouterSync from 'vuex-router-sync';
import VueObserveVisibility from 'vue-observe-visibility'
import Home from './views/Home.vue';
import Search from './views/Search.vue';
import SelectTarget from './views/SelectTarget.vue';
import SelectTraining from './views/SelectTraining.vue';
import Results from './views/Results.vue';
import Navigation from './views/Navigation.vue';
import ExportSuccess from './views/ExportSuccess.vue';
import AbortSuccess from './views/AbortSuccess.vue';
import { getters as routeGetters } from './store/route/module';
import { mutations as viewMutations } from './store/view/module';
import { getters as appGetters, actions as appActions } from './store/app/module';
import { ROOT_ROUTE, HOME_ROUTE, SEARCH_ROUTE, SELECT_ROUTE, CREATE_ROUTE, RESULTS_ROUTE, EXPORT_SUCCESS_ROUTE, ABORT_SUCCESS_ROUTE } from './store/route';
import store from './store/store';
import { setStore } from './store/storeProvider';
import BootstrapVue from 'bootstrap-vue';
import { createRouteEntry } from './util/routes';

import 'bootstrap-vue/dist/bootstrap-vue.css';

import './styles/bootstrap-v4beta2-custom.css';
import './styles/main.css';

import './assets/graphs/G1.gml';

Vue.use(VueRouter);
Vue.use(BootstrapVue);
Vue.use(VueObserveVisibility);

export const router = new VueRouter({
	routes: [
		{ path: ROOT_ROUTE, redirect: HOME_ROUTE },
		{ path: HOME_ROUTE, component: Home },
		{ path: SEARCH_ROUTE, component: Search },
		{ path: SELECT_ROUTE, component: SelectTarget },
		{ path: CREATE_ROUTE, component: SelectTraining },
		{ path: RESULTS_ROUTE, component: Results },
		{ path: EXPORT_SUCCESS_ROUTE, component: ExportSuccess },
		{ path: ABORT_SUCCESS_ROUTE, component: AbortSuccess }
	]
});

router.beforeEach((route, _, next) => {
	const dataset = route.query ? route.query.dataset : routeGetters.getRouteDataset(store);
	viewMutations.saveView(store, {
		view: route.path,
		dataset: dataset,
		route: route
	});
	next();
});

// sync store and router
VueRouterSync.sync(store, router, { moduleName: 'routeModule' });

// create globally accessible store so that we don't have to have reference
// to the component to use it.  Importing the instance directly leads to ciculcar
// dependency errors from webpack, so we use a store provider and lazy init.
setStore(store)

// init app
new Vue({
	store,
	router,
	components: {
		Navigation
	},
	template: `
		<div id="distil-app">
			<navigation></navigation>
			<router-view class="view"></router-view>
		</div>`,
	beforeMount() {
		// NOTE: eval only code
		appActions.fetchConfig(this.$store).then(() => {
			const path = routeGetters.getRoutePath(store);
			// if dataset / target exist in problem file, immediately route to
			// create models view.
			if (appGetters.isTask1(this.$store) && path == HOME_ROUTE) {
				const dataset = appGetters.getProblemDataset(this.$store);
				console.log(`Task 1: Routing directly to select target view with dataset=\`${dataset}\``, dataset);
				const entry = createRouteEntry(SELECT_ROUTE, {
					dataset: dataset
				});
				this.$router.push(entry);
			}

			if (appGetters.isTask2(this.$store) && (path == HOME_ROUTE || path == SELECT_ROUTE)) {
				const dataset = appGetters.getProblemDataset(this.$store);
				const target = appGetters.getProblemTarget(this.$store);
				console.log(`Task 2: Routing directly to create models view with dataset=\`${dataset}\` and target=\`${target}\``, dataset, target);
				const entry = createRouteEntry(CREATE_ROUTE, {
					dataset: dataset,
					target: target
				});
				this.$router.push(entry);
			}
		});

	}
}).$mount('#app');
