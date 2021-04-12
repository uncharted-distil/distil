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
      <color-scale-drop-down v-if="isMultiBandImage" />
      <layer-selection v-if="isMultiBandImage" class="layer-button" />
    </view-type-toggle>
    <search-bar
      class="mb-3"
      :variables="allVariables"
      :highlights="routeHighlight"
      @lex-query="updateFilterAndHighlightFromLexQuery"
    />
    <p v-if="hasResults" class="predictions-data-slot-summary">
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
        @tileClicked="onTileClick"
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
import GeoPlot, { TileClickData } from "./GeoPlot.vue";
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
import { PredictStatus } from "../store/requests/index";
import { Dictionary } from "../util/dict";
import { updateTableRowSelection, getNumIncludedRows } from "../util/row";
import ViewTypeToggle from "../components/ViewTypeToggle.vue";
import { MULTIBAND_IMAGE_TYPE } from "../util/types";
import ColorScaleDropDown from "./ColorScaleDropDown.vue";
import LayerSelection from "./LayerSelection.vue";
import { Filter, INCLUDE_FILTER } from "../util/filters";
import { actions as viewActions } from "../store/view/module";
import { isGeoLocatedType } from "../util/types";
import { getVariableSummariesByState } from "../util/data";
import { updateHighlight, UPDATE_ALL } from "../util/highlights";
import { lexQueryToFiltersAndHighlight } from "../util/lex";
import { resultSummariesToVariables } from "../util/summaries";
import { overlayRouteEntry } from "../util/routes";
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
    ColorScaleDropDown,
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
        summaryDictionary
      );
      return currentSummaries.filter((cs) => {
        return isGeoLocatedType(cs.varType);
      });
    },
    trainingVariables(): Variable[] {
      return requestGetters.getActivePredictionTrainingVariables(this.$store);
    },
    dataset(): string {
      return routeGetters.getRoutePredictionsDataset(this.$store);
    },
    allVariables(): Variable[] {
      let predictionVariables = [];
      const requestIds = requestGetters
        .getRelevantPredictions(this.$store)
        .map((p) => p.requestId);
      requestIds.forEach((id) => {
        predictionVariables = predictionVariables.concat(
          resultSummariesToVariables(id)
        );
      });
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
  },

  created() {
    this.viewTypeModel = TABLE_VIEW;
  },

  methods: {
    updateFilterAndHighlightFromLexQuery(lexQuery) {
      const lqfh = lexQueryToFiltersAndHighlight(lexQuery, this.dataset);
      updateHighlight(this.$router, lqfh.highlights, UPDATE_ALL);
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
      await viewActions.updatePredictionAreaOfInterest(this.$store, filter);
    },
    /* When the user request to fetch a different size of data. */
    onDataSizeSubmit(dataSize: number) {
      const entry = overlayRouteEntry(this.$route, { dataSize });
      this.$router.push(entry).catch((err) => console.warn(err));
      predictionsActions.fetchPredictionTableData(this.$store, {
        dataset: this.dataset,
        highlights: this.highlights,
        produceRequestId: this.produceRequestId,
        size: dataSize,
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
