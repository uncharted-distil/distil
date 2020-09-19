<template>
  <div class="distil-table-container">
    <b-table
      bordered
      hover
      small
      :items="items"
      :current-page="currentPage"
      :fields="tableFields"
      :per-page="perPage"
      :total-rows="itemCount"
      @row-clicked="onRowClick"
      sticky-header="100%"
      class="distil-table"
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
        v-for="timeseriesGrouping in timeseriesGroupings"
        v-slot:[cellSlot(timeseriesGrouping.idCol)]="data"
      >
        <div class="container" :key="data.item[timeseriesGrouping.idCol].value">
          <div class="row">
            <sparkline-preview
              :truth-dataset="dataset"
              :x-col="timeseriesGrouping.xCol"
              :y-col="timeseriesGrouping.yCol"
              :timeseries-col="timeseriesGrouping.idCol"
              :timeseries-id="data.item[timeseriesGrouping.idCol].value"
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
      size="sm"
      v-model="currentPage"
      :per-page="perPage"
      :total-rows="itemCount"
    ></b-pagination>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import IconBase from "./icons/IconBase.vue";
import IconFork from "./icons/IconFork.vue";
import SparklinePreview from "./SparklinePreview.vue";
import ImagePreview from "./ImagePreview.vue";
import { getters as datasetGetters } from "../store/dataset/module";
import { Dictionary } from "../util/dict";
import { Filter } from "../util/filters";
import {
  Extrema,
  TableColumn,
  TableRow,
  Grouping,
  Variable,
  D3M_INDEX_FIELD,
  RowSelection,
  TimeseriesGrouping,
  TableData,
  TableValue,
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { TIMESERIES_TYPE, hasComputedVarPrefix } from "../util/types";
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
  getImageFields,
  getListFields,
} from "../util/data";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";

export default Vue.extend({
  name: "selected-data-table",

  components: {
    ImagePreview,
    SparklinePreview,
    IconBase,
    IconFork,
  },

  props: {
    instanceName: String as () => string,
    includedActive: Boolean as () => boolean,
  },

  data() {
    return {
      currentPage: 1,
      perPage: 100,
    };
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

    items(): TableRow[] {
      let items = this.includedActive
        ? datasetGetters.getIncludedTableDataItems(this.$store)
        : datasetGetters.getExcludedTableDataItems(this.$store);

      // In the case of timeseries, we add their Min/Max/Mean.
      if (this.isTimeseries) {
        items = items?.map((item) => {
          const timeserieId = item[this.timeseriesGroupings[0].idCol].value;
          const minMaxMean = this.timeseriesInfo(timeserieId);
          return { ...item, ...minMaxMean };
        });
      }
      return updateTableRowSelection(
        items,
        this.rowSelection,
        this.instanceName
      );
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

    imageFields(): { key: string; type: string }[] {
      return getImageFields(this.fields);
    },

    timeseriesGroupings(): TimeseriesGrouping[] {
      return getTimeseriesGroupingsFromFields(this.variables, this.fields);
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
      return routeGetters.isTimeseries(this.$store);
    },
  },

  methods: {
    timeseriesInfo(id: string): Extrema {
      const timeseries = datasetGetters.getTimeseries(this.$store);
      return timeseries?.[this.dataset]?.info?.[id];
    },

    onRowClick(row: TableRow) {
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

    cellSlot(key: string): string {
      return formatSlot(key, "cell");
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
