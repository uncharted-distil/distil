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
  <div class="select-graph-view">
    <div id="graph-container"></div>
    <div v-if="!isLoaded" v-html="spinnerHTML"></div>
  </div>
</template>

<script lang="ts">
import sigma from "sigma";
import Vue from "vue";
import { Dictionary } from "../util/dict";
import { circleSpinnerHTML } from "../util/spinner";
import { getters as routeGetters } from "../store/route/module";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";

export default Vue.extend({
  name: "select-graph-view",

  props: {
    instanceName: String as () => string,
    // graphUrl: String as () => string
  },

  data() {
    return {
      s: null,
      graphUrl: "mock",
    };
  },

  computed: {
    files(): Dictionary<any> {
      return datasetGetters.getFiles(this.$store);
    },
    isLoaded(): boolean {
      return !!this.files[this.graphUrl];
    },
    graph(): Object {
      return this.files[this.graphUrl];
    },
    spinnerHTML(): string {
      return circleSpinnerHTML();
    },
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
  },

  methods: {
    injectGraph() {
      this.s = new sigma({
        graph: this.graph,
        container: "graph-container",
      });
    },
  },

  mounted() {
    datasetActions
      .fetchGraph(this.$store, {
        dataset: this.dataset,
        url: this.graphUrl,
      })
      .then(() => {
        this.injectGraph();
      });
  },
});
</script>

<style>
.select-graph-view {
  flex: 1;
}

#graph-container {
  position: relative;
  height: 100%;
  width: 100%;
}
</style>
