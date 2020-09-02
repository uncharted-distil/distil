<template>
  <div class="join-data-preview-slot">
    <div class="join-data-preview-container">
      <div class="join-data-preview-no-results" v-if="!hasData">
        <div v-html="spinnerHTML"></div>
      </div>
      <template v-if="hasData">
        <join-data-preview-table
          :items="items"
          :fields="fields"
          :numRows="numRows"
          :hasData="hasData"
          :instance-name="instanceName"
        ></join-data-preview-table>
      </template>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { spinnerHTML } from "../util/spinner";
import { Dictionary } from "../util/dict";
import JoinDataPreviewTable from "./JoinDataPreviewTable";
import FilterBadge from "./FilterBadge";
import { TableRow, TableColumn } from "../store/dataset/index";

export default Vue.extend({
  name: "join-data-preview-slot",

  components: {
    FilterBadge,
    JoinDataPreviewTable,
  },

  props: {
    items: Array as () => TableRow[],
    fields: Object as () => Dictionary<TableColumn>,
    numRows: Number as () => number,
    hasData: Boolean as () => boolean,
    instanceName: String as () => string,
  },

  computed: {
    spinnerHTML(): string {
      return spinnerHTML();
    },
  },
});
</script>

<style>
.join-data-preview-container {
  display: flex;
  flex-grow: 0;
  overflow: auto;
  margin: 12px 0;
  max-height: 512px !important;
}
.join-data-preview-no-results {
  display: flex;
  background-color: #eee;
  padding: 8px;
  text-align: center;
}
</style>
