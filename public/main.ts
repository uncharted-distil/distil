import Vue from 'vue';
import VueRouter from 'vue-router';
import VueRouterSync from 'vuex-router-sync';
import Home from './views/Home.vue';
import Search from './views/Search.vue';
import Explore from './views/Explore.vue';
import Select from './views/Select.vue';
import Pipelines from './views/Pipelines.vue';
import Results from './views/results.vue';
import Navigation from './views/Navigation.vue';
import store from './store/store';
import BootstrapVue from 'bootstrap-vue';

import './assets/favicons/apple-touch-icon.png';
import './assets/favicons/favicon-32x32.png';
import './assets/favicons/favicon-16x16.png';
import './assets/favicons/manifest.json';
import './assets/favicons/safari-pinned-tab.svg';

import 'bootstrap/dist/css/bootstrap.css';
import 'bootstrap-vue/dist/bootstrap-vue.css';

import './styles/main.css';

Vue.use(VueRouter);
Vue.use(BootstrapVue);

const router = new VueRouter({
	routes: [
		{ path: '/', redirect: '/home' },
		{ path: '/home', component: Home },
		{ path: '/search', component: Search },
		{ path: '/explore', component: Explore },
		{ path: '/select', component: Select },
		{ path: '/pipelines', component: Pipelines },
		{ path: '/results', component: Results }
	]
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
			<navigation/>
			<router-view class="view"></router-view>
		</div>`
}).$mount('#app');
