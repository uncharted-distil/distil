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
      :sort-compare="sortFunction"
      :per-page="perPage"
      :total-rows="itemCount"
      sticky-header="100%"
      class="distil-table mb-1"
      @row-clicked="onRowClick"
      @sort-changed="onSortChanged"
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
          {{ data.label }}<sup>{{ solutionIndex }}</sup>
        </span>
      </template>

      <template
        v-for="imageField in imageFields"
        v-slot:[cellSlot(imageField.key)]="data"
      >
        <image-preview
          :key="imageField.key"
          :row="data.item"
          :type="imageField.type"
          :image-url="data.item[imageField.key].value"
          :unique-trail="uniqueTrail"
          :should-clean-up="false"
          :should-fetch-image="false"
        />
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
import Vue from "vue";
import _ from "lodash";
import IconBase from "./icons/IconBase.vue";
import IconFork from "./icons/IconFork.vue";
import SparklinePreview from "./SparklinePreview.vue";
import ImagePreview from "./ImagePreview.vue";
import {
  Extrema,
  TableRow,
  TableColumn,
  D3M_INDEX_FIELD,
  Variable,
  RowSelection,
  TaskTypes,
  TimeseriesGrouping,
  TableValue,
  Highlight,
} from "../store/dataset/index";
import {
  getters as datasetGetters,
  mutations as datasetMutations,
  actions as datasetActions,
} from "../store/dataset/module";
import {
  getters as resultsGetters,
  actions as resultsActions,
} from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";

import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { Solution } from "../store/requests/index";
import { Dictionary } from "../util/dict";
import {
  getVarType,
  isTextType,
  hasComputedVarPrefix,
  TIMESERIES_TYPE,
  Field,
} from "../util/types";
import {
  addRowSelection,
  removeRowSelection,
  isRowSelected,
  updateTableRowSelection,
} from "../util/row";
import {
  formatSlot,
  formatFieldsAsArray,
  explainCellColor,
  getImageFields,
  getListFields,
  removeTimeseries,
  getTimeseriesVariablesFromFields,
  bulkRemoveImages,
  debounceFetchImagePack,
} from "../util/data";
import { getSolutionIndex } from "../util/solutions";

