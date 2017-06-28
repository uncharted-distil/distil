import Vue from 'vue';
import VueRouter from 'vue-router';
import Dataset from './views/Dataset';
import Search from './views/Search';
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

new Vue({
	store,
	router,
	template: '<router-view class="view"></router-view>'
}).$mount('#app');
