<template>
  <div class="prediction-group">
    <component
      enable-highlighting
      :summary="predictedSummary"
      :key="predictedSummary.key"
      :is="getFacetByType(predictedSummary.type)"
      :highlights="highlights"
      :enabled-type-changes="[]"
      :row-selection="rowSelection"
      :instanceName="predictionInstanceName"
      @facet-click="onCategoricalClick"
      @numerical-click="onNumericalClick"
      @range-change="onRangeChange"
    />
    <component
      v-if="!!confidenceSummary"
      enable-highlighting
      :summary="confidenceSummary"
      :key="confidenceSummary.key"
      :is="getFacetByType(confidenceSummary.type)"
      :highlights="highlights"
      :enabled-type-changes="[]"
      :row-selection="rowSelection"
      :instanceName="confidenceInstanceName"
      @facet-click="onCategoricalClick"
      @numerical-click="onNumericalClick"
      @range-change="onRangeChange"
    />
    <component
      v-if="!!rankingSummary"
      enable-highlighting
      :summary="rankingSummary"
      :key="rankingSummary.key"
      :is="getFacetByType(rankingSummary.type)"
      :highlights="highlights"
      :enabled-type-changes="[]"
      :row-selection="rowSelection"
      :instanceName="rankingInstanceName"
      @facet-click="onCategoricalClick"
      @numerical-click="onNumericalClick"
      @range-change="onRangeChange"
    />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import FacetNumerical from "../components/facets/FacetNumerical.vue";
import FacetCategorical from "../components/facets/FacetCategorical.vue";
import {
  VariableSummary,
  RowSelection,
  Highlight,
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { getFacetByType } from "../util/facets";

export default Vue.extend({
  name: "PredictionGroup",
  components: {
    FacetNumerical,
    FacetCategorical,
  },
  props: {
    confidenceSummary: Object as () => VariableSummary,
    predictedSummary: Object as () => VariableSummary,
    rankingSummary: Object as () => VariableSummary,
    highlights: Array as () => Highlight[],
  },
  computed: {
    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },
    confidenceInstanceName(): string {
      return `confidence-prediction-facet-${this.predictionDataset}`;
    },
    predictionDataset(): string {
      return routeGetters.getRoutePredictionsDataset(this.$store);
    },
    predictionInstanceName(): string {
      return `prediction-facet-${this.predictionDataset}`;
    },
    rankingInstanceName(): string {
      return `ranking-prediction-facet-${this.predictionDataset}`;
    },
  },
  methods: {
    getFacetByType(type: string) {
      return getFacetByType(type);
    },
    onCategoricalClick(
      context: string,
      key: string,
      value: string,
      dataset: string
    ) {
      this.$emit("categorical-click", { context, key, value, dataset });
    },

    onNumericalClick(
      context: string,
      key: string,
      value: { from: number; to: number },
      dataset: string
    ) {
      this.$emit("numerical-click", { context, key, value, dataset });
    },

    onRangeChange(
      context: string,
      key: string,
      value: { from: { label: string[] }; to: { label: string[] } },
      dataset: string
    ) {
      this.$emit("range-change", { context, key, value, dataset });
    },
  },
});
</script>
<style scoped>
.prediction-group {
  margin: 5px;
  padding: 10px;
  border-bottom-style: solid;
  border-bottom-color: lightgray;
  border-bottom-width: 1px;
}

.prediction-group-title {
  vertical-align: middle;
}

.prediction-group-title .badge {
  display: inline;
  vertical-align: middle;
  padding: 0.45em 0.4em 0.3em 0.4em;
}

.prediction-group-body {
  padding: 4px 0;
}
</style>
