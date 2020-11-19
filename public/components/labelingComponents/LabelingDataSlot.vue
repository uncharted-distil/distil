<template>
  <div class="h-75">
    <div class="d-flex justify-content-around m-1">
      <div class="pt-2">
        <b-button @click="onAnnotationClicked(positive)">
          <i class="fa fa-check text-success" aria-hidden="true"></i>
          Positive
        </b-button>
        <b-button @click="onAnnotationClicked(negative)">
          <i class="fa fa-times red" aria-hidden="true"></i>
          Negative</b-button
        >
        <layer-selection v-if="isRemoteSensing" />
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
        pagination
        includedActive
      />
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import ViewTypeToggle from "../ViewTypeToggle.vue";
import { Dictionary } from "../../util/dict";
import SelectGeoPlot from "../SelectGeoPlot.vue";
import ImageMosaic from "../ImageMosaic.vue";
import SelectDataTable from "../SelectDataTable.vue";
import LayerSelection from "../LayerSelection.vue";
import {
  Variable,
  VariableSummary,
  TableRow,
  TableColumn,
  RowSelection,
} from "../../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../../store/dataset/module";
import { getters as routeGetters } from "../../store/route/module";
import { clearRowSelection } from "../../util/row";
import { LowShotLabels, LOW_SHOT_LABEL_COLUMN_NAME } from "../../util/data";
import { MULTIBAND_IMAGE_TYPE } from "../../util/types";
const GEO_VIEW = "geo";
const IMAGE_VIEW = "image";
const TABLE_VIEW = "table";

export default Vue.extend({
  name: "labeling-data-slot",
  components: {
    ViewTypeToggle,
    SelectGeoPlot,
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
      if (this.viewTypeModel === GEO_VIEW) return "SelectGeoPlot";
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
    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },
    isRemoteSensing(): boolean {
      return this.summaries.some((s) => {
        return s.varType === MULTIBAND_IMAGE_TYPE;
      });
    },
    negative(): string {
      return LowShotLabels.negative;
    },
    positive(): string {
      return LowShotLabels.positive;
    },
  },
  methods: {
    onAnnotationClicked(label: LowShotLabels) {
      if (!this.rowSelection) {
        return;
      }
      this.updateData(label);
      this.$emit(this.eventLabel);
    },
    updateData(label: LowShotLabels) {
      const updateData = this.rowSelection.d3mIndices.map((i) => {
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
.label-headers {
  margin: 5px;
  display: flex;
  justify-content: space-around;
}
.red {
  color: var(--red);
}
</style>
