<template>
  <div class="dataslot-container">
    <div class="label-headers">
      <div>
        <b-button @click="onPositiveClicked">
          <i class="fa fa-check" aria-hidden="true"></i>
          Positive
        </b-button>
        <b-button @click="onNegativeClicked">
          <i class="fa fa-times" aria-hidden="true"></i>
          Negative</b-button
        >
        <layer-selection />
      </div>
      <view-type-toggle
        v-model="viewTypeModel"
        :variables="variables"
        :available-variables="variables"
      />
    </div>
    <div class="label-data-container">
      <component
        :is="viewComponent"
        :data-fields="dataFields"
        :data-items="dataItems"
        :instance-name="instanceName"
        :summaries="summaries"
        includedActive
      />
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import ViewTypeToggle from "../ViewTypeToggle.vue";
import { Dictionary } from "../../util/dict";
import GeoPlot from "../GeoPlot.vue";
import ImageMosaic from "../ImageMosaic.vue";
import SelectDataTable from "../SelectDataTable.vue";
import LayerSelection from "../LayerSelection.vue";
import {
  Variable,
  VariableSummary,
  TableRow,
  TableColumn,
} from "../../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../../store/dataset/module";
import { getters as routeGetters } from "../../store/route/module";
import { clearRowSelection } from "../../util/row";
import { LowShotLabels, LOW_SHOT_LABEL_COLUMN_NAME } from "../../util/data";
const GEO_VIEW = "geo";
const IMAGE_VIEW = "image";
const TABLE_VIEW = "table";
const TIMESERIES_VIEW = "timeseries";

export default Vue.extend({
  name: "labeling-data-slot",
  components: {
    ViewTypeToggle,
    GeoPlot,
    ImageMosaic,
    SelectDataTable,
    LayerSelection,
  },
  props: {
    variables: Array as () => Variable[],
    summaries: Array as () => VariableSummary[],
    instanceName: { type: String, default: "label" },
  },
  data() {
    return {
      viewTypeModel: TABLE_VIEW,
      eventLabel: "DataChanged",
    };
  },
  computed: {
    viewComponent(): string {
      if (this.viewTypeModel === GEO_VIEW) return "GeoPlot";
      if (this.viewTypeModel === IMAGE_VIEW) return "ImageMosaic";
      if (this.viewTypeModel === TABLE_VIEW) return "SelectDataTable";
      console.error(`viewType ${this.viewTypeModel} invalid`);
      return "";
    },
    dataItems(): TableRow[] {
      return datasetGetters.getIncludedTableDataItems(this.$store);
    },
    dataFields(): Dictionary<TableColumn> {
      return datasetGetters.getIncludedTableDataFields(this.$store);
    },
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
  },
  methods: {
    onPositiveClicked() {
      this.updateData(LowShotLabels.positive);
      this.$emit(this.eventLabel);
    },
    onNegativeClicked() {
      this.updateData(LowShotLabels.negative);
      this.$emit(this.eventLabel);
    },
    updateData(label: LowShotLabels) {
      const rowSelection = routeGetters.getDecodedRowSelection(this.$store);
      const updateData = rowSelection.d3mIndices.map((i) => {
        return {
          index: i.toString(),
          name: LOW_SHOT_LABEL_COLUMN_NAME,
          value: label,
        };
      });
      datasetActions.updateDataset(this.$store, {
        dataset: this.dataset,
        updateData,
      });
      clearRowSelection(this.$router);
    },
  },
});
</script>

<style scoped>
.label-data-container {
  display: flex;
  flex-flow: wrap;
  height: 90%;
  position: relative;
  width: 100%;
  background: #e0e0e0;
}
.dataslot-container {
  height: 90%;
}
.label-headers {
  margin: 5px;
  display: flex;
  justify-content: space-around;
}
</style>
