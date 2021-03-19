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
      enable-highlighting
      enable-type-change
      :summaries="targetSummaries"
      :instance-name="instanceName"
      :log-activity="logActivity"
    />

    <positive-label
      v-if="isBinaryClassification"
      :target-summary="targetSummaries[0]"
    />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import PositiveLabel from "./buttons/PositiveLabel.vue";
import VariableFacets from "./facets/VariableFacets.vue";
import { getters as routeGetters } from "../store/route/module";
import { TARGET_VAR_INSTANCE } from "../store/route/index";
import { VariableSummary } from "../store/dataset/index";
import { Activity } from "../util/userEvents";

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
    isBinaryClassification(): boolean {
      return routeGetters.isBinaryClassification(this.$store);
    },

    targetSummaries(): VariableSummary[] {
      return routeGetters.getTargetVariableSummaries(this.$store);
    },
  },
});
</script>
