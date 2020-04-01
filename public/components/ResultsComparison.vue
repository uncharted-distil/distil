<template>
  <div
    class="results-slots"
    v-bind:class="{ 'one-slot': !hasHighlights, 'two-slots': hasHighlights }"
  >
    <view-type-toggle
      class="flex-shrink-0"
      v-model="viewTypeModel"
      :variables="variables"
    >
      Samples Modeled
    </view-type-toggle>

    <div v-if="hasHighlights" class="flex-grow-1">
      <results-data-slot
        instance-name="results-slot-top"
        :title="topSlotTitle"
        :data-fields="includedDataTableFields"
        :data-items="includedTableDataItems"
        :view-type="viewType"
      ></results-data-slot>
      <results-data-slot
        instance-name="results-slot-bottom"
        :title="bottomSlotTitle"
        :data-fields="excludedResultTableDataFields"
        :data-items="excludedTableDataItems"
        :view-type="viewType"
      ></results-data-slot>
    </div>
    <template v-if="!hasHighlights">
      <results-data-slot
        :title="singleSlotTitle"
        instance-name="results-slot"
        :data-fields="includedDataTableFields"
        :data-items="includedTableDataItems"
        :view-type="viewType"
      ></results-data-slot>
    </template>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import ResultsDataSlot from "../components/ResultsDataSlot";
import ViewTypeToggle from "../components/ViewTypeToggle";
import { Dictionary } from "../util/dict";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as resultsGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import { getters as predictionGetters } from "../store/predictions/module";
import { Solution } from "../store/requests/index";
import {
  Variable,
  TableRow,
  TableColumn,
  TaskTypes
} from "../store/dataset/index";
import { PREDICTION_ROUTE } from "../store/route/index";

const TABLE_VIEW = "table";
const TIMESERIES_VIEW = "timeseries";

export default Vue.extend({
  name: "results-comparison",

  components: {
    ResultsDataSlot,
    ViewTypeToggle
  },

  data() {
    return {
      viewTypeModel: TABLE_VIEW
    };
  },
  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    solutionId(): string {
      return routeGetters.getRouteSolutionId(this.$store);
    },

    solution(): Solution {
      return requestGetters.getActiveSolution(this.$store);
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    viewType(): string {
      return this.viewTypeModel;
    },

    hasHighlights(): boolean {
      const highlight = routeGetters.getDecodedHighlight(this.$store);
      return highlight && highlight.value;
    },

    includedTableDataItems(): TableRow[] {
      return this.isPrediction
        ? predictionGetters.getIncludedPredictionTableDataItems(this.$store)
        : resultsGetters.getIncludedResultTableDataItems(this.$store);
    },

    includedDataTableFields(): Dictionary<TableColumn> {
      return this.isPrediction
        ? predictionGetters.getIncludedPredictionTableDataFields(this.$store)
        : resultsGetters.getIncludedResultTableDataFields(this.$store);
    },

    numIncludedResultItems(): number {
      return this.includedTableDataItems
        ? this.includedTableDataItems.length
        : 0;
    },

    numIncludedResultErrors(): number {
      if (!this.includedTableDataItems) {
        return 0;
      }
      return this.errorCount(this.includedTableDataItems);
    },

    excludedTableDataItems(): TableRow[] {
      return this.isPrediction
        ? predictionGetters.getExcludedPredictionTableDataItems(this.$store)
        : resultsGetters.getExcludedResultTableDataItems(this.$store);
    },

    excludedResultTableDataFields(): Dictionary<TableColumn> {
      return this.isPrediction
        ? predictionGetters.getExcludedPredictionTableDataFields(this.$store)
        : resultsGetters.getExcludedResultTableDataFields(this.$store);
    },

    numExcludedResultItems(): number {
      return this.excludedTableDataItems
        ? this.excludedTableDataItems.length
        : 0;
    },

    numExcludedResultErrors(): number {
      if (!this.excludedTableDataItems) {
        return 0;
      }
      return this.errorCount(this.excludedTableDataItems);
    },

    residualThresholdMin(): number {
      return _.toNumber(routeGetters.getRouteResidualThresholdMin(this.$store));
    },

    residualThresholdMax(): number {
      return _.toNumber(routeGetters.getRouteResidualThresholdMax(this.$store));
    },

    regressionEnabled(): boolean {
      const routeArgs = routeGetters.getRouteTask(this.$store);
      return routeArgs && routeArgs.includes(TaskTypes.REGRESSION);
    },

    numRows(): number {
      return this.isPrediction
        ? predictionGetters.getPredictionDataNumRows(this.$store)
        : resultsGetters.getResultDataNumRows(this.$store);
    },

    isForecasting(): boolean {
      const routeArgs = routeGetters.getRouteTask(this.$store);
      return routeArgs && routeArgs.includes(TaskTypes.FORECASTING);
    },

    isPrediction(): boolean {
      const routePath = routeGetters.getRoutePath(this.$store);
      return routePath && routePath === PREDICTION_ROUTE;
    },

    topSlotTitle(): string {
      return this.errorTitle(
        this.numIncludedResultItems,
        this.numIncludedResultErrors
      );
    },

    bottomSlotTitle(): string {
      return this.errorTitle(
        this.numExcludedResultItems,
        this.numExcludedResultErrors
      );
    },

    singleSlotTitle(): string {
      return this.errorTitle(
        this.numExcludedResultItems,
        this.numExcludedResultErrors
      );
    }
  },
  methods: {
    errorTitle(itemCount: number, errorCount: number): string {
      const matchesLabel = `Displaying ${itemCount} of ${this.numRows}`;
      const erroneousLabel = `, including ${errorCount} <b class="erroneous-color">erroneous</b> predictions`;
      return this.isForecasting || this.isPrediction
        ? matchesLabel
        : matchesLabel + erroneousLabel;
    },
    errorCount(dataColumn: TableRow[]): number {
      if (this.isPrediction) {
        return 0;
      }
      return dataColumn.filter(item => {
        if (this.regressionEnabled) {
          if (!item[this.solution.errorKey]) {
            return false;
          }
          const err = _.toNumber(item[this.solution.errorKey].value);
          return (
            (item[this.solution.errorKey] && err < this.residualThresholdMin) ||
            err > this.residualThresholdMax
          );
        } else {
          return (
            item[this.solution.predictedKey] &&
            item[this.target] &&
            item[this.target].value !== item[this.solution.predictedKey].value
          );
        }
      }).length;
    }
  }
});
</script>

<style>
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
.matching-color {
  color: #00c6e1;
}
.other-color {
  color: #333;
}
.erroneous-color {
  color: #e05353;
}
</style>
