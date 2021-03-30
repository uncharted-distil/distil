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
      class="distil-table mb-1 noselect"
      @row-clicked="onRowClick"
    >
      <template
        v-for="computedField in computedFields"
        v-slot:[cellSlot(computedField)]="data"
      >
        <span :key="computedField" :title="data.value.value">
          {{ data.value.value }}
          <icon-base icon-name="fork" class="icon-fork" width="14" height="14">
            <icon-fork />
          </icon-base>
        </span>
      </template>

      <template
        v-for="(imageField, idx) in imageFields"
        v-slot:[cellSlot(imageField.key)]="data"
      >
        <div :key="idx" class="position-relative">
          <image-preview
            :key="imageField.key"
            :type="imageField.type"
            :row="data.item"
            :image-url="data.item[imageField.key].value"
            :unique-trail="uniqueTrail"
            :should-clean-up="false"
            :should-fetch-image="false"
          />
          <image-label
            class="image-label"
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
          <div class="row">
            <sparkline-preview
              :truth-dataset="dataset"
              :x-col="variable.grouping.xCol"
              :y-col="variable.grouping.yCol"
              :variable-key="variable.key"
              :timeseries-id="data.item[variable.key].value"
              :unique-trail="uniqueTrail"
            />
          </div>
        </div>
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
        <span v-else :title="data.value.value">{{ data.value.value }}</span>
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
      @input="onPagination"
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
import {
  getters as datasetGetters,
  actions as datasetActions,
  mutations as datasetMutations,
} from "../store/dataset/module";
import { Dictionary } from "../util/dict";
import { Filter } from "../util/filters";
import {
  Extrema,
  TableColumn,
  TableRow,
  Variable,
  D3M_INDEX_FIELD,
  RowSelection,
  TimeseriesGrouping,
  TableValue,
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { hasComputedVarPrefix, MULTIBAND_IMAGE_TYPE } from "../util/types";
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
} from "../util/data";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import ImageLabel from "./ImageLabel.vue";

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
    dataItems: { type: Array as () => TableRow[], default: null },
    includedActive: { type: Boolean, default: true },
    labelFeatureName: { type: String, default: "" },
  },

  data() {
    return {
      currentPage: 1,
      perPage: 100,
      uniqueTrail: "selected-table",
      initialized: false,
      shiftClickInfo: { first: null, second: null },
      // this is v-model with b-table (it contains what is on the page in the sorted order)
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
    items(): TableRow[] {
      let items = this.includedActive
        ? datasetGetters.getIncludedTableDataItems(this.$store)
        : datasetGetters.getExcludedTableDataItems(this.$store);
      items = this.dataItems ? this.dataItems : items;
      // In the case of timeseries, we add their Min/Max/Mean.
      if (this.isTimeseries) {
        items = items?.map((item) => {
          const timeserieId = item[this.timeseriesVariables?.[0]?.key]?.value;
          const minMaxMean = this.timeseriesInfo(
            timeserieId + this.uniqueTrail
          );
          return { ...item, ...minMaxMean };
        });
      }
      return updateTableRowSelection(
        items,
        this.rowSelection,
        this.instanceName
      );
    },
    pageItems(): TableRow[] {
      const end =
        this.currentPage * this.perPage > this.items.length
          ? this.items.length
          : this.currentPage * this.perPage;
      return this.items.slice((this.currentPage - 1) * this.perPage, end);
    },
    itemCount(): number {
      return this.includedActive
        ? datasetGetters.getIncludedTableDataLength(this.$store)
        : datasetGetters.getExcludedTableDataLength(this.$store);
    },

    fields(): Dictionary<TableColumn> {
      return this.includedActive
        ? datasetGetters.getIncludedTableDataFields(this.$store)
        : datasetGetters.getExcludedTableDataFields(this.$store);
    },

    tableFields(): TableColumn[] {
      const tableFields = formatFieldsAsArray(this.fields);

      if (!this.isTimeseries || _.isEmpty(tableFields)) return tableFields;
      // For Timeseries we want to display the Min/Max/Mean
      // disable sorting for timeseries tables
      tableFields.forEach((tf) => {
        tf.sortable = false;
      });
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

    imageFields(): { key: string; type: string }[] {
      return getImageFields(this.fields);
    },

    timeseriesGroupings(): TimeseriesGrouping[] {
      return getTimeseriesGroupingsFromFields(this.variables, this.fields);
    },

    timeseriesVariables(): Variable[] {
      return getTimeseriesVariablesFromFields(this.variables, this.fields);
    },

    computedFields(): string[] {
      const computedColumns = Object.keys(this.fields).filter((key) => {
        return hasComputedVarPrefix(key);
      });
      return computedColumns;
    },

    listFields(): { key: string; type: string }[] {
      return getListFields(this.fields);
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
  },

  watch: {
    visibleRows(prev, cur) {
      if (this.sameData(prev, cur)) {
        return;
      }
      this.debounceImageFetch();
    },
    includedActive() {
      if (this.items.length) {
        this.fetchTimeSeries();
      }
    },
    filters() {
      // new data will be coming through pipeline
      this.initialized = false;
    },
    items(cur, prev) {
      // checks to see if items exist and if the timeseries has been queried for the new data
      if (!this.initialized && this.items.length) {
        this.fetchTimeSeries();
        this.initialized = true;
      }
      if (prev?.length !== this.items.length) {
        this.fetchTimeSeries();
      }
      // if the itemCount changes such that it's less than page
      // we were on, reset to page 1.
      if (this.itemCount < this.perPage * this.currentPage) {
        this.currentPage = 1;
      }
    },
    band() {
      this.debounceImageFetch();
    },
  },
  destroyed() {
    window.removeEventListener("keyup", this.shiftRelease);
  },
  mounted() {
    window.addEventListener("keyup", this.shiftRelease);
  },
  methods: {
    sameData(old: [], cur: []): boolean {
      if (old === null || cur === null) {
        return false;
      }
      if (old.length !== cur.length) {
        return false;
      }
      for (let i = 0; i < old.length; ++i) {
        if (old[i][D3M_INDEX_FIELD] !== cur[i][D3M_INDEX_FIELD]) {
          return false;
        }
      }
      return true;
    },
    debounceImageFetch() {
      clearTimeout(this.debounceKey);
      this.debounceKey = setTimeout(() => {
        this.removeImages();
        this.fetchImagePack(this.visibleRows);
      }, 1000);
    },
    fetchTimeSeries() {
      if (!this.isTimeseries) {
        return;
      }
      this.timeseriesVariables.forEach((tsv) => {
        const grouping = tsv.grouping as TimeseriesGrouping;
        datasetActions.fetchTimeseries(this.$store, {
          dataset: this.dataset,
          variableKey: tsv.key,
          xColName: grouping.xCol,
          yColName: grouping.yCol,
          uniqueTrail: this.uniqueTrail,
          timeseriesIds: this.pageItems.map((item) => {
            return item[tsv.key].value as string;
          }),
        });
      });
    },
    removeImages() {
      if (!this.imageFields.length) {
        return;
      }
      const imageKey = this.imageFields[0].key;
      datasetMutations.bulkRemoveFiles(this.$store, {
        urls: this.visibleRows.map((item) => {
          return `${item[imageKey].value}/${this.uniqueTrail}`;
        }),
      });
    },
    fetchImagePack(items) {
      if (!this.imageFields.length) {
        return;
      }
      const key = this.imageFields[0].key;
      const type = this.imageFields[0].type;
      // if band is "" the route assumes it is an image not a multi-band image
      datasetActions.fetchImagePack(this.$store, {
        multiBandImagePackRequest: {
          imageIds: items.map((item) => {
            return item[key].value;
          }),
          dataset: this.dataset,
          band: type === MULTIBAND_IMAGE_TYPE ? this.band : "",
        },
        uniqueTrail: this.uniqueTrail,
      });
    },
    onPagination(page: number) {
      // remove old data from store
      removeTimeseries(
        { dataset: this.dataset },
        this.pageItems,
        this.uniqueTrail
      );
      this.currentPage = page;
      // fetch new data
      this.fetchTimeSeries();
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
    timeseriesInfo(id: string): Extrema {
      const timeseries = datasetGetters.getTimeseries(this.$store);
      return timeseries?.[this.dataset]?.info?.[id];
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
</style>