export default Vue.extend({
  name: "ResultsDataTable",

  components: {
    ImagePreview,
    SparklinePreview,
    IconBase,
    IconFork,
  },

  filters: {
    /* Display number with only two decimal. */
    cleanNumber(value) {
      return _.isNumber(value) ? value.toFixed(2) : "—";
    },
  },

  props: {
    dataItems: {
      type: Array as () => TableRow[],
      default: () => {
        return [];
      },
    },
    dataFields: {
      type: Object as () => Dictionary<TableColumn>,
      default: () => {
        return {};
      },
    },
    instanceName: {
      type: String as () => string,
      default: () => {
        return "";
      },
    },
  },

  data() {
    return {
      sortingBy: undefined,
      currentPage: 1,
      perPage: 100,
      uniqueTrail: "result-table",
      initialized: false,
      // visibleRows is v-model with the b-table and contains all the items in the current b-table page
      visibleRows: [],
      debounceKey: null,
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    solution(): Solution {
      return requestGetters.getActiveSolution(this.$store);
    },

    solutionId(): string {
      return routeGetters.getRouteSolutionId(this.$store);
    },

    solutionIndex(): number {
      return getSolutionIndex(this.solutionId);
    },

    hasResults(): boolean {
      return this.hasData && this.items.length > 0;
    },
    band(): string {
      return routeGetters.getBandCombinationId(this.$store);
    },
    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    isTargetCategorical(): boolean {
      return isTextType(getVarType(this.target));
    },

    isTargetNumerical(): boolean {
      return !this.isTargetCategorical;
    },

    isTargetTimeseries(): boolean {
      return getVarType(this.target) === TIMESERIES_TYPE;
    },

    predictedCol(): string {
      return this.solution ? `${this.solution.predictedKey}` : "";
    },

    errorCol(): string {
      return this.solution ? this.solution.errorKey : "";
    },

    residualExtrema(): Extrema {
      return resultsGetters.getResidualsExtrema(this.$store);
    },

    hasData(): boolean {
      return !!this.dataItems;
    },
    pageItems(): TableRow[] {
      const end =
        this.currentPage * this.perPage > this.items.length
          ? this.items.length
          : this.currentPage * this.perPage;
      return this.items.slice((this.currentPage - 1) * this.perPage, end);
    },
    items(): TableRow[] {
      if (this.hasData) {
        let items = this.dataItems;

        // In the case of timeseries, we add their Min/Max/Mean.
        if (this.isTimeseries) {
          items = items?.map((item) => {
            const timeseriesId = item[this.timeseriesVariables[0].key].value;
            const minMaxMean = this.timeserieInfo(
              timeseriesId + this.uniqueTrail
            );
            return { ...item, ...minMaxMean };
          });
        }

        return updateTableRowSelection(
          items,
          this.rowSelection,
          this.instanceName
        );
      } else {
        return [];
      }
    },

    itemCount(): number {
      return this.hasData ? this.dataItems.length : 0;
    },

    fields(): Dictionary<TableColumn> {
      return this.dataFields;
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    residualThresholdMin(): number {
      return _.toNumber(routeGetters.getRouteResidualThresholdMin(this.$store));
    },

    residualThresholdMax(): number {
      return _.toNumber(routeGetters.getRouteResidualThresholdMax(this.$store));
    },

    tableFields(): TableColumn[] {
      const tableFields = formatFieldsAsArray(this.fields);

      // Add a specific class to the predicted values
      tableFields.forEach((tf) => {
        if (this.predictedCol === tf.key) {
          tf.class = "predicted-value"; // tdClass for the TD only
        }
      });

      if (!this.isTimeseries || _.isEmpty(tableFields)) return tableFields;
      // disable sorting for timeseries tables
      tableFields.forEach((tf) => {
        tf.sortable = false;
      });
      // For Timeseries we want to display the Min/Max/Mean
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

    computedFields(): string[] {
      return Object.keys(this.fields).filter((key) => {
        return hasComputedVarPrefix(key);
      });
    },

    listFields(): { key: string; type: string }[] {
      return getListFields(this.fields);
    },

    imageFields(): Field[] {
      return getImageFields(this.fields);
    },

    timeseriesVariables(): Variable[] {
      return getTimeseriesVariablesFromFields(this.variables, this.fields);
    },

    isRegression(): boolean {
      return routeGetters.getRouteTask(this.$store) === TaskTypes.REGRESSION;
    },

    sortingByResidualError(): boolean {
      return (
        this.isRegression &&
        (this.sortingBy === this.errorCol || this.sortingBy === undefined)
      );
    },

    sortFunction(): any {
      return this.sortingByResidualError ? this.sortingByErrorFunction : null;
    },

    isTimeseries(): boolean {
      return routeGetters.isTimeseries(this.$store);
    },
  },

  watch: {
    band() {
      this.debounceImageFetch();
    },
    visibleRows() {
      this.debounceImageFetch();
    },
    highlights() {
      this.initialized = false;
    },

    items() {
      // if the itemCount changes such that it's less than page
      // we were on, reset to page 1.
      if (!this.initialized && this.items.length) {
        this.fetchTimeseries();
        this.initialized = true;
      }
      if (this.itemCount < this.perPage * this.currentPage) {
        this.currentPage = 1;
      }
    },
  },

  methods: {
    debounceImageFetch() {
      debounceFetchImagePack({
        items: this.visibleRows,
        imageFields: this.imageFields,
        dataset: this.dataset,
        uniqueTrail: this.uniqueTrail,
        debounceKey: this.debounceKey,
      });
    },

    removeImages() {
      bulkRemoveImages({
        items: this.visibleRows,
        imageFields: this.imageFields,
        uniqueTrail: this.uniqueTrail,
      });
    },

    timeserieInfo(id: string): Extrema {
      const timeseries = resultsGetters.getPredictedTimeseries(this.$store);
      return timeseries?.[this.solutionId]?.info?.[id];
    },

    onRowClick(row: TableRow) {
      if (!isRowSelected(this.rowSelection, row[D3M_INDEX_FIELD])) {
        addRowSelection(
          this.$router,
          this.instanceName,
          this.rowSelection,
          row[D3M_INDEX_FIELD]
        );

        appActions.logUserEvent(this.$store, {
          feature: Feature.CHANGE_SELECTION,
          activity: Activity.MODEL_SELECTION,
          subActivity: SubActivity.MODEL_EXPLANATION,
          details: { selected: row[D3M_INDEX_FIELD] },
        });
      } else {
        removeRowSelection(
          this.$router,
          this.instanceName,
          this.rowSelection,
          row[D3M_INDEX_FIELD]
        );

        appActions.logUserEvent(this.$store, {
          feature: Feature.CHANGE_SELECTION,
          activity: Activity.MODEL_SELECTION,
          subActivity: SubActivity.MODEL_EXPLANATION,
          details: { deselected: row[D3M_INDEX_FIELD] },
        });
      }
    },

    normalizeError(error: number): number {
      const range = this.residualExtrema.max - this.residualExtrema.min;
      return ((error - this.residualExtrema.min) / range) * 2 - 1;
    },

    // TODO: fix these to work for correctness values too

    errorBarWidth(error: number): string {
      return `${Math.abs(this.normalizeError(error) * 50)}%`;
    },

    highlights(): Highlight[] {
      return routeGetters.getDecodedHighlights(this.$store);
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

    // Sort error column by absolute values.
    sortingByErrorFunction(aRow, bRow, key): number {
      const a = Math.abs(_.toNumber(aRow[key].value));
      const b = Math.abs(_.toNumber(bRow[key].value));
      return a - b;
    },

    onSortChanged(event) {
      this.sortingBy = event.sortBy;
    },

    cellSlot(key: string): string {
      return formatSlot(key, "cell");
    },

    headSlot(key: string): string {
      const hs = formatSlot(key, "head");
      return hs;
    },

    cellColor(weight: number, data: any): string {
      return explainCellColor(weight, data, this.tableFields, this.dataItems);
    },

    formatList(value: TableValue) {
      return value.value.value;
    },
    onPagination(page: number) {
      removeTimeseries(
        { solutionId: this.solutionId },
        this.pageItems,
        this.uniqueTrail
      );
      this.currentPage = page;
      this.fetchTimeseries();
      this.removeImages();
    },

    fetchTimeseries() {
      if (!this.isTimeseries) {
        return;
      }
      this.timeseriesVariables.forEach((tsv) => {
        const tsg = tsv.grouping as TimeseriesGrouping;
        resultsActions.fetchForecastedTimeseries(this.$store, {
          dataset: this.dataset,
          variableKey: tsv.key,
          xColName: tsg.xCol,
          yColName: tsg.yCol,
          solutionId: this.solutionId,
          uniqueTrail: this.uniqueTrail,
          timeseriesIds: this.pageItems.map((item) => {
            return item[tsv.key].value as string;
          }),
        });
      });
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

.table-selected-row {
  border-left: 4px solid #ff0067;
  background-color: rgba(255, 0, 103, 0.2);
}

.table-hover tbody .table-selected-row:hover {
  border-left: 4px solid #ff0067;
  background-color: rgba(255, 0, 103, 0.4);
}

/* Highlight the predicted column */
.table td.predicted-value {
  border-right: 2px solid var(--gray-900);
}
</style>
