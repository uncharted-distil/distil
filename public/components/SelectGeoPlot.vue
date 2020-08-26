<template>
  <geo-plot
    :instance-name="instanceName"
    :data-fields="fields"
    :data-items="items"
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
  name: "select-geo-plot",

  components: {
    GeoPlot,
  },

  props: {
    instanceName: String as () => string,
    includedActive: Boolean as () => boolean,
  },

  computed: {
    fields(): Dictionary<TableColumn> {
      return this.includedActive
        ? datasetGetters.getIncludedTableDataFields(this.$store)
        : datasetGetters.getExcludedTableDataFields(this.$store);
    },

    items(): TableRow[] {
      return this.includedActive
        ? datasetGetters.getIncludedTableDataItems(this.$store)
        : datasetGetters.getExcludedTableDataItems(this.$store);
    },
  },
});
</script>

<style></style>
