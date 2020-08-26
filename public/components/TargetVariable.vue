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
import {
  getNumericalFacetValue,
  getCategoricalFacetValue,
  getTimeseriesFacetValue,
  TOP_RANGE_HIGHLIGHT,
} from "../util/facets";
import { TARGET_VAR_INSTANCE } from "../store/route/index";
import { Variable, VariableSummary, Highlight } from "../store/dataset/index";
import { updateHighlight } from "../util/highlights";
import { isNumericType, TIMESERIES_TYPE, DATE_TIME_TYPE } from "../util/types";
import { Activity } from "../util/userEvents";

export default Vue.extend({
  name: "target-variable",

  components: {
    VariableFacets,
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    includedActive(): boolean {
      return routeGetters.getRouteInclude(this.$store);
    },

    targetVariable(): Variable {
      return routeGetters.getTargetVariable(this.$store);
    },

    targetSummaries(): VariableSummary[] {
      return routeGetters.getTargetVariableSummaries(this.$store);
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    hasFilters(): boolean {
      return routeGetters.getDecodedFilters(this.$store).length > 0;
    },

    instanceName(): string {
      return TARGET_VAR_INSTANCE;
    },

    defaultHighlightType(): string {
      return TOP_RANGE_HIGHLIGHT;
    },
  },

  data() {
    return {
      hasDefaultedAlready: false,
      logActivity: Activity.DATA_PREPARATION,
    };
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
