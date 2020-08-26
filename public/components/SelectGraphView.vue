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
