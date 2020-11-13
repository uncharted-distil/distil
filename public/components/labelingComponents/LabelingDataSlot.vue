<template>
  <div class="dataslot-container">
    <div class="label-headers">
      <div>
        <b-button>
          <i class="fa fa-check" aria-hidden="true"></i>
          Positive
        </b-button>
        <b-button>
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
import { getters as datasetGetters } from "../../store/dataset/module";
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
