<template>
  <fixed-header-table ref="fixedHeaderTable">
    <b-table
      bordered
      hover
      small
      :items="items"
      :fields="tableFields"
      :sort-by="errorCol"
      :sort-compare="
        sortingByResidualError ? sortingByErrorFunction : undefined
      "
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
            <icon-fork
          /></icon-base>
        </div>
      </template>

      <template v-slot:[headSlot(predictedCol)]="data">
        <span
          >{{ data.label }}<sup>{{ solutionIndex }}</sup></span
        >
      </template>

      <template
        v-for="imageField in imageFields"
        v-slot:[cellSlot(imageField)]="data"
      >
        <image-preview
          :key="imageField"
          :image-url="data.item[imageField]"
        ></image-preview>
      </template>

      <template
        v-for="timeseriesGrouping in timeseriesGroupings"
        v-slot:[cellSlot(timeseriesGrouping.idCol)]="data"
      >
        <sparkline-preview
          :key="data.item[timeseriesGrouping.idCol]"
          :dataset="dataset"
          :x-col="timeseriesGrouping.properties.xCol"
          :y-col="timeseriesGrouping.properties.yCol"
          :timeseries-col="timeseriesGrouping.idCol"
          :timeseries-id="data.item[timeseriesGrouping.idCol]"
          :solution-id="solutionId"
          :include-forecast="isTargetTimeseries"
        >
        </sparkline-preview>
      </template>

      <template v-slot:[cellSlot(errorCol)]="data">
        <!-- residual error -->
        <div>
          <div class="error-bar-container" v-if="isTargetNumerical">
            <div
              class="error-bar"
              v-bind:style="{
                'background-color': errorBarColor(data.value.value),
                width: errorBarWidth(data.value.value),
                left: errorBarLeft(data.value.value)
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
        <div
          v-if="data.value.value.length > 0"
          :title="data.value.value"
          :style="cellColor(data.value.weight, data)"
        >
          {{ data.value.value }}
        </div>
      </template>
    </b-table>
  </fixed-header-table>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";
import IconBase from "./icons/IconBase";
import IconFork from "./icons/IconFork";
import FixedHeaderTable from "./FixedHeaderTable";
import SparklinePreview from "./SparklinePreview";
import ImagePreview from "./ImagePreview";
import {
  Extrema,
  TableRow,
  TableColumn,
  D3M_INDEX_FIELD,
  Grouping,
  Variable,
  RowSelection,
  TaskTypes
} from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as resultsGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as solutionGetters } from "../store/solutions/module";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { Solution } from "../store/solutions/index";
import { Dictionary } from "../util/dict";
import {
  getVarType,
  isTextType,
  IMAGE_TYPE,
  hasComputedVarPrefix
} from "../util/types";
import {
  addRowSelection,
  removeRowSelection,
  isRowSelected,
  updateTableRowSelection
} from "../util/row";
import {
  getTimeseriesGroupingsFromFields,
  formatSlot,
  formatFieldsAsArray
} from "../util/data";
import { getSolutionIndex } from "../util/solutions";

export default Vue.extend({
  name: "results-data-table",

  components: {
    ImagePreview,
    SparklinePreview,
    FixedHeaderTable,
    IconBase,
    IconFork
  },

  data() {
    return {
      sortingBy: undefined
    };
  },

  props: {
    dataItems: Array as () => any[],
    dataFields: Object as () => Dictionary<TableColumn>,
    instanceName: String as () => string
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    solution(): Solution {
      return solutionGetters.getActiveSolution(this.$store);
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
      return updateTableRowSelection(
        this.dataItems,
        this.rowSelection,
        this.instanceName
      );
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
      return formatFieldsAsArray(this.fields);
    },

    computedFields(): string[] {
      return Object.keys(this.fields).filter(key => {
        return hasComputedVarPrefix(key);
      });
    },

    imageFields(): string[] {
      return _.map(this.fields, (field, key) => {
        return {
          key: key,
          type: field.type
        };
      })
        .filter(field => field.type === IMAGE_TYPE)
        .map(field => field.key);
    },

    timeseriesGroupings(): Grouping[] {
      return getTimeseriesGroupingsFromFields(this.variables, this.fields);
    },

    isRegression(): boolean {
      return routeGetters.getRouteTask(this.$store) === TaskTypes.REGRESSION;
    },

    sortingByResidualError(): boolean {
      if (
        this.isRegression &&
        (this.sortingBy === this.errorCol || this.sortingBy === undefined)
      ) {
        return true;
      }
      return false;
    },
    d3mRowWeightExtrema(): Object {
      return this.dataItems.reduce((extremas, item) => {
        extremas[item[D3M_INDEX_FIELD]] = this.tableFields.reduce((rowMax, tableCol) => {
          if (item[tableCol.key].weight) {
            const currentWeight = Math.abs(item[tableCol.key].weight);
            return currentWeight > rowMax ? currentWeight : rowMax;
          } else {
            return rowMax;
          }
        }, 0);
        return extremas;
      }, {});
    },
    hasMultipleFeatures(): boolean {
      const featureNames = this.tableFields.reduce((uniqueNames, field) => {
        uniqueNames[field.label] = true;
        return uniqueNames;
      }, {});
      return Object.keys(featureNames).length > 2;
    }
  },

  updated() {
    if (this.hasResults) {
      const fixedHeaderTable = this.$refs.fixedHeaderTable as any;
      fixedHeaderTable.resizeTableCells();
    }
  },

  methods: {
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
          details: { selected: row[D3M_INDEX_FIELD] }
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
          details: { deselected: row[D3M_INDEX_FIELD] }
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

    sortingByErrorFunction(a, b, key): number {
      return Math.abs(_.toNumber(a[key])) - Math.abs(_.toNumber(b[key]));
    },

    onSortChanged(event) {
      this.sortingBy = event.sortBy;
      // need a `nextTick` otherwise the cells get immediately overwritten
      const currentScrollLeft = this.$el.querySelector("tbody").scrollLeft;
      Vue.nextTick(() => {
        const fixedHeaderTable = this.$refs.fixedHeaderTable as any;
        fixedHeaderTable.resizeTableCells();
        fixedHeaderTable.setScrollLeft(currentScrollLeft);
      });
    },

    cellSlot(key: string): string {
      return formatSlot(key, "cell");
    },

    headSlot(key: string): string {
      const hs = formatSlot(key, "head");
      return hs;
    },

    cellColor(weight: number, data: any): string {
      if (!weight || !this.hasMultipleFeatures) {
        return "";
      }
      const absoluteWeight = Math.abs(
        weight / this.d3mRowWeightExtrema[data.item[D3M_INDEX_FIELD]]
      );
      const red = 255 - 128 * absoluteWeight;
      const green = 255 - 64 * absoluteWeight;
      const blue = 255;
      return `background: rgba(${red}, ${green}, ${blue}, .75)`;
    }
  }
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
