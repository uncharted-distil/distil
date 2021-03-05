<!--

    Copyright © 2021 Uncharted Software Inc.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
-->

<template>
  <div id="distil-app">
    <nav-bar />
    <router-view class="view" />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import VueRouterSync from "vuex-router-sync";
import VueObserveVisibility from "vue-observe-visibility";
import BootstrapVue from "bootstrap-vue";
import NavBar from "./components/NavBar.vue";
import store from "./store/store";
import { actions as appActions } from "./store/app/module";
import router from "./router/router";

import "font-awesome/css/font-awesome.css";
import "bootstrap-vue/dist/bootstrap-vue.css";
import "./styles/uncharted-bootstrap-v4.5-custom.css";
import "./styles/main.css";

// DEBUG: this is a mocked graph until we support actual graph data
import "./assets/graphs/G1.gml";

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
