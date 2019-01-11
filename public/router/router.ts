import Vue from 'vue';
import VueRouter from 'vue-router';
import Home from '../views/Home.vue';
import Search from '../views/Search.vue';
import SelectTarget from '../views/SelectTarget.vue';
import SelectTraining from '../views/SelectTraining.vue';
import Results from '../views/Results.vue';
import ExportSuccess from '../views/ExportSuccess.vue';
import AbortSuccess from '../views/AbortSuccess.vue';
import store from '../store/store';
import { getters as routeGetters } from '../store/route/module';
import { mutations as viewMutations } from '../store/view/module';
import { ROOT_ROUTE, HOME_ROUTE, SEARCH_ROUTE,
	SELECT_TARGET_ROUTE, SELECT_TRAINING_ROUTE, RESULTS_ROUTE,
	EXPORT_SUCCESS_ROUTE, ABORT_SUCCESS_ROUTE } from '../store/route';

Vue.use(VueRouter);

const router = new VueRouter({
	routes: [
		{ path: ROOT_ROUTE, redirect: HOME_ROUTE },
		{ path: HOME_ROUTE, component: Home },
		{ path: SEARCH_ROUTE, component: Search },
		{ path: SELECT_TARGET_ROUTE, component: SelectTarget },
		{ path: SELECT_TRAINING_ROUTE, component: SelectTraining },
		{ path: RESULTS_ROUTE, component: Results },
		{ path: EXPORT_SUCCESS_ROUTE, component: ExportSuccess },
		{ path: ABORT_SUCCESS_ROUTE, component: AbortSuccess }
	]
});

// router.beforeEach((toRoute, _, next) => {
// 	const dataset = routeGetters.getRouteDataset(store);
// 	console.log('Saving view:', toRoute.path, dataset);
// 	viewMutations.saveView(store, {
// 		view: toRoute.path,
// 		dataset: dataset,
// 		route: toRoute
// 	});
// 	next();
// });

router.afterEach((_, fromRoute) => {
	const dataset = routeGetters.getRouteDataset(store);
	console.log('Saving view:', fromRoute.path, dataset);
	viewMutations.saveView(store, {
		view: fromRoute.path,
		dataset: dataset,
		route: fromRoute
	});
});

export default router;
