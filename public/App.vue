<template>
  <div id="distil-app">
    <nav-bar />
    <router-view class="view"></router-view>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import VueRouterSync from "vuex-router-sync";
import VueObserveVisibility from "vue-observe-visibility";
import BootstrapVue from "bootstrap-vue";
import NavBar from "./components/NavBar";
import store from "./store/store";
import {
  getters as appGetters,
  actions as appActions,
} from "./store/app/module";
import router from "./router/router";

import "font-awesome/css/font-awesome.css";
import "bootstrap-vue/dist/bootstrap-vue.css";
import "./styles/uncharted-bootstrap-v4.5-custom.css";
import "./styles/main.css";

// DEBUG: this is a mocked graph until we support actual graph data
import "./assets/graphs/G1.gml";
import { TaskTypes } from "./store/dataset";

Vue.use(BootstrapVue);
Vue.use(VueObserveVisibility);

// sync store and router
VueRouterSync.sync(store, router, { moduleName: "routeModule" });

// main app component
export default Vue.extend({
  store: store,
  router: router,
  components: {
    NavBar,
  },
  beforeMount() {
    appActions.fetchConfig(this.$store);
  },
});
</script>
