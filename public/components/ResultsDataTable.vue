<template>
  <div class="distil-table-container">
    <b-table
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
      @row-clicked="onRowClick"
      @sort-changed="onSortChanged"
      sticky-header="100%"
      class="distil-table mb-1"
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
        <span
          >{{ data.label }}<sup>{{ solutionIndex }}</sup></span
        >
      </template>

      <template
        v-for="imageField in imageFields"
        v-slot:[cellSlot(imageField.key)]="data"
      >
        <image-preview
          :key="imageField.key"
          :type="imageField.type"
          :image-url="data.item[imageField.key].value"
        ></image-preview>
      </template>

      <template
        v-for="timeseriesGrouping in timeseriesGroupings"
        v-slot:[cellSlot(timeseriesGrouping.idCol)]="data"
      >
        <sparkline-preview
          :key="data.item[timeseriesGrouping.idCol].value"
          :truth-dataset="dataset"
          :x-col="timeseriesGrouping.xCol"
          :y-col="timeseriesGrouping.yCol"
          :timeseries-col="timeseriesGrouping.idCol"
          :timeseries-id="data.item[timeseriesGrouping.idCol].value"
          :solution-id="solutionId"
          :include-forecast="isTargetTimeseries"
        />
      </template>

      <template
        v-for="(listField, index) in listFields"
        v-slot:[cellSlot(listField.key)]="data"
      >
        <span :key="index" :title="formatList(data)">
          {{ formatList(data) }}
        </span>
      </template>

      <template v-slot:[cellSlot(errorCol)]="data">
        <!-- residual error -->
        <div>
          <div
            class="error-bar-container"
            v-if="isTargetNumerical"
            :title="data.value.value"
          >
            <div
              class="error-bar"
              v-bind:style="{
                'background-color': errorBarColor(data.value.value),
                width: errorBarWidth(data.value.value),
                left: errorBarLeft(data.value.value),
              }"
            ></div>
            <div class="error-bar-center"></div>
          </div>

          <!-- correctness error -->
          <div v-if="isTargetCategorical">
            <div v-if="data.item[predictedCol].value == data.value.value">
              Correct
            </div>
            <div v-if="data.item[predictedCol].value != data.value.value">
              Incorrect
            </div>
          </div>
        </div>
      </template>

      <template v-slot:cell()="data">
        <template v-if="['min', 'max', 'mean'].includes(data.field.key)">
          {{ data.value | cleanNumber }}
        </template>
        <div
          v-else
          :title="data.value.value"
          :style="cellColor(data.value.weight, data)"
        >
          {{ data.value.value }}
        </div>
      </template>
    </b-table>
    <b-pagination
      v-if="items && items.length > perPage"
      align="center"
      size="sm"
      v-model="currentPage"
      :per-page="perPage"
      :total-rows="itemCount"
    ></b-pagination>
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
  Grouping,
  Variable,
  RowSelection,
  TaskTypes,
  TimeseriesGrouping,
  TableValue,
} from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as resultsGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { Solution } from "../store/requests/index";
import { Dictionary } from "../util/dict";
import { getVarType, isTextType, hasComputedVarPrefix } from "../util/types";
import {
  addRowSelection,
  removeRowSelection,
  isRowSelected,
  updateTableRowSelection,
} from "../util/row";
import {
  getTimeseriesGroupingsFromFields,
  formatSlot,
  formatFieldsAsArray,
  explainCellColor,
  getImageFields,
  getListFields,
} from "../util/data";
import { getSolutionIndex } from "../util/solutions";

export default Vue.extend({
  name: "results-data-table",

  components: {
    ImagePreview,
    SparklinePreview,
    IconBase,
    IconFork,
  },

  data() {
    return {
      sortingBy: undefined,
      currentPage: 1,
      perPage: 100,
    };
  },

  props: {
    dataItems: Array as () => any[],
    dataFields: Object as () => Dictionary<TableColumn>,
    instanceName: String as () => string,
  },

  filters: {
    /* Display number with only two decimal. */
    cleanNumber(value) {
      return _.isNumber(value) ? value.toFixed(2) : "—";
    },
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
      return getVarType(this.target) === "timeseries";
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

    items(): TableRow[] {
      if (this.hasData) {
        let items = this.dataItems;

        // In the case of timeseries, we add their Min/Max/Mean.
        if (this.isTimeseries) {
          items = items?.map((item) => {
            const timeserieId = item[this.timeseriesGroupings[0].idCol].value;
            const minMaxMean = this.timeserieInfo(timeserieId);
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

      if (!this.isTimeseries || _.isEmpty(tableFields)) return tableFields;

      // For Timeseries we want to display the Min/Max/Mean
      return tableFields.concat([
        {
          key: "min",
          sortable: true,
        },
        {
          key: "max",
          sortable: true,
        },
        {
          key: "mean",
          sortable: true,
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

    imageFields(): { key: string; type: string }[] {
      return getImageFields(this.fields);
    },

    timeseriesGroupings(): TimeseriesGrouping[] {
      return getTimeseriesGroupingsFromFields(this.variables, this.fields);
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

  methods: {
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
      const listData = value.value.value.Elements as {
        Float: number;
        Status: number;
      }[];
      return listData.map((l) => l.Float);
    },
  },
  watch: {
    items() {
      // if the itemCount changes such that it's less than page
      // we were on, reset to page 1.
      if (this.itemCount < this.perPage * this.currentPage) {
        this.currentPage = 1;
      }
    },
  },
});
</script>

<style>
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
</style>
