import Vue from 'vue';
import App from './App';
import store from './store';
import BootstrapVue from 'bootstrap-vue';

Vue.use(BootstrapVue);

new Vue({
	el: '#app',
	store,
	template: '<App/>',
	components: { App }
});
