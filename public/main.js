import Vue from 'vue';
import VueRouter from 'vue-router';
import VueRouterSync from 'vuex-router-sync';
import Dataset from './views/Dataset';
import Search from './views/Search';
import Navigation from './views/Navigation';
import store from './store';
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
		{ path: '/', redirect: '/search' },
		{ path: '/search', component: Search },
		{ path: '/dataset', component: Dataset },
	]
});

// sync store and router
VueRouterSync.sync(store, router);

new Vue({
	store,
	router,
	components: {
		Navigation
	},
	template: `
		<div id="distil-app" class="container-fluid">
			<navigation/>
			<router-view class="view"></router-view>
		</div>`
}).$mount('#app');

// init the websocket connection
store.dispatch('openWebSocketConnection', '/ws');
