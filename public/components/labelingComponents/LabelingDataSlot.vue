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
    <div class="fake-search-input">
      <filter-badge
        v-for="(highlight, index) in activeHighlights"
        :key="index"
        :filter="highlight"
      />
      <filter-badge
        v-for="(filter, index) in filters"
        :key="index"
        :filter="filter"
      />
    </div>
    <div class="d-flex justify-content-between m-1">
      <p class="selection-data-slot-summary">
        <strong class="matching-color">matching</strong> samples of
        {{ numRows }} to model<template v-if="selectionNumRows > 0">
          , {{ selectionNumRows }}
          <strong class="selected-color">selected</strong>
        </template>
      </p>
      <label-header-buttons
        @button-event="onAnnotationClicked"
        @select-all="onSelectAll"
      />
      <view-type-toggle
        v-model="viewTypeModel"
        :variables="variables"
        :available-variables="variables"
      />
    </div>
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
  Highlight,
} from "../../store/dataset/index";
import { getters as datasetGetters } from "../../store/dataset/module";
import { getters as routeGetters } from "../../store/route/module";
import { LowShotLabels, getAllDataItems } from "../../util/data";
import { createFiltersFromHighlights } from "../../util/highlights";
import { Filter, INCLUDE_FILTER } from "../../util/filters";
import LabelHeaderButtons from "./LabelHeaderButtons.vue";
import FilterBadge from "../FilterBadge.vue";
import { getNumIncludedRows } from "../../util/row";

const GEO_VIEW = "geo";
const IMAGE_VIEW = "image";
const TABLE_VIEW = "table";
interface DataView {
  selectAll: () => void;
}
export default Vue.extend({
  name: "labeling-data-slot",
  components: {
    ViewTypeToggle,
    LabelGeoPlot,
    ImageMosaic,
    SelectDataTable,
    LabelHeaderButtons,
    FilterBadge,
  },
  props: {
    variables: Array as () => Variable[],
    summaries: Array as () => VariableSummary[],
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
    highlights(): Highlight[] {
      return routeGetters.getDecodedHighlights(this.$store);
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
.fake-search-input {
  background-color: var(--gray-300);
  border: 1px solid var(--gray-500);
  border-radius: 0.2rem;
  display: flex;
  flex-wrap: wrap;
  min-height: 2.5rem;
  padding: 3px;
}
.selection-data-slot-summary {
  font-size: 90%;
  margin: auto 5px -3px 0; /* Display against the table */
}
</style>
