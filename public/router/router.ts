import Vue from 'vue';
import VueRouter from 'vue-router';
import Home from '../views/Home.vue';
import Search from '../views/Search.vue';
import JoinDatasets from '../views/JoinDatasets.vue';
import SelectTarget from '../views/SelectTarget.vue';
import SelectTraining from '../views/SelectTraining.vue';
import Results from '../views/Results.vue';
import ExportSuccess from '../views/ExportSuccess.vue';
import AbortSuccess from '../views/AbortSuccess.vue';
import store from '../store/store';
import { getters as routeGetters } from '../store/route/module';
import { mutations as viewMutations } from '../store/view/module';
import { ROOT_ROUTE, HOME_ROUTE, SEARCH_ROUTE, JOIN_DATASETS_ROUTE,
	SELECT_TARGET_ROUTE, SELECT_TRAINING_ROUTE, RESULTS_ROUTE,
	EXPORT_SUCCESS_ROUTE, ABORT_SUCCESS_ROUTE } from '../store/route';

Vue.use(VueRouter);

const router = new VueRouter({
	routes: [
		{ path: ROOT_ROUTE, redirect: HOME_ROUTE },
		{ path: HOME_ROUTE, component: Home },
		{ path: SEARCH_ROUTE, component: Search },
		{ path: JOIN_DATASETS_ROUTE, component: JoinDatasets },
		{ path: SELECT_TARGET_ROUTE, component: SelectTarget },
		{ path: SELECT_TRAINING_ROUTE, component: SelectTraining },
		{ path: RESULTS_ROUTE, component: Results },
		{ path: EXPORT_SUCCESS_ROUTE, component: ExportSuccess },
		{ path: ABORT_SUCCESS_ROUTE, component: AbortSuccess }
	]
});

router.afterEach((_, fromRoute) => {
	let dataset = routeGetters.getRouteDataset(store);
	if (dataset === '' || !dataset)  {
		dataset = routeGetters.getRouteJoinDatasetsHash(store);
	}
	if (dataset) {
		console.log(`Saving view: ${fromRoute.path} for dataset ${dataset}`);
	} else {
		console.log(`Saving view: ${fromRoute.path}`);
	}
	viewMutations.saveView(store, {
		view: fromRoute.path,
		dataset: dataset,
		route: fromRoute
	});
});

export default router;
