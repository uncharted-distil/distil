<template>
  <div>
    <variable-facets
      class="result-target-summary"
      enable-highlighting
      :summaries="summaries"
      :instance-name="instanceName"
      :log-activity="logActivity"
    />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import VariableFacets from "./facets/VariableFacets.vue";
import { getters as resultsGetters } from "../store/results/module";
import { RESULT_TARGET_VAR_INSTANCE } from "../store/route/index";
import { VariableSummary } from "../store/dataset/index";
import { Activity } from "../util/userEvents";

export default Vue.extend({
  name: "result-target-variable",

  components: {
    VariableFacets,
  },

  data() {
    return {
      hasDefaultedAlready: false,
      logActivity: Activity.MODEL_SELECTION,
    };
  },

  computed: {
    resultTargetSummary(): VariableSummary {
      return resultsGetters.getTargetSummary(this.$store);
    },

    summaries(): VariableSummary[] {
      return this.resultTargetSummary ? [this.resultTargetSummary] : [];
    },

    instanceName(): string {
      return RESULT_TARGET_VAR_INSTANCE;
    },
  },
});
</script>

<style>
.result-target-summary
  .variable-facets-container
  .facets-root-container
  .facets-group-container
  .facets-group {
  box-shadow: none;
}
.result-target-variable {
  flex-shrink: 0;
  margin-bottom: 0;
}
</style>
