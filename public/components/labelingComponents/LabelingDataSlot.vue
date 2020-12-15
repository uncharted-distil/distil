<template>
  <div class="h-75">
    <div class="d-flex justify-content-around m-1">
      <label-header-buttons @button-event="onAnnotationClicked" />
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
import _ from "lodash";
import ViewTypeToggle from "../ViewTypeToggle.vue";
import { Dictionary } from "../../util/dict";
import LabelGeoPlot from "./LabelGeoplot.vue";
import ImageMosaic from "../ImageMosaic.vue";
import SelectDataTable from "../SelectDataTable.vue";
import {
  Variable,
  VariableSummary,
  TableRow,
  TableColumn,
  RowSelection,
} from "../../store/dataset/index";
import { getters as datasetGetters } from "../../store/dataset/module";
import { getters as routeGetters } from "../../store/route/module";
import {
  LowShotLabels,
  LOW_SHOT_LABEL_COLUMN_NAME,
  RankedSet,
  ScoreInfo,
  getAllDataItems,
} from "../../util/data";
import LabelHeaderButtons from "./LabelHeaderButtons.vue";

const GEO_VIEW = "geo";
const IMAGE_VIEW = "image";
const TABLE_VIEW = "table";

export default Vue.extend({
  name: "labeling-data-slot",
  components: {
    ViewTypeToggle,
    LabelGeoPlot,
    ImageMosaic,
    SelectDataTable,
    LabelHeaderButtons,
  },
  props: {
    variables: Array as () => Variable[],
    summaries: Array as () => VariableSummary[],
    instanceName: { type: String, default: "label" },
    rankedSet: {
      type: Object as () => RankedSet,
      default: () => {
        return { data: [] as ScoreInfo[] };
      },
    },
  },
  data() {
    return {
      viewTypeModel: TABLE_VIEW,
      eventLabel: "DataChanged",
    };
  },
  computed: {
    viewComponent(): string {
      if (this.viewTypeModel === GEO_VIEW) return "LabelGeoPlot";
      if (this.viewTypeModel === IMAGE_VIEW) return "ImageMosaic";
      if (this.viewTypeModel === TABLE_VIEW) return "SelectDataTable";
      console.error(`viewType ${this.viewTypeModel} invalid`);
      return "";
    },
    dataItems(): TableRow[] {
      const items = _.cloneDeep(
        this.viewTypeModel === GEO_VIEW
          ? getAllDataItems(true)
          : datasetGetters.getIncludedTableDataItems(this.$store)
      );
      if (this.rankedSet?.data.length) {
        const itemMap = new Map(
          items.map((d) => {
            return [d.d3mIndex, d];
          })
        );
        const unlabledItems = [];
        const labeledItems = [];
        const confidence = "confidence";
        this.rankedSet.data.forEach((item, i) => {
          const updatedItem = itemMap.get(item.d3mIndex);
          updatedItem[confidence] = {
            value: 1.0 - i / this.rankedSet.data.length,
          };
          itemMap.delete(item.d3mIndex);
          unlabledItems.push(updatedItem);
        });
        itemMap.forEach((item) => {
          item[confidence] = {
            value:
              item[LOW_SHOT_LABEL_COLUMN_NAME] === LowShotLabels.positive
                ? 1.0
                : 0.0,
          };
          labeledItems.push(item);
        });
        return unlabledItems.concat(labeledItems);
      }
      return items;
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
    negative(): string {
      return LowShotLabels.negative;
    },
    positive(): string {
      return LowShotLabels.positive;
    },
    unlabeled(): string {
      return LowShotLabels.unlabeled;
    },
  },
  methods: {
    onAnnotationClicked(label: LowShotLabels) {
      if (!this.rowSelection) {
        return;
      }
      this.$emit(this.eventLabel, label);
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
  background: #eee;
}
.label-headers {
  margin: 5px;
  display: flex;
  justify-content: space-around;
}
</style>
