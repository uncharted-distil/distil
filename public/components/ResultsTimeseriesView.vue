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
  <sparkline-timeseries-view
    disable-highlighting
    :instance-name="instanceName"
    :include-active="include"
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
    include: { type: Boolean as () => boolean, default: true },
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
  },
});
</script>
