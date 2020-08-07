<template>
  <div
    class="results-slots"
    :class="{ 'one-slot': !hasHighlights, 'two-slots': hasHighlights }"
  >
    <view-type-toggle
      v-model="viewType"
      :variables="variables"
      class="view-toggle"
    >
      <p class="font-weight-bold" :class="{ 'mr-auto': !hasWeight }">Samples</p>
      <legend-weight v-if="hasWeight" class="ml-5 mr-auto" />
      <layer-selection v-if="isRemoteSensing" class="layer-button" />
    </view-type-toggle>

    <div v-if="hasHighlights" class="flex-grow-1">
      <results-data-slot
        instance-name="results-slot-top"
        :data-fields="includedDataTableFields"
        :data-items="includedTableDataItems"
        :view-type="viewType"
      />
      <results-data-slot
        instance-name="results-slot-bottom"
        :data-fields="excludedResultTableDataFields"
        :data-items="excludedTableDataItems"
        :view-type="viewType"
      />
    </div>
    <results-data-slot
      v-else
      instance-name="results-slot"
      :data-fields="includedDataTableFields"
      :data-items="includedTableDataItems"
      :view-type="viewType"
    />
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import LayerSelection from "../components/LayerSelection";
import LegendWeight from "../components/LegendWeight";
import ResultsDataSlot from "../components/ResultsDataSlot";
import ViewTypeToggle from "../components/ViewTypeToggle";
import { Dictionary } from "../util/dict";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as resultsGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { Variable, TableRow, TableColumn } from "../store/dataset/index";

const TABLE_VIEW = "table";

export default Vue.extend({
  name: "results-comparison",

  components: {
    LayerSelection,
    LegendWeight,
    ResultsDataSlot,
    ViewTypeToggle
  },

  data() {
    return {
      viewType: TABLE_VIEW
    };
  },

  computed: {
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    /* Check if any items display on the table have a weight property. */
    hasWeight(): boolean {
      const data = this.includedTableDataItems ?? [];
      return data.some(item =>
        Object.values(item).some(variable => variable.hasOwnProperty("weight"))
      );
    },

    hasHighlights(): boolean {
      const highlight = routeGetters.getDecodedHighlight(this.$store);
      return highlight && highlight.value;
    },

    includedTableDataItems(): TableRow[] {
      return resultsGetters.getIncludedResultTableDataItems(this.$store);
    },

    includedDataTableFields(): Dictionary<TableColumn> {
      return resultsGetters.getIncludedResultTableDataFields(this.$store);
    },

    excludedTableDataItems(): TableRow[] {
      return resultsGetters.getExcludedResultTableDataItems(this.$store);
    },

    excludedResultTableDataFields(): Dictionary<TableColumn> {
      return resultsGetters.getExcludedResultTableDataFields(this.$store);
    },

    numRows(): number {
      return resultsGetters.getResultDataNumRows(this.$store);
    },

    isRemoteSensing(): boolean {
      return routeGetters.isRemoteSensing(this.$store);
    }
  }
});
</script>

<!-- used in generated strings so can't be scoped -->
<style>
.matching-color {
  color: #255dcc;
}
.other-color {
  color: #333;
}
.erroneous-color {
  color: #e05353;
}
</style>

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
