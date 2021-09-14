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
  <div class="distil-table-container">
    <b-table
      v-model="visibleRows"
      bordered
      hover
      small
      :current-page="currentPage"
      :items="items"
      :fields="tableFields"
      :sort-by="errorCol"
      :per-page="perPage"
      :total-rows="itemCount"
      sticky-header="100%"
      class="distil-table mb-1"
      @row-clicked="onRowClick"
    >
      <template
        v-for="computedField in computedFields"
        v-slot:[cellSlot(computedField)]="data"
      >
        <div :key="computedField" :title="data.value.value">
          {{ data.value.value }}
          <icon-base icon-name="fork" class="icon-fork" width="14" height="14">
            <icon-fork />
          </icon-base>
        </div>
      </template>

      <template v-slot:[headSlot(predictedCol)]="data">
        <span>
          {{ data.label }}<sup v-if="solution != null">{{ solutionIndex }}</sup>
        </span>
      </template>

      <template
        v-for="(imageField, idx) in imageFields"
        v-slot:[cellSlot(imageField.key)]="data"
      >
        <div :key="idx" class="position-relative">
          <image-preview
            :ref="`image-preview-${data.index}`"
            :key="imageField.key"
            enable-cycling
            :type="imageField.type"
            :row="data.item"
            :image-url="data.item[imageField.key].value"
            :unique-trail="uniqueTrail"
            :should-clean-up="false"
            :should-fetch-image="false"
            :index="parseInt(data.index)"
            :dataset-name="dataset"
            @cycle-images="onImageCycle"
          />
          <image-label
            class="image-label pt-1"
            included-active
            shorten-labels
            align-horizontal
            :item="data.item"
            :label-feature-name="labelFeatureName"
          />
        </div>
      </template>

      <template
        v-for="variable in timeseriesVariables"
        v-slot:[cellSlot(variable.key)]="data"
      >
        <div :key="data.item[variable.key].value" class="container">
          <sparkline-preview
            :variable-key="variable.key"
            :truth-dataset="dataset"
            :x-col="variable.grouping.xCol"
            :y-col="variable.grouping.yCol"
            :timeseries-id="data.item[variable.key].value"
            :solution-id="solutionId"
            :include-forecast="isTargetTimeseries"
            :unique-trail="uniqueTrail"
            :predictions-id="predictionId"
            :get-timeseries="getTimeseries"
          />
        </div>
      </template>

      <template
        v-for="(listField, index) in listFields"
        v-slot:[cellSlot(listField.key)]="data"
      >
        <span :key="index" :title="formatList(data)" class="min-height-20">
          {{ formatList(data) }}
        </span>
      </template>

      <template v-slot:[cellSlot(errorCol)]="data">
        <!-- residual error -->
        <div>
          <div
            v-if="isTargetNumerical"
            class="error-bar-container min-height-20"
            :title="data.value.value"
          >
            <div
              class="error-bar"
              :style="{
                'background-color': errorBarColor(data.value.value),
                width: errorBarWidth(data.value.value),
                left: errorBarLeft(data.value.value),
              }"
            />
            <div class="error-bar-center" />
          </div>

          <!-- correctness error -->
          <div v-if="isTargetCategorical">
            <div
              v-if="data.item[predictedCol].value == data.value.value"
              class="min-height-20"
            >
              Correct
            </div>
            <div
              v-if="data.item[predictedCol].value != data.value.value"
              class="min-height-20"
            >
              Incorrect
            </div>
          </div>
        </div>
      </template>

      <template v-slot:cell()="data">
        <template v-if="['min', 'max', 'mean'].includes(data.field.key)">
          <span class="min-height-20">{{ data.value | cleanNumber }}</span>
        </template>
        <div
          v-else
          :title="data.value.value"
          :style="cellColor(data.value.weight, data)"
          class="min-height-20"
        >
          {{ data.value.value }}
        </div>
      </template>
    </b-table>
    <b-pagination
      v-if="items && items.length > perPage"
      v-model="currentPage"
      align="center"
      first-number
      last-number
      size="sm"
      :per-page="perPage"
      :total-rows="itemCount"
      @change="onPagination"
    />
  </div>
</template>

