<template>
  <sparkline-timeseries-view
    :instance-name="instanceName"
    :include-active="includedActive"
    :variable-summaries="variableSummaries"
    :items="items"
    :fields="fields"
  >
  </sparkline-timeseries-view>
</template>

<script lang="ts">
import Vue from "vue";
import SparklineTimeseriesView from "./SparklineTimeseriesView";
import { Dictionary } from "../util/dict";
import { TableRow, TableColumn, VariableSummary } from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { getters as datasetGetters } from "../store/dataset/module";

export default Vue.extend({
  name: "select-timeseries-view",

  components: {
    SparklineTimeseriesView,
  },

  props: {
    instanceName: String as () => string,
    includedActive: Boolean as () => boolean,
  },

  computed: {
    variableSummaries(): VariableSummary[] {
      const training = routeGetters.getTrainingVariableSummaries(this.$store);
      const target = routeGetters.getTargetVariableSummaries(this.$store);
      return target ? training.concat(target) : training;
    },

    items(): TableRow[] {
      return this.includedActive
        ? datasetGetters.getIncludedTableDataItems(this.$store)
        : datasetGetters.getExcludedTableDataItems(this.$store);
    },

    fields(): Dictionary<TableColumn> {
      return this.includedActive
        ? datasetGetters.getIncludedTableDataFields(this.$store)
        : datasetGetters.getExcludedTableDataFields(this.$store);
    },
  },
});
</script>

<style></style>
