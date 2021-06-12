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
  <div class="predictions-data-slot">
    <view-type-toggle
      v-model="viewTypeModel"
      class="view-toggle"
      :variables="variables"
    >
      <p class="font-weight-bold mr-auto">Samples Predicted</p>
      <layer-selection v-if="isMultiBandImage" class="layer-button" />
    </view-type-toggle>
    <search-bar
      class="mb-3"
      :variables="allVariables"
      :highlights="routeHighlight"
      handle-updates
    />
    <p v-if="hasResults && !isGeoView" class="predictions-data-slot-summary">
      <data-size
        :current-size="numItems"
        :total="numRows"
        @updated="$refs.size.hide()"
        @submit="onDataSizeSubmit"
      />
      <strong class="matching-color">matching</strong> samples of
      {{ numRows }} processed by model<template v-if="numIncludedRows > 0">
        , {{ numIncludedRows }}
        <strong class="selected-color">selected</strong>
      </template>
    </p>
    <p v-else-if="isGeoView" class="selection-data-slot-summary">
      Selected Area Coverage:
      <strong class="matching-color">{{ areaCoverage }}km<sup>2</sup></strong>
    </p>
    <div class="predictions-data-slot-container" :class="{ pending: !hasData }">
      <div v-if="isPending || hasNoResults" class="predictions-data-no-results">
        <div v-if="isPending" v-html="spinnerHTML" />
        <p v-if="hasNoResults">No results available</p>
      </div>

      <component
        :is="viewComponent"
        :data-fields="dataFields"
        :data-items="dataItems"
        :baseline-items="baselineItems"
        :instance-name="instanceName"
        :summaries="summaries"
        :area-of-interest-items="{ inner: inner, outer: outer }"
        :dataset="dataset"
        @tile-clicked="onTileClick"
      />
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";
import DataSize from "../components/buttons/DataSize.vue";
import PredictionsDataTable from "./PredictionsDataTable.vue";
import ImageMosaic from "./ImageMosaic.vue";
import ResultsTimeseriesView from "./ResultsTimeseriesView.vue";
import GeoPlot from "./GeoPlot.vue";
import SearchBar from "./layout/SearchBar.vue";
import { spinnerHTML } from "../util/spinner";
import {
  Highlight,
  TableRow,
  TableColumn,
  Variable,
  RowSelection,
  VariableSummary,
} from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import {
  actions as predictionsActions,
  getters as predictionsGetters,
} from "../store/predictions/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import { PredictStatus, Predictions } from "../store/requests/index";
import { Dictionary } from "../util/dict";
import { updateTableRowSelection, getNumIncludedRows } from "../util/row";
import ViewTypeToggle from "../components/ViewTypeToggle.vue";
import { MULTIBAND_IMAGE_TYPE } from "../util/types";
import LayerSelection from "./LayerSelection.vue";
import { Filter, INCLUDE_FILTER } from "../util/filters";
import { actions as viewActions } from "../store/view/module";
import {
  getPredictionResultSummary,
  getPredictionConfidenceSummary,
  getPredictionRankSummary,
  summaryToVariable,
} from "../util/summaries";
import { getVariableSummariesByState, totalAreaCoverage } from "../util/data";
import { overlayRouteEntry } from "../util/routes";
import { EI } from "../util/events";
const TABLE_VIEW = "table";
const IMAGE_VIEW = "image";
const GRAPH_VIEW = "graph";
const GEO_VIEW = "geo";
const TIMESERIES_VIEW = "timeseries";

