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
    <search-bar
      class="mb-3"
      :variables="allVariables"
      :highlights="routeHighlight"
      @lex-query="updateFilterAndHighlightFromLexQuery"
    />
    <div v-if="hasHighlights && !isGeoView" class="h-80">
      <results-data-slot
        instance-name="results-slot-top"
        :view-type="viewType"
        class="h-50"
      />
      <results-data-slot
        excluded
        instance-name="results-slot-bottom"
        :view-type="viewType"
        class="h-50"
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
import SearchBar from "../components/layout/SearchBar.vue";
import LayerSelection from "../components/LayerSelection.vue";
import LegendWeight from "../components/LegendWeight.vue";
import ResultsDataSlot from "../components/ResultsDataSlot.vue";
import ViewTypeToggle from "../components/ViewTypeToggle.vue";
import { resultSummariesToVariables } from "../util/summaries";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as resultsGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { Variable } from "../store/dataset/index";
import { updateHighlight, UPDATE_ALL } from "../util/highlights";
import { lexQueryToFiltersAndHighlight } from "../util/lex";

const GEO_VIEW = "geo";
const TABLE_VIEW = "table";

export default Vue.extend({
  name: "results-comparison",

  components: {
    LayerSelection,
    LegendWeight,
    ResultsDataSlot,
    ViewTypeToggle,
    SearchBar,
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
      const highlights = routeGetters.getDecodedHighlights(this.$store);
      return highlights.length > 0;
    },
    allVariables(): Variable[] {
      const solutionID = routeGetters.getRouteSolutionId(this.$store);
      const resultVariables = resultSummariesToVariables(solutionID);
      return datasetGetters
        .getAllVariables(this.$store)
        .concat(resultVariables);
    },
    routeHighlight(): string {
      return routeGetters.getRouteHighlight(this.$store);
    },
    isMultiBandImage(): boolean {
      return routeGetters.isMultiBandImage(this.$store);
    },
    isGeoView(): boolean {
      return this.viewType === GEO_VIEW;
    },
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
  },
  methods: {
    updateFilterAndHighlightFromLexQuery(lexQuery) {
      const lqfh = lexQueryToFiltersAndHighlight(lexQuery, this.dataset);
      updateHighlight(this.$router, lqfh.highlights, UPDATE_ALL);
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
.two-slots {
  padding-top: 10px;
  height: 100%;
}
.one-slot {
  height: 100%;
}
.layer-button {
  display: flex;
  flex-direction: column;
  flex-grow: 0;
  margin-right: 10px;
  margin-left: 10px;
}
.h-80 {
  height: 80vh !important;
}
.view-toggle >>> .form-group {
  margin-bottom: 0px;
}
.view-toggle {
  flex-shrink: 0;
}
</style>
