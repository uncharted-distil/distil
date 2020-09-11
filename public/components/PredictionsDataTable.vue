<template>
  <b-table
    bordered
    hover
    small
    :items="items"
    :fields="tableFields"
    @row-clicked="onRowClick"
    @sort-changed="onSortChanged"
    sticky-header="100%"
    class="distil-table"
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
      <span>{{ data.label }}</span>
    </template>

    <template
      v-for="imageField in imageFields"
      v-slot:[cellSlot(imageField.key)]="data"
    >
      <image-preview
        :key="imageField.key"
        :type="imageField.type"
        :image-url="data.item[imageField.key].value"
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

    <template
      v-for="timeseriesGrouping in timeseriesGroupings"
      v-slot:[cellSlot(timeseriesGrouping.idCol)]="data"
    >
      <sparkline-preview
        :key="data.item[timeseriesGrouping.idCol].value"
        :truth-dataset="truthDataset"
        :forecast-dataset="predictions.dataset"
        :x-col="timeseriesGrouping.xCol"
        :y-col="timeseriesGrouping.yCol"
        :timeseries-col="timeseriesGrouping.idCol"
        :timeseries-id="data.item[timeseriesGrouping.idCol].value"
        :predictions-id="predictions.requestId"
        :include-forecast="true"
      />
    </template>
  </b-table>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";
import IconBase from "./icons/IconBase.vue";
import IconFork from "./icons/IconFork.vue";
import PredictionsDataSlot from "../components/PredictionsDataSlot.vue";
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
import { getters as predictionsGetters } from "../store/predictions/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as resultsGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { Solution, Predictions } from "../store/requests/index";
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
import { getPredictionsIndex } from "../util/predictions";

export default Vue.extend({
  name: "predictions-data-table",

  components: {
    PredictionsDataSlot,
    ImagePreview,
    SparklinePreview,
    IconBase,
    IconFork,
  },

  data() {
    return {
      sortingBy: undefined,
    };
  },

  props: {
    instanceName: String as () => string,
  },

  filters: {
    /* Display number with only two decimal. */
    cleanNumber(value) {
      return _.isNumber(value) ? value.toFixed(2) : "—";
    },
  },

  computed: {
    predictions(): Predictions {
      return requestGetters.getActivePredictions(this.$store);
    },

    predictedCol(): string {
      return this.predictions ? `${this.predictions.predictedKey}` : "";
    },

    fittedSolutionId(): string {
      return predictionsGetters.getFittedSolutionIdFromPrediction(this.$store);
    },

    truthDataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    hasData(): boolean {
      return !!predictionsGetters.getIncludedPredictionTableDataItems(
        this.$store
      );
    },

    items(): TableRow[] {
      let items = predictionsGetters.getIncludedPredictionTableDataItems(
        this.$store
      );

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
    },

    fields(): Dictionary<TableColumn> {
      return predictionsGetters.getIncludedPredictionTableDataFields(
        this.$store
      );
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
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
      const variables = datasetGetters.getVariables(this.$store);
      return getTimeseriesGroupingsFromFields(variables, this.fields);
    },

    isTimeseries(): boolean {
      return routeGetters.isTimeseries(this.$store);
    },
  },

  methods: {
    timeserieInfo(id: string): Extrema {
      const timeseries = predictionsGetters.getPredictedTimeseries(this.$store);
      return timeseries?.[this.predictions.requestId]?.info?.[id];
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
      const items = predictionsGetters.getIncludedPredictionTableDataItems(
        this.$store
      );
      return explainCellColor(weight, data, this.tableFields, items);
    },

    formatList(value: TableValue) {
      const listData = value.value.value.Elements as {
        Float: number;
        Status: number;
      }[];
      return listData.map((l) => l.Float);
    },
  },
});
</script>

<style>
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
</style>
