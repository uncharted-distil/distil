<template>
  <div :class="{ included: includedActive, excluded: !includedActive }">
    <variable-facets
      class="target-summary"
      enable-highlighting
      enable-type-change
      :summaries="targetSummaries"
      :instance-name="instanceName"
      :log-activity="logActivity"
    />
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import VariableFacets from "./facets/VariableFacets.vue";
import { getters as routeGetters } from "../store/route/module";
import { TARGET_VAR_INSTANCE } from "../store/route/index";
import { VariableSummary } from "../store/dataset/index";
import { Activity } from "../util/userEvents";

export default Vue.extend({
  name: "target-variable",

  components: {
    VariableFacets,
  },

  data() {
    return {
      hasDefaultedAlready: false,
      logActivity: Activity.DATA_PREPARATION,
    };
  },

  computed: {
    targetSummaries(): VariableSummary[] {
      return routeGetters.getTargetVariableSummaries(this.$store);
    },

    instanceName(): string {
      return TARGET_VAR_INSTANCE;
    },
  },
});
</script>

<style>
.target-summary
  .variable-facets-container
  .facets-root-container
  .facets-group-container
  .facets-group {
  box-shadow: none;
}
</style>
