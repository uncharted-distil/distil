<template>
  <div class="results-data-slot">
    <p class="results-data-slot-summary" v-if="hasResults">
      Displaying
      <data-size
        :currentSize="numItems"
        :total="numRows"
        @submit="onDataSizeSubmit"
      />
      of {{ numRows
      }}<template v-if="!isForecasting"
        >, including {{ numErrors }}
        <strong class="erroneous-color">erroneous</strong> predictions
      </template>
    </p>

    <div class="results-data-slot-container" :class="{ pending: !hasData }">
      <div class="results-data-no-results" v-if="isPending || hasNoResults">
        <div v-if="isPending" v-html="spinnerHTML"></div>
        <p v-if="hasNoResults">No results available</p>
      </div>

      <component
        :is="viewComponent"
        :data-fields="dataFields"
        :data-items="dataItems"
        :instance-name="instanceName"
        :summaries="trainingSummaries"
        @tileClicked="onTileClick"
      />
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";
import DataSize from "../components/buttons/DataSize.vue";
import GeoPlot, { TileClickData } from "./GeoPlot.vue";
import ImageMosaic from "./ImageMosaic.vue";
import ResultsDataTable from "./ResultsDataTable.vue";
import ResultsTimeseriesView from "./ResultsTimeseriesView.vue";
import {
  Highlight,
  TableRow,
  TableColumn,
  TaskTypes,
  Variable,
  RowSelection,
  VariableSummary,
} from "../store/dataset/index";
import { Solution, SOLUTION_ERRORED } from "../store/requests/index";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import {
  actions as resultsActions,
  getters as resultsGetters,
} from "../store/results/module";
import { actions as viewActions } from "../store/view/module";
import { getters as requestsGetters } from "../store/requests/module";
import { Dictionary } from "../util/dict";
import { updateTableRowSelection } from "../util/row";
import { spinnerHTML } from "../util/spinner";
import { getVariableSummariesByState, searchVariables } from "../util/data";
import { isGeoLocatedType } from "../util/types";
import { Filter, INCLUDE_FILTER } from "../util/filters";

const GEO_VIEW = "geo";
const GRAPH_VIEW = "graph";
const IMAGE_VIEW = "image";
const TABLE_VIEW = "table";
const TIMESERIES_VIEW = "timeseries";

/**
 * Display results based on a VIEW type.
 * @param {Boolean} excluded - display only excluded results
 */
