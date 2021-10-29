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
    <nav-bar @nav-event="onExplorerNav" />
    <router-view ref="view" class="view" />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import VueRouterSync from "vuex-router-sync";
import VueObserveVisibility from "vue-observe-visibility";
import EventBusPlugin from "./util/eventBus";
import BootstrapVue from "bootstrap-vue";
import NavBar from "./components/NavBar.vue";
import store from "./store/store";
import { actions as appActions } from "./store/app/module";
import vSelect from "vue-select";
import router from "./router/router";
import "@fortawesome/fontawesome-free/css/all.min.css";
import "font-awesome/css/font-awesome.css";
import "bootstrap-vue/dist/bootstrap-vue.css";
import "./styles/uncharted-bootstrap-v4.5-custom.css";
import "./styles/main.css";
import "vue-select/dist/vue-select.css";

// DEBUG: this is a mocked graph until we support actual graph data
import "./assets/graphs/G1.gml";
import { ExplorerStateNames } from "./util/explorer";
import { DataExplorerRef } from "./util/componentTypes";
import { getters as routeGetters } from "./store/route/module";

Vue.component("v-select", vSelect);
Vue.use(BootstrapVue);
Vue.use(VueObserveVisibility);
Vue.use(EventBusPlugin);
Vue.config.performance = true;
// sync store and router
VueRouterSync.sync(store, router, { moduleName: "routeModule" });
// main app component
export default Vue.extend({
  store: store,
  router: router,
  components: {
    NavBar,
  },
  methods: {
    async onExplorerNav(state: ExplorerStateNames) {
      const dataExplorer = (this.$refs.view as unknown) as DataExplorerRef;

      if (
        dataExplorer.isClone &&
        routeGetters.getDataExplorerState(store) ===
          ExplorerStateNames.LABEL_VIEW &&
        state !== ExplorerStateNames.LABEL_VIEW
      ) {
        const isDatasetSaved = await dataExplorer.isCurrentDatasetSaved();

        if (!isDatasetSaved) {
          dataExplorer.$bvModal.show(dataExplorer.unsaveModalId);
        } else {
          await dataExplorer.changeStatesByName(state);
        }
      } else {
        await dataExplorer.changeStatesByName(state);
      }
    },
  },
  beforeMount() {
    appActions.fetchConfig(this.$store);
  },
});
</script>