export default Vue.extend({
  name: "PredictionsDataSlot",

  components: {
    DataSize,
    GeoPlot,
    ImageMosaic,
    PredictionsDataTable,
    ResultsTimeseriesView,
    ViewTypeToggle,
    LayerSelection,
    SearchBar,
  },

  data() {
    return {
      instanceName: "predictions",
      viewTypeModel: null,
      TABLE_VIEW: TABLE_VIEW,
      IMAGE_VIEW: IMAGE_VIEW,
      GRAPH_VIEW: GRAPH_VIEW,
      GEO_VIEW: GEO_VIEW,
      TIMESERIES_VIEW: TIMESERIES_VIEW,
    };
  },

  computed: {
    summaries(): VariableSummary[] {
      const summaryDictionary = predictionsGetters.getTrainingSummariesDictionary(
        this.$store
      );
      const currentSummaries = getVariableSummariesByState(
        1,
        this.trainingVariables.length,
        this.trainingVariables,
        summaryDictionary,
        true,
        routeGetters.getRoutePredictionsDataset(this.$store)
      );

      const rank = getPredictionRankSummary(this.prediction?.resultId);
      const confidence = getPredictionConfidenceSummary(
        this.prediction?.resultId
      );
      const summary = getPredictionResultSummary(this.prediction?.requestId);
      if (rank) {
        currentSummaries.push(rank);
      }
      if (confidence) {
        currentSummaries.push(confidence);
      }
      if (summary) {
        currentSummaries.push(summary);
      }
      return currentSummaries;
    },
    trainingVariables(): Variable[] {
      return requestGetters.getActivePredictionTrainingVariables(this.$store);
    },
    dataset(): string {
      return routeGetters.getRoutePredictionsDataset(this.$store);
    },
    allVariables(): Variable[] {
      const predictionVariables = [] as Variable[];
      const activePred = this.prediction;
      const rankSum = getPredictionRankSummary(activePred?.resultId);
      const confidenceSum = getPredictionConfidenceSummary(
        activePred?.resultId
      );
      const predSum = getPredictionResultSummary(activePred?.requestId);
      if (rankSum) {
        predictionVariables.push(summaryToVariable(rankSum));
      }
      if (confidenceSum) {
        predictionVariables.push(summaryToVariable(confidenceSum));
      }
      if (predSum) {
        predictionVariables.push(summaryToVariable(predSum));
      }
      return datasetGetters
        .getAllVariables(this.$store)
        .concat(predictionVariables);
    },
    routeHighlight(): string {
      return routeGetters.getRouteHighlight(this.$store);
    },
    highlights(): Highlight[] {
      return routeGetters.getDecodedHighlights(this.$store);
    },
    prediction(): Predictions {
      return requestGetters.getActivePredictions(this.$store);
    },
    produceRequestId(): string {
      return routeGetters.getRouteProduceRequestId(this.$store);
    },
    baselineItems(): TableRow[] {
      const result = predictionsGetters.getBaselinePredictionTableDataItems(
        this.$store
      );
      return result?.sort((a, b) => {
        return a.d3mIndex - b.d3mIndex;
      });
    },
    dataItems(): TableRow[] {
      return predictionsGetters.getIncludedPredictionTableDataItems(
        this.$store
      );
    },

    dataFields(): Dictionary<TableColumn> {
      return predictionsGetters.getIncludedPredictionTableDataFields(
        this.$store
      );
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    numItems(): number {
      return this.items ? this.items.length : 0;
    },

    numRows(): number {
      return predictionsGetters.getPredictionDataNumRows(this.$store);
    },

    viewType(): string {
      return this.viewTypeModel;
    },

    hasErrored(): boolean {
      const predictions = requestGetters.getActivePredictions(this.$store);
      return predictions
        ? predictions.progress === PredictStatus.PREDICT_ERRORED
        : false;
    },

    isPending(): boolean {
      return !this.hasData && !this.hasErrored;
    },

    hasNoResults(): boolean {
      return this.hasErrored || (this.hasData && this.items.length === 0);
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

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    spinnerHTML(): string {
      return spinnerHTML();
    },

    numIncludedRows(): number {
      return getNumIncludedRows(this.rowSelection);
    },

    isMultiBandImage(): boolean {
      const variables = datasetGetters.getVariables(this.$store);
      return variables.some((v) => {
        return v.colType === MULTIBAND_IMAGE_TYPE;
      });
    },
    inner(): TableRow[] {
      return predictionsGetters.getAreaOfInterestInnerDataItems(this.$store);
    },
    outer(): TableRow[] {
      return predictionsGetters.getAreaOfInterestOuterDataItems(this.$store);
    },
    /* Select which component to display the data. */
    viewComponent(): string {
      if (this.viewType === GEO_VIEW) return "GeoPlot";
      if (this.viewType === IMAGE_VIEW) return "ImageMosaic";
      if (this.viewType === TABLE_VIEW) return "PredictionsDataTable";
      if (this.viewType === TIMESERIES_VIEW) return "ResultsTimeseriesView";
      return "";
    },
    isGeoView(): boolean {
      return this.viewType === GEO_VIEW;
    },
    dataSize(): number {
      return routeGetters.getRouteDataSize(this.$store);
    },
    areaCoverage(): number {
      return totalAreaCoverage(this.items, this.variables);
    },
  },

  created() {
    this.viewTypeModel = TABLE_VIEW;
  },

  methods: {
    async onTileClick(data: EI.MAP.TileClickData) {
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
      await viewActions.updatePredictionAreaOfInterest(this.$store, filter);
    },
    /* When the user request to fetch a different size of data. */
    onDataSizeSubmit(dataSize: number) {
      if (this.dataSize !== dataSize) {
        const entry = overlayRouteEntry(this.$route, { dataSize });
        this.$router.push(entry).catch((err) => console.warn(err));
      }
    },
  },
  watch: {
    dataSize() {
      predictionsActions.fetchPredictionTableData(this.$store, {
        dataset: this.dataset,
        highlights: this.highlights,
        produceRequestId: this.produceRequestId,
        size: this.dataSize,
        isBaseline: false,
      });
    },
  },
});
</script>

<style scoped>
.predictions-data-slot-summary {
  flex-shrink: 0;
  font-size: 90%;
  margin: 0;
}

.predictions-data-slot {
  display: flex;
  flex-direction: column;
}

.predictions-data-slot-container {
  position: relative;
  display: flex;
  flex-flow: wrap;
  height: 100%;
  width: 100%;
}

.predictions-data-no-results {
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

.pending {
  opacity: 0.5;
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
