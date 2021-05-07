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
  <div>
    <variable-facets
      class="target-summary"
      :enable-color-scales="geoVarExists"
      enable-highlighting
      enable-type-change
      :summaries="targetSummaries"
      :instance-name="instanceName"
      :log-activity="logActivity"
    />

    <!-- Dropdown to select a positive label for Binary Classification task -->
    <positive-label v-if="labels" :labels="labels" />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import PositiveLabel from "./buttons/PositiveLabel.vue";
import VariableFacets from "./facets/VariableFacets.vue";
import { getters as routeGetters } from "../store/route/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { TARGET_VAR_INSTANCE } from "../store/route/index";
import { getAllVariablesSummaries } from "../util/data";
import { VariableSummary, Variable } from "../store/dataset/index";
import { Activity } from "../util/userEvents";
import { isGeoLocatedType } from "../util/types";

export default Vue.extend({
  name: "TargetVariable",

  components: {
    PositiveLabel,
    VariableFacets,
  },

  data() {
    return {
      instanceName: TARGET_VAR_INSTANCE,
      logActivity: Activity.DATA_PREPARATION,
    };
  },

  computed: {
    targetSummaries(): VariableSummary[] {
      return routeGetters.getTargetVariableSummaries(this.$store);
    },
    variables(): Variable[] {
      return routeGetters.getTrainingVariables(this.$store);
    },
    geoVarExists(): boolean {
      const varSumsDict = datasetGetters.getIncludedVariableSummariesDictionary(
        this.$store
      );
      const varSums = getAllVariablesSummaries(this.variables, varSumsDict);
      return varSums.some((v) => {
        return isGeoLocatedType(v.type);
      });
    },
    labels(): string[] {
      // make sure we are only on a binary classification task
      if (!routeGetters.isBinaryClassification(this.$store)) return;

      // retreive the target variable buckets
      const buckets = this.targetSummaries?.[0]?.baseline?.buckets;
      if (!buckets) return;

      // use the buckets keys as labels
      return buckets.map((bucket) => bucket.key);
    },
  },
});
</script>
