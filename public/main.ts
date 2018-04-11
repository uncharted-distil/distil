import Vue from 'vue';
import VueRouter from 'vue-router';
import VueRouterSync from 'vuex-router-sync';
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
import { actions as pipelineActions, getters as pipelineGetters } from './store/pipelines/module';
import { ROOT_ROUTE, HOME_ROUTE, SEARCH_ROUTE, SELECT_ROUTE, CREATE_ROUTE, RESULTS_ROUTE, EXPORT_SUCCESS_ROUTE, ABORT_SUCCESS_ROUTE } from './store/route/index';
import store from './store/store';
import BootstrapVue from 'bootstrap-vue';

import './assets/favicons/apple-touch-icon.png';
import './assets/favicons/favicon-32x32.png';
import './assets/favicons/favicon-16x16.png';
import './assets/favicons/manifest.json';
import './assets/favicons/safari-pinned-tab.svg';

import 'bootstrap-vue/dist/bootstrap-vue.css';

import './styles/bootstrap-v4beta2-custom.css';
import './styles/main.css';

Vue.use(VueRouter);
Vue.use(BootstrapVue);

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
			<router-view class="view" v-if="hasActiveSession"></router-view>
		</div>`,
	computed: {
		sessionId(): string {
			return pipelineGetters.getPipelineSessionID(this.$store)
		},
		hasActiveSession(): boolean {
			return pipelineGetters.hasActiveSession(this.$store)
		}
	},
	mounted() {
		pipelineActions.startPipelineSession(this.$store, {
			sessionId: this.sessionId
		});
		appActions.fetchVersion(this.$store);
	}
}).$mount('#app');
