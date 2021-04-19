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
          :unique-trail="uniqueTrail"
          :should-clean-up="false"
          :should-fetch-image="false"
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
          :unique-trail="uniqueTrail"
        />
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
import PredictionsDataSlot from "../components/PredictionsDataSlot.vue";
import SparklinePreview from "./SparklinePreview.vue";
import ImagePreview from "./ImagePreview.vue";
import {
  Extrema,
  TableRow,
  TableColumn,
  D3M_INDEX_FIELD,
  RowSelection,
  TimeseriesGrouping,
  TableValue,
  Highlight,
} from "../store/dataset/index";
import {
  getters as predictionsGetters,
  actions as predictionsActions,
} from "../store/predictions/module";
import {
  getters as datasetGetters,
  actions as datasetActions,
  mutations as datasetMutations,
} from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { Predictions } from "../store/requests/index";
import { Dictionary } from "../util/dict";
import { hasComputedVarPrefix, Field } from "../util/types";
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
  removeTimeseries,
  bulkRemoveImages,
  debounceFetchImagePack,
} from "../util/data";

export default Vue.extend({
  name: "PredictionsDataTable",

  components: {
    PredictionsDataSlot,
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
    instanceName: String as () => string,
  },

  data() {
    return {
      currentPage: 1,
      perPage: 100,
      initialized: false,
      uniqueTrail: "predictions-table",
      // visibleRows contains the data being displayed on b-table
      visibleRows: [],
      debounceKey: null,
    };
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
    dataset(): string {
      return routeGetters.getRoutePredictionsDataset(this.$store);
    },
    truthDataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    hasData(): boolean {
      return !!predictionsGetters.getIncludedPredictionTableDataItems(
        this.$store
      );
    },

    pageItems(): TableRow[] {
      const end =
        this.currentPage * this.perPage > this.items.length
          ? this.items.length
          : this.currentPage * this.perPage;
      return this.items.slice((this.currentPage - 1) * this.perPage, end);
    },

    items(): TableRow[] {
      let items = predictionsGetters.getIncludedPredictionTableDataItems(
        this.$store
      );

      // In the case of timeseries, we add their Min/Max/Mean.
      if (this.isTimeseries) {
        items = items?.map((item) => {
          const timeserieId = item[this.timeseriesGroupings[0].idCol].value;
          const minMaxMean = this.timeserieInfo(timeserieId + this.uniqueTrail);
          return { ...item, ...minMaxMean };
        });
      }

      return updateTableRowSelection(
        items,
        this.rowSelection,
        this.instanceName
      );
    },

    highlights(): Highlight[] {
      return routeGetters.getDecodedHighlights(this.$store);
    },

    itemCount(): number {
      return this.hasData ? this.items.length : 0;
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

    listFields(): Field[] {
      return getListFields(this.fields);
    },

    imageFields(): Field[] {
      return getImageFields(this.fields);
    },

    timeseriesGroupings(): TimeseriesGrouping[] {
      const variables = datasetGetters.getVariables(this.$store);
      return getTimeseriesGroupingsFromFields(variables, this.fields);
    },

    isTimeseries(): boolean {
      return routeGetters.isTimeseries(this.$store);
    },
    band(): string {
      return routeGetters.getBandCombinationId(this.$store);
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
      if (!this.initialized && this.items?.length) {
        this.fetchTimeseries();
        this.initialized = true;
      }
      // if the itemCount changes such that it's less than page
      // we were on, reset to page 1.
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
      return value.value.value;
    },
    onPagination(page: number) {
      removeTimeseries(
        { predictionsId: this.predictions.requestId },
        this.pageItems,
        this.uniqueTrail
      );
      this.currentPage = page;
      this.fetchTimeseries();
      this.removeImages();
    },
    async fetchTimeseries() {
      if (!this.isTimeseries) {
        return;
      }

      this.timeseriesGroupings.forEach(async (tsg) => {
        await predictionsActions.fetchForecastedTimeseries(this.$store, {
          truthDataset: this.truthDataset,
          forecastDataset: this.predictions.dataset,
          xColName: tsg.xCol,
          yColName: tsg.yCol,
          timeseriesColName: tsg.idCol,
          predictionsId: this.predictions.requestId,
          uniqueTrail: this.uniqueTrail,
          timeseriesIds: this.pageItems.map((item) => {
            return item[tsg.idCol].value as string;
          }),
        });
      });
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

/* Highlight the predicted column */
.table td.predicted-value {
  border-right: 2px solid var(--gray-900);
}
</style>
