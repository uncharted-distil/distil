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
  <div class="results-data-slot">
    <p v-if="hasResults && !isGeoView" class="results-data-slot-summary">
      Displaying
      <data-size
        :current-size="numItems"
        :total="numRows"
        @submit="onDataSizeSubmit"
      />
      of {{ numRows }}
      <template v-if="!isForecasting">
        , including {{ numErrors }}
        <strong class="erroneous-color">erroneous</strong> predictions
      </template>
    </p>
    <p v-else-if="isGeoView" class="selection-data-slot-summary">
      Selected Area Coverage:
      <strong class="matching-color">{{ areaCoverage }}km<sup>2</sup></strong>
    </p>

    <div class="results-data-slot-container" :class="{ pending: !hasData }">
      <div v-if="isPending || hasNoResults" class="results-data-no-results">
        <div v-if="isPending" v-html="spinnerHTML" />
        <p v-if="hasNoResults">No results available</p>
      </div>

      <component
        :is="viewComponent"
        :data-fields="dataFields"
        :data-items="dataItems"
        :baseline-items="baselineItems"
        :baseline-map="baselineMap"
        :instance-name="instanceName"
        :summaries="trainingSummaries"
        :area-of-interest-items="{ inner: inner, outer: outer }"
        :confidence-access-func="colorTile"
        :is-result="true"
        :dataset="dataset"
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
import { Solution, SolutionStatus } from "../store/requests/index";
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
import {
  getVariableSummariesByState,
  searchVariables,
  totalAreaCoverage,
} from "../util/data";
import { isGeoLocatedType } from "../util/types";
import { Filter, INCLUDE_FILTER } from "../util/filters";
import { overlayRouteEntry } from "../util/routes";

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
  name: "ResultsDataSlot",

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

    highlights(): Highlight[] {
      return routeGetters.getDecodedHighlights(this.$store);
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
    areaCoverage(): number {
      return totalAreaCoverage(this.items, this.variables);
    },
    confidenceSummaries(): VariableSummary {
      return resultsGetters.getConfidenceSummaries(this.$store).filter((cf) => {
        return cf.solutionId === this.solutionId;
      })[0];
    },
    rankSummary(): VariableSummary {
      return resultsGetters.getRankingSummaries(this.$store).filter((rank) => {
        return rank.solutionId === this.solutionId;
      })[0];
    },
    solutionHasErrored(): boolean {
      return this.solution
        ? this.solution.progress === SolutionStatus.SOLUTION_ERRORED
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
    baselineMap(): Dictionary<number> {
      const result = {};
      const base = this.baselineItems ?? [];
      base.forEach((item, i) => {
        result[item.d3mIndex] = i;
      });
      return result;
    },
    baselineItems(): TableRow[] {
      return resultsGetters.getFullIncludedResultTableDataItems(this.$store);
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
    viewComponent(): string {
      if (this.viewType === GEO_VIEW) return "GeoPlot";
      if (this.viewType === IMAGE_VIEW) return "ImageMosaic";
      if (this.viewType === TABLE_VIEW) return "ResultsDataTable";
      if (this.viewType === TIMESERIES_VIEW) return "ResultsTimeseriesView";
      console.error(`viewType ${this.viewType} invalid`);
      return "";
    },
    isGeoView(): boolean {
      return this.viewType === GEO_VIEW;
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
    inner(): TableRow[] {
      return resultsGetters.getAreaOfInterestInnerDataItems(this.$store);
    },
    outer(): TableRow[] {
      return resultsGetters.getAreaOfInterestOuterDataItems(this.$store);
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
    dataSize(): number {
      return routeGetters.getRouteDataSize(this.$store);
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
    colorTile(d) {
      if (d.rank !== undefined) {
        return d.rank.value / this.rankSummary.baseline.extrema.max;
      }
      if (d.confidence !== undefined) {
        return d.confidence.value;
      }
      return undefined;
    },
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
      if (this.dataSize !== dataSize) {
        const entry = overlayRouteEntry(this.$route, { dataSize: dataSize });
        this.$router.push(entry).catch((err) => console.warn(err));
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
        set: "",
      };
      // fetch surrounding tiles
      await viewActions.updateResultAreaOfInterest(this.$store, filter);
    },
  },
  watch: {
    dataSize() {
      const args: any = {
        dataset: this.dataset,
        highlights: this.highlights,
        size: this.dataSize,
        solutionId: this.solutionId,
      };
      if (this.excluded) {
        resultsActions.fetchExcludedResultTableData(this.$store, args);
      } else {
        resultsActions.fetchIncludedResultTableData(this.$store, args);
      }
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
  max-height: 98%;
  height: 98%;
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
