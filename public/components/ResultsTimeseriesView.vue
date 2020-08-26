<template>
  <sparkline-timeseries-view
    disable-highlighting
    :instance-name="instanceName"
    :include-active="includedActive"
    :variable-summaries="variableSummaries"
    :items="items"
    :fields="fields"
    :predictedCol="predictedCol"
  />
</template>

<script lang="ts">
import Vue from "vue";
import SparklineTimeseriesView from "./SparklineTimeseriesView";
import { Dictionary } from "../util/dict";
import { VariableSummary, TableColumn, TableRow } from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import { getters as resultsGetters } from "../store/results/module";
import { Solution } from "../store/requests/index";

export default Vue.extend({
  name: "results-timeseries-view",

  components: {
    SparklineTimeseriesView,
  },

  props: {
    dataItems: Array as () => TableRow[],
    dataFields: Object as () => Dictionary<TableColumn>,
    instanceName: String as () => string,
  },

  computed: {
    variableSummaries(): VariableSummary[] {
      const training = resultsGetters.getTrainingSummaries(this.$store);
      const target = resultsGetters.getTargetSummary(this.$store);
      return target ? training.concat(target) : training;
    },

    solution(): Solution {
      return requestGetters.getActiveSolution(this.$store);
    },

    predictedCol(): string {
      return this.solution ? this.solution.predictedKey : "";
    },

    includedActive(): boolean {
      return routeGetters.getRouteInclude(this.$store);
    },
  },
});
</script>
