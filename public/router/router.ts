import Vue from 'vue';
import VueRouter from 'vue-router';
import Home from '../views/Home';
import Search from '../views/Search';
import JoinDatasets from '../views/JoinDatasets';
import SelectTarget from '../views/SelectTarget';
import SelectTraining from '../views/SelectTraining';
import Results from '../views/Results';
import ExportSuccess from '../views/ExportSuccess';
import AbortSuccess from '../views/AbortSuccess';
import store from '../store/store';
import { getters as routeGetters } from '../store/route/module';
import { mutations as viewMutations } from '../store/view/module';
import { saveView } from '../util/view';
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
	let key = routeGetters.getRouteDataset(store);
	if (key === '' || !key)  {
		key = routeGetters.getRouteJoinDatasetsHash(store);
	}
	if (key) {
		console.log(`Saving view: ${fromRoute.path} for key ${key}`);
	} else {
		console.log(`Saving view: ${fromRoute.path}`);
	}
	saveView({
		view: fromRoute.path,
		key: key,
		route: fromRoute
	});
});

export default router;
