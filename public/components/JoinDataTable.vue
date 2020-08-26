<template>
  <fixed-header-table ref="fixedHeaderTable">
    <b-table
      bordered
      hover
      small
      :items="items"
      :fields="emphasizedFields"
      @sort-changed="onSortChanged"
      @head-clicked="onColumnClicked"
    >
      <template v-slot:cell()="data">
        {{ data.value.value }}
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
          :key="timeseriesGrouping.idCol"
          :dataset="dataset"
          :x-col="timeseriesGrouping.xCol"
          :y-col="timeseriesGrouping.yCol"
          :timeseries-col="timeseriesGrouping.idCol"
          :timeseries-id="data.item[timeseriesGrouping.idCol]"
        >
        </sparkline-preview>
      </template>
    </b-table>
  </fixed-header-table>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import FixedHeaderTable from "./FixedHeaderTable";
import SparklinePreview from "./SparklinePreview";
import ImagePreview from "./ImagePreview";
import { Dictionary } from "../util/dict";
import {
  TableColumn,
  TableRow,
  D3M_INDEX_FIELD,
  Grouping,
  Variable,
  TimeseriesGrouping,
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { TIMESERIES_TYPE, isJoinable } from "../util/types";
import {
  getTimeseriesGroupingsFromFields,
  formatFieldsAsArray,
  formatSlot,
  getImageFields,
} from "../util/data";

function findSuggestionIndex(
  columnSuggestions: string[],
  colName: string
): number {
  return columnSuggestions.findIndex((col) => {
    // col can be something like "lat+lng" for multi column suggestions
    const colNames = col.split("+");
    return Boolean(colNames.find((c) => c === colName));
  });
}

export default Vue.extend({
  name: "join-data-table",

  components: {
    ImagePreview,
    SparklinePreview,
    FixedHeaderTable,
  },

  props: {
    dataset: String as () => string,
    items: Array as () => TableRow[],
    fields: Object as () => Dictionary<TableColumn>,
    selectedColumn: Object as () => TableColumn,
    otherSelectedColumn: Object as () => TableColumn,
    instanceName: String as () => string,
  },

  computed: {
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
    isBaseJoinTable(): boolean {
      return this.instanceName === "join-dataset-top";
    },
    selectedBaseColumn(): TableColumn {
      return this.isBaseJoinTable
        ? this.selectedColumn
        : this.otherSelectedColumn;
    },
    selectedJoinColumn(): TableColumn {
      return this.isBaseJoinTable
        ? this.otherSelectedColumn
        : this.selectedColumn;
    },
    baseColumnSuggestions(): string[] {
      const columns = routeGetters
        .getBaseColumnSuggestions(this.$store)
        .split(",");
      return columns;
    },
    selectedSuggestedBaseColumn(): string {
      const index = findSuggestionIndex(
        this.baseColumnSuggestions,
        this.selectedBaseColumn.key
      );
      return index >= 0 ? this.selectedBaseColumn.key : undefined;
    },
    joinColumnSuggestions(): string[] {
      const columns = routeGetters
        .getJoinColumnSuggestions(this.$store)
        .split(",");
      return columns;
    },
    selectedSuggestedJoinColumn(): string {
      const index = findSuggestionIndex(
        this.baseColumnSuggestions,
        this.selectedJoinColumn.key
      );
      return index >= 0 ? this.selectedJoinColumn.key : undefined;
    },

    emphasizedBaseTableFields(): Dictionary<TableColumn> {
      const emphasized = {};
      _.forIn(this.fields, (field) => {
        const emph = {
          label: field.label,
          key: field.key,
          type: field.type,
          sortable: field.sortable,
          variant: null,
        };
        const isFieldSuggested =
          findSuggestionIndex(this.baseColumnSuggestions, field.key) >= 0;
        const isFieldSelected =
          this.selectedBaseColumn && field.key === this.selectedBaseColumn.key;
        if (isFieldSuggested) {
          emph.variant = "success";
        }
        if (isFieldSelected) {
          emph.variant = "primary";
        }
        emphasized[field.key] = emph;
      });
      return emphasized;
    },

    emphasizedJoinTableFields(): Dictionary<TableColumn> {
      const emphasized = {};
      _.forIn(this.fields, (field) => {
        const emph = {
          label: field.label,
          key: field.key,
          type: field.type,
          sortable: field.sortable,
          variant: null,
        };
        const isFieldSelected =
          this.selectedJoinColumn && field.key === this.selectedJoinColumn.key;
        // if a suggested base column is selected, highlgiht the corresponding suggested join column
        if (this.selectedBaseColumn) {
          const isFieldSuggested =
            findSuggestionIndex(
              this.baseColumnSuggestions,
              this.selectedBaseColumn.key
            ) === findSuggestionIndex(this.joinColumnSuggestions, field.key);
          if (
            this.selectedSuggestedBaseColumn !== undefined &&
            isFieldSuggested
          ) {
            emph.variant = "success";
          }
          if (isFieldSelected) {
            emph.variant = "primary";
          }
          if (
            isFieldSelected &&
            !isJoinable(field.type, this.selectedBaseColumn.type)
          ) {
            emph.variant = "danger";
          }
        }
        emphasized[field.key] = emph;
      });
      return emphasized;
    },

    emphasizedFields(): TableColumn[] {
      return formatFieldsAsArray(
        this.isBaseJoinTable
          ? this.emphasizedBaseTableFields
          : this.emphasizedJoinTableFields
      );
    },

    imageFields(): { key: string; type: string }[] {
      return getImageFields(this.fields);
    },

    timeseriesGroupings(): TimeseriesGrouping[] {
      return getTimeseriesGroupingsFromFields(this.variables, this.fields);
    },
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
    onColumnClicked(key, field) {
      if (this.selectedColumn && this.selectedColumn.key === key) {
        this.$emit("col-clicked", null);
      } else {
        this.$emit("col-clicked", field);
      }
    },
    cellSlot(key: string): string {
      return formatSlot(key, "cell");
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
</style>
