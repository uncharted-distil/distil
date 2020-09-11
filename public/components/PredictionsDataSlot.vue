<template>
  <div class="predictions-data-slot">
    <view-type-toggle
      class="view-toggle"
      v-model="viewTypeModel"
      :variables="variables"
    >
      <p class="font-weight-bold mr-auto">Samples Predicted</p>
      <layer-selection v-if="isRemoteSensing" class="layer-button" />
    </view-type-toggle>

    <p class="predictions-data-slot-summary" v-if="hasResults">
      <data-size
        :currentSize="numItems"
        :total="numRows"
        @updated="$refs.size.hide()"
        @submit="onDataSizeSubmit"
      />
      <strong class="matching-color">matching</strong> samples of
      {{ numRows }} processed by model<template v-if="numIncludedRows > 0"
        >, {{ numIncludedRows }}
        <strong class="selected-color">selected</strong>
      </template>
    </p>

    <div class="predictions-data-slot-container" :class="{ pending: !hasData }">
      <div class="predictions-data-no-results" v-if="isPending || hasNoResults">
        <div v-if="isPending" v-html="spinnerHTML"></div>
        <p v-if="hasNoResults">No results available</p>
      </div>

      <component
        :is="viewComponent"
        :data-fields="dataFields"
        :data-items="dataItems"
        :instance-name="instanceName"
      />
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";
import DataSize from "../components/buttons/DataSize";
import PredictionsDataTable from "./PredictionsDataTable";
import ImageMosaic from "./ImageMosaic";
import ResultsTimeseriesView from "./ResultsTimeseriesView";
import GeoPlot from "./GeoPlot";
import { spinnerHTML } from "../util/spinner";
import {
  Highlight,
  TableRow,
  TableColumn,
  Variable,
  RowSelection,
} from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import {
  actions as predictionsActions,
  getters as predictionsGetters,
} from "../store/predictions/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import {
  Solution,
  SOLUTION_ERRORED,
  PREDICT_ERRORED,
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
    DataSize,
    GeoPlot,
    ImageMosaic,
    PredictionsDataTable,
    ResultsTimeseriesView,
    ViewTypeToggle,
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

  created() {
    this.viewTypeModel = TABLE_VIEW;
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    produceRequestId(): string {
      return routeGetters.getRouteProduceRequestId(this.$store);
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

    numIncludedRows(): number {
      return getNumIncludedRows(this.rowSelection);
    },

    isRemoteSensing(): boolean {
      return routeGetters.isRemoteSensing(this.$store);
    },

    /* Select which component to display the data. */
    viewComponent() {
      if (this.viewType === GEO_VIEW) return "GeoPlot";
      if (this.viewType === IMAGE_VIEW) return "ImageMosaic";
      if (this.viewType === TABLE_VIEW) return "PredictionsDataTable";
      if (this.viewType === TIMESERIES_VIEW) return "ResultsTimeseriesView";
    },
  },

  methods: {
    /* When the user request to fetch a different size of data. */
    onDataSizeSubmit(dataSize: number) {
      predictionsActions.fetchPredictionTableData(this.$store, {
        dataset: this.dataset,
        highlight: this.highlight,
        produceRequestId: this.produceRequestId,
        size: dataSize,
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
