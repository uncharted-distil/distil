<template>
  <fixed-header-table ref="fixedHeaderTable">
    <b-table
      bordered
      hover
      small
      :items="items"
      :fields="tableFields"
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
        <span>{{ data.label }}</span>
      </template>

      <template
        v-for="imageField in imageFields"
        v-slot:[cellSlot(imageField)]="data"
      >
        <image-preview
          :key="imageField"
          :image-url="data.item[imageField].value"
        ></image-preview>
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
import PredictionsDataSlot from "../components/PredictionsDataSlot";
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
import { getters as predictionsGetters } from "../store/predictions/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as resultsGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { Solution, Predictions } from "../store/requests/index";
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
  formatFieldsAsArray,
  explainCellColor
} from "../util/data";
import { getSolutionIndex } from "../util/solutions";
import { getPredictionsIndex } from "../util/predictions";

export default Vue.extend({
  name: "predictions-data-table",

  components: {
    PredictionsDataSlot,
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
    instanceName: String as () => string
  },

  computed: {
    predictions(): Predictions {
      return requestGetters.getActivePredictions(this.$store);
    },

    predictedCol(): string {
      return this.predictions ? `${this.predictions.predictedKey}` : "";
    },

    isTargetTimeseries(): boolean {
      const target = this.predictions.feature;
      return getVarType(target) === "timeseries";
    },

    hasData(): boolean {
      return !!predictionsGetters.getIncludedPredictionTableDataItems(
        this.$store
      );
    },

    items(): TableRow[] {
      const items = predictionsGetters.getIncludedPredictionTableDataItems(
        this.$store
      );
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
      return formatFieldsAsArray(this.fields);
    },

    computedFields(): string[] {
      return Object.keys(this.fields).filter(key => {
        return hasComputedVarPrefix(key);
      });
    },

    imageFields(): string[] {
      return _.map(this.fields, (field, key) => ({
        key: key,
        type: field.type
      }))
        .filter(field => field.type === IMAGE_TYPE)
        .map(field => field.key);
    },

    timeseriesGroupings(): Grouping[] {
      const variables = datasetGetters.getVariables(this.$store);
      return getTimeseriesGroupingsFromFields(variables, this.fields);
    }
  },

  updated() {
    if (this.hasData && this.items.length > 0) {
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
      const items = predictionsGetters.getIncludedPredictionTableDataItems(
        this.$store
      );
      return explainCellColor(weight, data, this.tableFields, items);
    }
  }
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
