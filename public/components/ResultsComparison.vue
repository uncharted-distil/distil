<template>
  <div
    class="results-slots"
    :class="{
      'one-slot': !hasHighlights || isGeoView,
      'two-slots': hasHighlights && !isGeoView,
    }"
  >
    <view-type-toggle
      v-model="viewType"
      :variables="variables"
      class="view-toggle"
      :available-variables="variables"
    >
      <p class="font-weight-bold" :class="{ 'mr-auto': !hasWeight }">Samples</p>
      <legend-weight v-if="hasWeight" class="ml-5 mr-auto" />
      <layer-selection
        :hasImageAttention="true"
        v-if="isMultiBandImage"
        class="layer-button"
      />
    </view-type-toggle>

    <div v-if="hasHighlights && !isGeoView" class="flex-grow-1">
      <results-data-slot
        instance-name="results-slot-top"
        :view-type="viewType"
      />
      <results-data-slot
        excluded
        instance-name="results-slot-bottom"
        :view-type="viewType"
      />
    </div>
    <results-data-slot
      v-else
      instance-name="results-slot"
      :view-type="viewType"
    />
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import LayerSelection from "../components/LayerSelection.vue";
import LegendWeight from "../components/LegendWeight.vue";
import ResultsDataSlot from "../components/ResultsDataSlot.vue";
import ViewTypeToggle from "../components/ViewTypeToggle.vue";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as resultsGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { Variable } from "../store/dataset/index";

const GEO_VIEW = "geo";
const TABLE_VIEW = "table";

export default Vue.extend({
  name: "results-comparison",

  components: {
    LayerSelection,
    LegendWeight,
    ResultsDataSlot,
    ViewTypeToggle,
  },

  data() {
    return {
      viewType: TABLE_VIEW,
    };
  },

  computed: {
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    /* Check if any items display on the table have a weight property. */
    hasWeight(): boolean {
      return resultsGetters.hasResultTableDataItemsWeight(this.$store);
    },

    hasHighlights(): boolean {
      const highlight = routeGetters.getDecodedHighlight(this.$store);
      return highlight && highlight.value;
    },

    isMultiBandImage(): boolean {
      return routeGetters.isMultiBandImage(this.$store);
    },
    isGeoView(): boolean {
      return this.viewType === GEO_VIEW;
    },
  },
});
</script>

<style scoped>
.results-slots {
  display: flex;
  flex-direction: column;
  flex: none;
}
.two-slots .results-data-slot {
  padding-top: 10px;
  height: 50%;
}
.one-slot .results-data-slot {
  height: 100%;
}
.layer-button {
  display: flex;
  flex-direction: column;
  flex-grow: 0;
  margin-right: 10px;
  margin-left: auto;
}
.view-toggle >>> .form-group {
  margin-bottom: 0px;
}
.view-toggle {
  flex-shrink: 0;
}
</style>
