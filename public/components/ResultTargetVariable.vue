<template>
  <div>
    <variable-facets
      class="result-target-summary"
      enable-highlighting
      :summaries="summaries"
      :instance-name="instanceName"
      :log-activity="logActivity"
    >
    </variable-facets>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import VariableFacets from "./facets/VariableFacets.vue";
import { getters as routeGetters } from "../store/route/module";
import { getters as resultsGetters } from "../store/results/module";
import {
  getNumericalFacetValue,
  getCategoricalFacetValue,
  TOP_RANGE_HIGHLIGHT,
} from "../util/facets";
import { updateHighlight, clearHighlight } from "../util/highlights";
import { RESULT_TARGET_VAR_INSTANCE } from "../store/route/index";
import {
  Variable,
  VariableSummary,
  Highlight,
  RowSelection,
} from "../store/dataset/index";
import { isNumericType, TIMESERIES_TYPE } from "../util/types";
import { Activity } from "../util/userEvents";

export default Vue.extend({
  name: "result-target-variable",

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

    targetVariable(): Variable {
      return routeGetters.getTargetVariable(this.$store);
    },

    resultTargetSummary(): VariableSummary {
      return resultsGetters.getTargetSummary(this.$store);
    },

    summaries(): VariableSummary[] {
      return this.resultTargetSummary ? [this.resultTargetSummary] : [];
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    hasFilters(): boolean {
      return routeGetters.getDecodedFilters(this.$store).length > 0;
    },

    instanceName(): string {
      return RESULT_TARGET_VAR_INSTANCE;
    },

    defaultHighlightType(): string {
      return TOP_RANGE_HIGHLIGHT;
    },
  },

  data() {
    return {
      hasDefaultedAlready: false,
      logActivity: Activity.MODEL_SELECTION,
    };
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
