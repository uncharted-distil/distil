<template>
  <fixed-header-table ref="fixedHeaderTable">
    <b-table
      bordered
      hover
      small
      :items="items"
      :fields="tableFields"
      @sort-changed="onSortChanged"
      @row-clicked="onRowClick"
    >
      <template
        v-for="computedField in computedFields"
        v-slot:[cellSlot(computedField)]="data"
      >
        <span :key="computedField" :title="data.value.value"
          >{{ data.value.value }}
          <icon-base icon-name="fork" class="icon-fork" width="14" height="14">
            <icon-fork /></icon-base
        ></span>
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
        <div class="container" :key="data.item[timeseriesGrouping.idCol].value">
          <div class="row">
            <sparkline-preview
              :truth-dataset="dataset"
              :x-col="timeseriesGrouping.xCol"
              :y-col="timeseriesGrouping.yCol"
              :timeseries-col="timeseriesGrouping.idCol"
              :timeseries-id="data.item[timeseriesGrouping.idCol].value"
            >
            </sparkline-preview>
          </div>
        </div>
      </template>

      <template
        v-for="listField in listFields"
        v-slot:[cellSlot(listField.key)]="data"
      >
        <span :title="formatList(data)">{{ formatList(data) }}</span>
      </template>

      <template v-slot:cell()="data">
        <span :title="data.value.value">{{ data.value.value }}</span>
      </template>
    </b-table>
  </fixed-header-table>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import IconBase from "./icons/IconBase";
import IconFork from "./icons/IconFork";
import FixedHeaderTable from "./FixedHeaderTable";
import SparklinePreview from "./SparklinePreview";
import ImagePreview from "./ImagePreview";
import { getters as datasetGetters } from "../store/dataset/module";
import { Dictionary } from "../util/dict";
import { Filter } from "../util/filters";
import {
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
    FixedHeaderTable,
    IconBase,
    IconFork,
  },

  props: {
    instanceName: String as () => string,
    includedActive: Boolean as () => boolean,
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    items(): TableRow[] {
      const items = this.includedActive
        ? datasetGetters.getIncludedTableDataItems(this.$store)
        : datasetGetters.getExcludedTableDataItems(this.$store);
      return updateTableRowSelection(
        items,
        this.rowSelection,
        this.instanceName
      );
    },

    fields(): Dictionary<TableColumn> {
      return this.includedActive
        ? datasetGetters.getIncludedTableDataFields(this.$store)
        : datasetGetters.getExcludedTableDataFields(this.$store);
    },

    tableFields(): TableColumn[] {
      return formatFieldsAsArray(this.fields);
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
  },
  updated() {
    const fixedHeaderTable = this.$refs.fixedHeaderTable as any;
    fixedHeaderTable.resizeTableCells();
  },
  methods: {
    onSortChanged() {
      // need a `nextTick` otherwise the cells get immediately overwritten
      const currentScrollLeft = this.$el.querySelector("tbody").scrollLeft;
      Vue.nextTick(() => {
        const fixedHeaderTable = this.$refs.fixedHeaderTable as any;
        fixedHeaderTable.resizeTableCells();
        fixedHeaderTable.setScrollLeft(currentScrollLeft);
      });
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
table.b-table > tfoot > tr > th.sorting:before,
table.b-table > thead > tr > th.sorting:before,
table.b-table > tfoot > tr > th.sorting:after,
table.b-table > thead > tr > th.sorting:after {
  top: 0;
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
</style>
