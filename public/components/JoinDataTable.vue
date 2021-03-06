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
  <b-table
    bordered
    hover
    small
    :items="items"
    :fields="emphasizedFields"
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
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
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

function findSuggestionIndex(columnSuggestions: string[], key: string): number {
  return columnSuggestions.findIndex((col) => {
    // col can be something like "lat+lng" for multi column suggestions
    const keys = col.split("+");
    return Boolean(keys.find((c) => c === key));
  });
}

export default Vue.extend({
  name: "join-data-table",

  components: {
    ImagePreview,
    SparklinePreview,
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
      const columnRoute = routeGetters.getBaseColumnSuggestions(this.$store);
      const columns = columnRoute ? columnRoute.split(",") : [];
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
      const columnRoute = routeGetters.getJoinColumnSuggestions(this.$store);
      const columns = columnRoute ? columnRoute.split(",") : [];
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
table tr {
  cursor: pointer;
}
</style>