export default Vue.extend({
  name: "results-data-slot",

  components: {
    DataSize,
    GeoPlot,
    ImageMosaic,
    ResultsDataTable,
    ResultsTimeseriesView,
  },

  props: {
    instanceName: String,
    viewType: String,
    excluded: Boolean,
  },

  data() {
    return {
      GEO_VIEW: GEO_VIEW,
      GRAPH_VIEW: GRAPH_VIEW,
      IMAGE_VIEW: IMAGE_VIEW,
      TABLE_VIEW: TABLE_VIEW,
      TIMESERIES_VIEW: TIMESERIES_VIEW,
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    solution(): Solution {
      return requestsGetters.getActiveSolution(this.$store);
    },

    solutionId(): string {
      return this.solution?.solutionId;
    },

    solutionHasErrored(): boolean {
      return this.solution
        ? this.solution.progress === SOLUTION_ERRORED
        : false;
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    isForecasting(): boolean {
      const routeArgs = routeGetters.getRouteTask(this.$store);
      return routeArgs && routeArgs.includes(TaskTypes.FORECASTING);
    },

    isPending(): boolean {
      return !this.hasData && !this.solutionHasErrored;
    },

    hasNoResults(): boolean {
      return (
        this.solutionHasErrored || (this.hasData && this.items.length === 0)
      );
    },

    hasResults(): boolean {
      return this.hasData && this.items.length > 0;
    },

    hasData(): boolean {
      return !!this.dataItems;
    },

    items(): TableRow[] {
      return updateTableRowSelection(
        this.dataItems,
        this.rowSelection,
        this.instanceName
      );
    },

    dataItems(): TableRow[] {
      if (this.excluded) {
        return resultsGetters.getExcludedResultTableDataItems(this.$store);
      }
      // included or none get the same data
      return resultsGetters.getIncludedResultTableDataItems(this.$store);
    },

    dataFields(): Dictionary<TableColumn> {
      if (this.excluded) {
        return resultsGetters.getExcludedResultTableDataFields(this.$store);
      }
      // included or none get the same data
      return resultsGetters.getIncludedResultTableDataFields(this.$store);
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    spinnerHTML(): string {
      return spinnerHTML();
    },

    /* Select which component to display the data. */
    viewComponent() {
      if (this.viewType === GEO_VIEW) return "GeoPlot";
      if (this.viewType === IMAGE_VIEW) return "ImageMosaic";
      if (this.viewType === TABLE_VIEW) return "ResultsDataTable";
      if (this.viewType === TIMESERIES_VIEW) return "ResultsTimeseriesView";
    },

    /* Count the number of items */
    numItems(): number {
      return this.dataItems ? this.dataItems.length : 0;
    },

    /* Get the total number of items available */
    numRows(): number {
      if (this.excluded) {
        return resultsGetters.getExcludedResultTableDataCount(this.$store);
      }
      // included or none get the same data
      return resultsGetters.getIncludedResultTableDataCount(this.$store);
    },

    numErrors(): number {
      return this.dataItems ? this.errorCount(this.dataItems) : 0;
    },

    regressionEnabled(): boolean {
      const routeArgs = routeGetters.getRouteTask(this.$store);
      return routeArgs && routeArgs.includes(TaskTypes.REGRESSION);
    },

    residualThresholdMin(): number {
      return _.toNumber(routeGetters.getRouteResidualThresholdMin(this.$store));
    },

    residualThresholdMax(): number {
      return _.toNumber(routeGetters.getRouteResidualThresholdMax(this.$store));
    },
    resultTrainingVarsSearch(): string {
      return routeGetters.getRouteResultTrainingVarsSearch(this.$store);
    },
    trainingVariables(): Variable[] {
      return searchVariables(
        requestsGetters.getActiveSolutionTrainingVariables(this.$store),
        this.resultTrainingVarsSearch
      );
    },
    trainingSummaries(): VariableSummary[] {
      const summaryDictionary = resultsGetters.getTrainingSummariesDictionary(
        this.$store
      );
      const pageIndex = routeGetters.getRouteResultTrainingVarsPage(
        this.$store
      );
      const currentSummaries = getVariableSummariesByState(
        pageIndex,
        this.trainingVariables.length,
        this.trainingVariables,
        summaryDictionary
      );
      return currentSummaries.filter((cs) => {
        return isGeoLocatedType(cs.varType);
      });
    },
  },

  methods: {
    errorCount(dataColumn: TableRow[]): number {
      return dataColumn.filter((item) => {
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
    },

    /* When the user request to fetch a different size of data. */
    onDataSizeSubmit(dataSize: number) {
      const args: any = {
        dataset: this.dataset,
        highlight: this.highlight,
        size: dataSize,
        solutionId: this.solutionId,
      };

      if (this.excluded) {
        resultsActions.fetchExcludedResultTableData(this.$store, args);
      } else {
        resultsActions.fetchIncludedResultTableData(this.$store, args);
      }
    },
    async onTileClick(data: TileClickData) {
      // build filter
      const filter: Filter = {
        displayName: data.displayName,
        key: data.key,
        maxX: data.bounds[1][1],
        maxY: data.bounds[0][0],
        minX: data.bounds[0][1],
        minY: data.bounds[1][0],
        mode: INCLUDE_FILTER,
        type: data.type,
      };
      // fetch surrounding tiles
      await viewActions.updateResultAreaOfInterest(this.$store, filter);
      // get area of interest
      const inner = resultsGetters.getAreaOfInterestInnerDataItems(this.$store);
      const outer = resultsGetters.getAreaOfInterestOuterDataItems(this.$store);
      // callback after fetch
      data.callback(inner, outer);
      return;
    },
  },
});
</script>

<style scoped>
.results-data-slot-summary {
  flex-shrink: 0;
  font-size: 90%;
  margin: 0;
}

.results-data-slot {
  display: flex;
  flex-direction: column;
}

.results-data-slot-container {
  position: relative;
  display: flex;
  flex-grow: 1;
}

.results-data-no-results {
  position: absolute;
  display: block;
  top: 0;
  height: 100%;
  width: 100%;
  padding: 32px;
  text-align: center;
  opacity: 1;
  z-index: 1;
}

.erroneous-color {
  color: var(--red);
}

.pending {
  opacity: 0.5;
}
</style>
