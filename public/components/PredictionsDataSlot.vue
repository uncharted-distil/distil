<template>
  <div class="predictions-data-slot">
    <view-type-toggle
      class="flex-shrink-0"
      v-model="viewTypeModel"
      :variables="variables"
    >
      Samples Predicted
    </view-type-toggle>

    <p class="predictions-data-slot-summary" v-if="hasResults">
      <small v-html="title"></small>
    </p>

    <div
      class="predictions-data-slot-container"
      v-bind:class="{ pending: !hasData }"
    >
      <div class="predictions-data-no-results" v-if="isPending">
        <div v-html="spinnerHTML"></div>
      </div>
      <div class="predictions-data-no-results" v-if="hasNoResults">
        No results available
      </div>

      <template>
        <predictions-data-table
          v-if="viewType === TABLE_VIEW"
          :data-fields="dataFields"
          :data-items="dataItems"
          :instance-name="instanceName"
        ></predictions-data-table>
        <results-timeseries-view
          v-if="viewType === TIMESERIES_VIEW"
          :fields="dataFields"
          :items="dataItems"
          :instance-name="instanceName"
        ></results-timeseries-view>
        <results-geo-plot
          v-if="viewType === GEO_VIEW"
          :data-fields="dataFields"
          :data-items="dataItems"
          :instance-name="instanceName"
        ></results-geo-plot>
        <image-mosaic
          v-if="viewType === IMAGE_VIEW"
          :included-active="includedActive"
          :instance-name="instanceName"
          :data-fields="dataFields"
          :data-items="dataItems"
        ></image-mosaic>
      </template>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";
import PredictionsDataTable from "./PredictionsDataTable";
import ImageMosaic from "./ImageMosaic";
import ResultsTimeseriesView from "./ResultsTimeseriesView";
import ResultsGeoPlot from "./ResultsGeoPlot";
import { spinnerHTML } from "../util/spinner";
import {
  TableRow,
  TableColumn,
  Variable,
  RowSelection
} from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as predictionsGetters } from "../store/predictions/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import {
  Solution,
  SOLUTION_ERRORED,
  PREDICT_ERRORED
} from "../store/requests/index";
import { Dictionary } from "../util/dict";
import { updateTableRowSelection, getNumIncludedRows } from "../util/row";
import ViewTypeToggle from "../components/ViewTypeToggle";

const TABLE_VIEW = "table";
const IMAGE_VIEW = "image";
const GRAPH_VIEW = "graph";
const GEO_VIEW = "geo";
const TIMESERIES_VIEW = "timeseries";

export default Vue.extend({
  name: "predictions-data-slot",

  components: {
    PredictionsDataTable,
    ResultsTimeseriesView,
    ResultsGeoPlot,
    ImageMosaic,
    ViewTypeToggle
  },

  data() {
    return {
      instanceName: "predictions",
      viewTypeModel: null,
      TABLE_VIEW: TABLE_VIEW,
      IMAGE_VIEW: IMAGE_VIEW,
      GRAPH_VIEW: GRAPH_VIEW,
      GEO_VIEW: GEO_VIEW,
      TIMESERIES_VIEW: TIMESERIES_VIEW
    };
  },

  created() {
    this.viewTypeModel = TABLE_VIEW;
  },

  computed: {
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
      return predictions ? predictions.progress === PREDICT_ERRORED : false;
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

    title(): string {
      const included = getNumIncludedRows(this.rowSelection);
      if (included > 0) {
        return `${
          this.numItems
        } <b class="matching-color">matching</b> samples of ${
          this.numRows
        } processed by model, ${included} <b class="selected-color">selected</b>`;
      } else {
        return `${
          this.numItems
        } <b class="matching-color">matching</b> samples of ${
          this.numRows
        } processed by model`;
      }
    }
  }
});
</script>

<style>
.predictions-data-slot-summary {
  margin: 10px, 0, 0, 0;
  flex-shrink: 0;
}

.predictions-data-slot {
  display: flex;
  flex-direction: column;
}

.predictions-data-slot-container {
  position: relative;
  display: flex;
  background-color: white;
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
</style>
