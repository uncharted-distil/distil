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
import SparklineTimeseriesView from "./SparklineTimeseriesView.vue";
import { Dictionary } from "../util/dict";
import { TableRow, TableColumn, VariableSummary } from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { datasetGetters } from "../store";

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
      const training = routeGetters.getTrainingVariableSummaries(this.$store)(
        this.includedActive
      );
      const target = routeGetters.getTargetVariableSummaries(this.$store)(
        this.includedActive
      );
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
