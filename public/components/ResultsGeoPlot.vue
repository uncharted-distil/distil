<template>
  <geo-plot
    :instance-name="instanceName"
    :data-fields="dataFields"
    :data-items="dataItems"
  >
  </geo-plot>
</template>

<script lang="ts">
import Vue from "vue";
import GeoPlot from "./GeoPlot";
import { getters as datasetGetters } from "../store/dataset/module";
import { Dictionary } from "../util/dict";
import { TableColumn, TableRow } from "../store/dataset/index";

export default Vue.extend({
  name: "results-geo-plot",

  components: {
    GeoPlot
  },

  props: {
    instanceName: String as () => string,
    dataItems: Array as () => any[],
    dataFields: Object as () => Dictionary<TableColumn>
  }
});
</script>

<style></style>
