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
  <div class="flex-1 d-flex flex-column pb-1 pt-2">
    <search-bar
      class="mb-3"
      :variables="allVariables"
      :highlights="routeHighlights"
      handle-updates
    />
    <b-container class="m-0 p-0 mw-100">
      <b-row class="d-flex justify-content-between m-0 w-100">
        <label-header-buttons
          class="height-36"
          @button-event="onAnnotationClicked"
          @select-all="onSelectAll"
        />
        <div class="d-flex">
          <layer-selection />
          <view-type-toggle
            v-model="viewTypeModel"
            class="m-0 p-0 pl-2 height-36"
            :variables="variables"
            :available-variables="variables"
          />
        </div>
      </b-row>
      <b-row class="m-0 mt-2 mb-1 w-100">
        <p v-if="!isGeoView" class="selection-data-slot-summary">
          <strong class="matching-color">matching</strong> samples of
          {{ numRows }} to model<template v-if="selectionNumRows > 0">
            , {{ selectionNumRows }}
            <strong class="selected-color">selected</strong>
          </template>
        </p>
        <p v-else class="selection-data-slot-summary">
          Selected Area Coverage:
          <strong class="matching-color"
            >{{ areaCoverage }}km<sup>2</sup></strong
          >
          <template v-if="selectionNumRows > 0">
            , {{ selectionNumRows }}
            <strong class="selected-color">selected</strong>
          </template>
        </p>
      </b-row>
    </b-container>
    <div class="label-data-container">
      <component
        :is="viewComponent"
        ref="dataView"
        :data-fields="dataFields"
        :data-items="dataItems"
        :instance-name="instanceName"
        :summaries="summaries"
        :has-confidence="hasConfidence"
        :label-feature-name="labelFeatureName"
        :label-score-name="labelScoreName"
        :dataset="dataset"
        pagination
        included-active
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
import SearchBar from "../layout/SearchBar.vue";
import ImageMosaic from "../ImageMosaic.vue";
import LayerSelection from "../LayerSelection.vue";
import SelectDataTable from "../SelectDataTable.vue";
import {
  Variable,
  VariableSummary,
  TableRow,
  TableColumn,
  RowSelection,
  Highlight,
} from "../../store/dataset/index";
import { getters as datasetGetters } from "../../store/dataset/module";
import { getters as routeGetters } from "../../store/route/module";
import { LowShotLabels, totalAreaCoverage } from "../../util/data";
import { createFiltersFromHighlights } from "../../util/highlights";
import { Filter, INCLUDE_FILTER } from "../../util/filters";
import LabelHeaderButtons from "./LabelHeaderButtons.vue";
import { getNumIncludedRows } from "../../util/row";

const GEO_VIEW = "geo";
const IMAGE_VIEW = "image";
const TABLE_VIEW = "table";
interface DataView {
  selectAll: () => void;
}
export default Vue.extend({
  name: "LabelingDataSlot",
  components: {
    ViewTypeToggle,
    LabelGeoPlot,
    ImageMosaic,
    SelectDataTable,
    LabelHeaderButtons,
    SearchBar,
    LayerSelection,
  },
  props: {
    variables: {
      type: Array as () => Variable[],
      default: () => {
        return [] as Variable[];
      },
    },
    summaries: {
      type: Array as () => VariableSummary[],
      default: () => {
        return [] as Variable[];
      },
    },
    instanceName: { type: String, default: "label" },
    hasConfidence: { type: Boolean as () => boolean, default: false },
    labelFeatureName: { type: String, default: "" },
    labelScoreName: { type: String, default: "" },
  },
  data() {
    return {
      viewTypeModel: TABLE_VIEW,
      eventLabel: "DataChanged",
    };
  },
  computed: {
    activeHighlights(): Filter[] {
      if (!this.highlights || this.highlights.length < 1) {
        return [];
      }
      return createFiltersFromHighlights(this.highlights, INCLUDE_FILTER);
    },
    allVariables(): Variable[] {
      return datasetGetters.getAllVariables(this.$store);
    },
    highlights(): Highlight[] {
      return routeGetters.getDecodedHighlights(this.$store);
    },
    routeHighlights(): string {
      return routeGetters.getRouteHighlight(this.$store);
    },
    filters(): Filter[] {
      return routeGetters
        .getDecodedFilters(this.$store)
        .filter((f) => f.type !== "row");
    },
    numRows(): number {
      return datasetGetters.getIncludedTableDataNumRows(this.$store);
    },
    selectionNumRows(): number {
      return getNumIncludedRows(this.rowSelection);
    },
    isGeoView(): boolean {
      return this.viewTypeModel === GEO_VIEW;
    },
    viewComponent(): string {
      if (this.viewTypeModel === GEO_VIEW) return "LabelGeoPlot";
      if (this.viewTypeModel === IMAGE_VIEW) return "ImageMosaic";
      if (this.viewTypeModel === TABLE_VIEW) return "SelectDataTable";
      console.error(`viewType ${this.viewTypeModel} invalid`);
      return "";
    },
    hasLowShotScores(): boolean {
      const orderBy = routeGetters.getOrderBy(this.$store);
      return !orderBy ? false : orderBy.includes(this.labelScoreName);
    },
    dataItems(): TableRow[] {
      return datasetGetters.getIncludedTableDataItems(this.$store);
    },
    numItems(): number {
      return this.dataItems?.length;
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
    areaCoverage(): number {
      return totalAreaCoverage(this.dataItems, this.variables);
    },
  },
  methods: {
    onAnnotationClicked(label: LowShotLabels) {
      if (!this.rowSelection) {
        return;
      }
      this.$emit(this.eventLabel, label);
    },
    onSelectAll() {
      const dataView = (this.$refs.dataView as unknown) as DataView;
      dataView.selectAll();
    },
  },
});
</script>

<style scoped>
.label-data-container {
  display: flex;
  flex-flow: wrap;
  height: 80%;
  position: relative;
  width: 100%;
  background: #eee;
}
.label-headers {
  margin: 5px;
  display: flex;
  justify-content: space-around;
}
.height-36 {
  height: 36px;
}
.selection-data-slot-summary {
  font-size: 90%;
  margin: auto 5px -3px 0; /* Display against the table */
}
</style>