<script lang="ts">
import _, { isEmpty } from "lodash";
import Vue from "vue";
import IconBase from "./icons/IconBase.vue";
import IconFork from "./icons/IconFork.vue";
import SparklinePreview from "./SparklinePreview.vue";
import ImagePreview from "./ImagePreview.vue";
import { appActions } from "../store";
import { Dictionary } from "../util/dict";
import { Filter } from "../util/filters";
import {
  TableColumn,
  TableRow,
  Variable,
  D3M_INDEX_FIELD,
  RowSelection,
  TimeseriesGrouping,
  TableValue,
  TimeSeries,
  Extrema,
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import {
  hasComputedVarPrefix,
  Field,
  getVarType,
  TIMESERIES_TYPE,
  NUMERIC_TYPE,
  CATEGORICAL_TYPE,
  isNumericType,
} from "../util/types";
import {
  addRowSelection,
  removeRowSelection,
  isRowSelected,
  updateTableRowSelection,
  bulkRowSelectionUpdate,
} from "../util/row";
import {
  getTimeseriesGroupingsFromFields,
  getTimeseriesVariablesFromFields,
  formatSlot,
  formatFieldsAsArray,
  getImageFields,
  getListFields,
  removeTimeseries,
  sameData,
  bulkRemoveImages,
  debounceFetchImagePack,
  explainCellColor,
} from "../util/data";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import ImageLabel from "./ImageLabel.vue";
import { EI, EventList } from "../util/events";
import { Solution } from "../store/requests";
import { getSolutionIndex } from "../util/solutions";

export default Vue.extend({
  name: "SelectedDataTable",

  components: {
    ImagePreview,
    SparklinePreview,
    IconBase,
    IconFork,
    ImageLabel,
  },

  filters: {
    /* Display number with only two decimal. */
    cleanNumber(value) {
      return _.isNumber(value) ? value.toFixed(2) : "—";
    },
  },

  props: {
    instanceName: { type: String as () => string, default: "" },
    dataItems: {
      type: Array as () => TableRow[],
      default: () => {
        return [];
      },
    },
    includedActive: { type: Boolean, default: true },
    labelFeatureName: { type: String, default: "" },
    dataFields: {
      type: Object as () => Dictionary<TableColumn>,
      default: {} as Dictionary<TableColumn>,
    },
    variables: {
      type: Array as () => Variable[],
      default: () => [] as Variable[],
    },
    itemCount: { type: Number as () => number, default: 0 },
    timeseriesInfo: {
      type: Object as () => TimeSeries,
      default: () => {
        return {} as TimeSeries;
      },
    },
    residualExtrema: {
      type: Object as () => Extrema,
      default: () => {
        return { min: 0, max: 0, mean: 0 };
      },
    },
    solution: { type: Object as () => Solution, default: null },
    getTimeseries: { type: Function, default: null }, // this is supplied for the sparklines
    dataset: { type: String as () => string, default: "" },
  },

  data() {
    return {
      currentPage: 1,
      perPage: 100,
      uniqueTrail: "selected-table",
      shiftClickInfo: { first: null, second: null },
      // this is v-model with b-table (it contains what is on the page in the sorted order)
      visibleRows: [],
      debounceKey: null,
    };
  },

  computed: {
    items(): TableRow[] {
      let items = this.dataItems;
      if (items === null) {
        items = [];
      }

      // In the case of timeseries, we add their Min/Max/Mean.
      if (this.isTimeseries) {
        items = items?.map((item) => {
          const timeserieId = item[this.timeseriesVariables?.[0]?.key]?.value;
          const minMaxMean = this.timeseriesInfo?.info?.[
            this.timeseriesVariables?.[0]?.key + timeserieId + this.uniqueTrail
          ];
          return { ...item, ...minMaxMean };
        });
      }
      return updateTableRowSelection(
        items,
        this.rowSelection,
        this.instanceName
      );
    },
    predictionId(): string {
      return routeGetters.getRouteProduceRequestId(this.$store);
    },
    pageItems(): TableRow[] {
      const end =
        this.currentPage * this.perPage > this.items.length
          ? this.items.length
          : this.currentPage * this.perPage;
      return this.items.slice((this.currentPage - 1) * this.perPage, end);
    },

    tableFields(): TableColumn[] {
      const tableFields = formatFieldsAsArray(this.dataFields);
      // Add a specific class to the predicted values
      tableFields.forEach((tf) => {
        if (this.predictedCol === tf.key) {
          tf.class = "predicted-value"; // tdClass for the TD only
        }
      });
      if (!this.isTimeseries || _.isEmpty(tableFields)) return tableFields;

      return tableFields.concat([
        {
          key: "min",
          sortable: false,
        },
        {
          key: "max",
          sortable: false,
        },
        {
          key: "mean",
          sortable: false,
        },
      ] as TableColumn[]);
    },
    predictedCol(): string {
      return this.solution ? `${this.solution.predictedKey}` : "";
    },
    errorCol(): string {
      return this.solution ? this.solution.errorKey : "";
    },
    imageFields(): Field[] {
      return getImageFields(this.dataFields);
    },

    timeseriesGroupings(): TimeseriesGrouping[] {
      return getTimeseriesGroupingsFromFields(this.variables, this.dataFields);
    },

    timeseriesVariables(): Variable[] {
      return getTimeseriesVariablesFromFields(this.variables, this.dataFields);
    },

    computedFields(): string[] {
      const computedColumns = Object.keys(this.dataFields).filter((key) => {
        return hasComputedVarPrefix(key);
      });
      return computedColumns;
    },

    listFields(): Field[] {
      return getListFields(this.dataFields);
    },
    solutionIndex(): number {
      return getSolutionIndex(this.solution?.solutionId);
    },
    solutionId(): string {
      return this.solution?.solutionId;
    },
    isTargetTimeseries(): boolean {
      return (
        getVarType(routeGetters.getRouteTargetVariable(this.$store)) ===
        TIMESERIES_TYPE
      );
    },
    isTargetNumerical(): boolean {
      return isNumericType(
        getVarType(routeGetters.getRouteTargetVariable(this.$store))
      );
    },
    isTargetCategorical(): boolean {
      const type = getVarType(routeGetters.getRouteTargetVariable(this.$store));
      return type === CATEGORICAL_TYPE;
    },
    filters(): Filter[] {
      return routeGetters.getDecodedFilters(this.$store);
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    isTimeseries(): boolean {
      return (
        routeGetters.isTimeseries(this.$store) ||
        !isEmpty(this.timeseriesGroupings)
      );
    },

    band(): string {
      return routeGetters.getBandCombinationId(this.$store);
    },
    residualThresholdMin(): number {
      return _.toNumber(routeGetters.getRouteResidualThresholdMin(this.$store));
    },

    residualThresholdMax(): number {
      return _.toNumber(routeGetters.getRouteResidualThresholdMax(this.$store));
    },
  },

  watch: {
    visibleRows(prev: TableRow[], cur: TableRow[]) {
      if (sameData(prev, cur)) {
        return;
      }
      this.debounceImageFetch();
      this.fetchTimeSeries();
    },

    includedActive() {
      if (this.items.length) {
        this.fetchTimeSeries();
      }
    },

    band() {
      this.debounceImageFetch();
    },

    imageFields(cur: Field[], prev: Field[]) {
      if (prev.length == 0 && cur.length > 0) {
        this.debounceImageFetch();
      }
    },
  },

  destroyed() {
    window.removeEventListener("keyup", this.shiftRelease);
  },

  mounted() {
    this.debounceImageFetch();
    this.fetchTimeSeries();
    window.addEventListener("keyup", this.shiftRelease);
  },

  methods: {
    onImageCycle(cycleInfo: EI.IMAGES.CycleImage) {
      const imagePreview = this.$refs[
        `image-preview-${cycleInfo.index + cycleInfo.side}`
      ]?.[0] as InstanceType<typeof ImagePreview>;
      imagePreview?.showZoomedImage();
    },
    debounceImageFetch() {
      debounceFetchImagePack({
        items: this.visibleRows,
        imageFields: this.imageFields,
        dataset: this.dataset,
        uniqueTrail: this.uniqueTrail,
        debounceKey: this.debounceKey,
      });
    },

    fetchTimeSeries() {
      if (!this.isTimeseries) {
        return;
      }
      this.$emit(EventList.TABLE.FETCH_TIMESERIES_EVENT, {
        variables: this.timeseriesVariables,
        uniqueTrail: this.uniqueTrail,
        timeseriesIds: this.pageItems,
      });
    },

    removeImages() {
      bulkRemoveImages({
        imageFields: this.imageFields,
        items: this.visibleRows,
        uniqueTrail: this.uniqueTrail,
      });
    },
    headSlot(key: string): string {
      const hs = formatSlot(key, "head");
      return hs;
    },
    onPagination(page: number) {
      // remove old data from store
      removeTimeseries(
        {
          dataset: this.dataset,
          predictionsId: this.predictionId,
          solutionId: this.solutionId,
        },
        this.pageItems,
        this.uniqueTrail
      );
      this.currentPage = page;
      this.removeImages();
    },

    selectAll() {
      bulkRowSelectionUpdate(
        this.$router,
        this.instanceName,
        this.rowSelection,
        this.pageItems.map((pi) => pi.d3mIndex)
      );
    },

    onRowClick(row: TableRow, idx: number, event) {
      if (event.shiftKey) {
        this.onRowShiftClick(row);
        return;
      }
      if (!isRowSelected(this.rowSelection, row[D3M_INDEX_FIELD])) {
        appActions.logUserEvent(this.$store, {
          feature: Feature.CHANGE_SELECTION,
          activity: Activity.DATA_PREPARATION,
          subActivity: SubActivity.DATA_TRANSFORMATION,
          details: { select: row[D3M_INDEX_FIELD] },
        });

        addRowSelection(
          this.$router,
          this.instanceName,
          this.rowSelection,
          row[D3M_INDEX_FIELD]
        );
      } else {
        appActions.logUserEvent(this.$store, {
          feature: Feature.CHANGE_SELECTION,
          activity: Activity.DATA_PREPARATION,
          subActivity: SubActivity.DATA_TRANSFORMATION,
          details: { deselect: row[D3M_INDEX_FIELD] },
        });

        removeRowSelection(
          this.$router,
          this.instanceName,
          this.rowSelection,
          row[D3M_INDEX_FIELD]
        );
      }
    },

    onRowShiftClick(data: TableRow) {
      if (this.shiftClickInfo.first !== null) {
        this.shiftClickInfo.second = this.items.findIndex(
          (x) => x.d3mIndex === data.d3mIndex
        );
        this.onShiftSelect();
        return;
      }
      this.shiftClickInfo.first = this.items.findIndex(
        (x) => x.d3mIndex === data.d3mIndex
      );
    },

    onShiftSelect() {
      const start = Math.min(
        this.shiftClickInfo.second,
        this.shiftClickInfo.first
      );
      const end =
        Math.max(this.shiftClickInfo.second, this.shiftClickInfo.first) + 1; // +1 deals with slicing being exclusive
      const subSet = this.items.slice(start, end).map((item) => item.d3mIndex);
      this.resetShiftClickInfo();
      bulkRowSelectionUpdate(
        this.$router,
        this.instanceName,
        this.rowSelection,
        subSet
      );
    },

    shiftRelease(event) {
      if (event.key === "Shift") {
        this.resetShiftClickInfo();
      }
    },

    resetShiftClickInfo() {
      this.shiftClickInfo.first = null;
      this.shiftClickInfo.second = null;
    },

    cellSlot(key: string): string {
      return formatSlot(key, "cell");
    },

    formatList(value: TableValue) {
      return value.value.value;
    },
    normalizeError(error: number): number {
      const range = this.residualExtrema.max - this.residualExtrema.min;
      return ((error - this.residualExtrema.min) / range) * 2 - 1;
    },
    errorBarWidth(error: number): string {
      return `${Math.abs(this.normalizeError(error) * 50)}%`;
    },

    errorBarLeft(error: number): string {
      const nerr = this.normalizeError(error);
      if (nerr > 0) {
        return "50%";
      }
      return `${50 + nerr * 50}%`;
    },

    errorBarColor(error: number): string {
      if (
        error < this.residualThresholdMin ||
        error > this.residualThresholdMax
      ) {
        return "#e05353";
      }
      return "#9e9e9e";
    },
    cellColor(weight: number, data: any): string {
      return explainCellColor(weight, data, this.tableFields, this.dataItems);
    },
  },
});
</script>

<style>
.min-height-20 {
  min-height: 20px;
}
table tr {
  cursor: pointer;
}

.table-selected-row {
  border-left: 4px solid #ff0067;
  background-color: rgba(255, 0, 103, 0.2);
}

.table-hover tbody .table-selected-row:hover {
  border-left: 4px solid #ff0067;
  background-color: rgba(255, 0, 103, 0.4);
}
.noselect {
  -webkit-touch-callout: none; /* iOS Safari */
  -webkit-user-select: none; /* Safari */
  -khtml-user-select: none; /* Konqueror HTML */
  -moz-user-select: none; /* Old versions of Firefox */
  -ms-user-select: none; /* Internet Explorer/Edge */
  user-select: none; /* Non-prefixed version, currently
                                  supported by Chrome, Edge, Opera and Firefox */
}
/*
  This keep the pagination from being squished by the table.
  The double _.distil-table-container is to increase
  specificity over the <b-pagination> component style.
*/
.distil-table-container.distil-table-container > .pagination {
  flex-shrink: 0;
}
/* Highlight the predicted column */
.table td.predicted-value {
  border-right: 2px solid var(--gray-900);
}
.table td {
  padding: 0px !important;
}
.table td > div {
  text-align: left;
  padding: 0.3rem;
  overflow: hidden;
  text-overflow: ellipsis;
  min-height: 30px;
}
.error-bar-container {
  position: relative;
  width: 80px;
  height: 18px;
}

.error-bar {
  position: absolute;
  height: 80%;
  bottom: 0;
}

.error-bar-center {
  position: absolute;
  width: 1px;
  height: 90%;
  left: 50%;
  bottom: 0;
  background-color: #666;
}
/* Highlight the predicted column */
.table td.predicted-value {
  border-right: 2px solid var(--gray-900);
}
</style>
