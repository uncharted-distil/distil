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
  <div class="join-data-slot d-flex flex-column">
    <search-bar
      class="mb-3"
      :variables="variables"
      :highlights="routeHighlight"
      :search-title="dataset"
      @lex-query="updateFilterAndHighlightFromLexQuery"
    />

    <div class="join-data-container flex-1">
      <div v-if="!hasData" class="join-data-no-results">
        <div v-html="spinnerHTML" />
      </div>
      <template v-if="hasData">
        <join-data-table
          :dataset="dataset"
          :items="items"
          :fields="fields"
          :num-rows="numRows"
          :has-data="hasData"
          :instance-name="instanceName"
          :selected-column="selectedColumn"
          :other-selected-column="otherSelectedColumn"
          @col-clicked="onColumnClicked"
        />
      </template>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { spinnerHTML } from "../util/spinner";
import { Dictionary } from "../util/dict";
import { getters as routeGetters } from "../store/route/module";
import {
  TableRow,
  TableColumn,
  Highlight,
  Variable,
} from "../store/dataset/index";
import { createFiltersFromHighlights, UPDATE_ALL } from "../util/highlights";
import { Filter, INCLUDE_FILTER } from "../util/filters";
import { updateHighlight } from "../util/highlights";
import { lexQueryToFiltersAndHighlight } from "../util/lex";
// components
import JoinDataTable from "./JoinDataTable.vue";
import SearchBar from "./layout/SearchBar.vue";
import { EventList } from "../util/events";
export default Vue.extend({
  name: "join-data-slot",

  components: {
    JoinDataTable,
    SearchBar,
  },

  props: {
    dataset: String as () => string,
    items: Array as () => TableRow[],
    fields: Object as () => Dictionary<TableColumn>,
    numRows: Number as () => number,
    hasData: Boolean as () => boolean,
    selectedColumn: Object as () => TableColumn,
    otherSelectedColumn: Object as () => TableColumn,
    instanceName: String as () => string,
  },

  computed: {
    spinnerHTML(): string {
      return spinnerHTML();
    },

    highlights(): Highlight[] {
      return routeGetters.getDecodedHighlights(this.$store);
    },

    activeHighlights(): Filter[] {
      if (
        (this.highlights && this.highlights.length < 1) ||
        this.highlights.reduce(
          (acc, highlight) => acc || highlight.dataset !== this.dataset,
          false
        )
      ) {
        return [];
      }
      return createFiltersFromHighlights(this.highlights, INCLUDE_FILTER);
    },
    variables(): Variable[] {
      return routeGetters.getJoinDatasetsVariables(this.$store).filter((v) => {
        return v.datasetName === this.dataset;
      });
    },
    routeHighlight(): string {
      return routeGetters.getRouteHighlight(this.$store);
    },
  },

  methods: {
    onColumnClicked(field) {
      this.$emit(EventList.TABLE.COLUMN_CLICKED_EVENT, field);
    },
    updateFilterAndHighlightFromLexQuery(lexQuery) {
      const lqfh = lexQueryToFiltersAndHighlight(
        lexQuery,
        this.dataset,
        this.variables
      );
      const highlights = routeGetters.getDecodedHighlights(this.$store);
      lqfh.highlights = lqfh.highlights.concat(
        highlights.filter((highlight) => {
          return highlight.dataset !== this.dataset;
        })
      );
      updateHighlight(this.$router, lqfh.highlights, UPDATE_ALL);
    },
  },
});
</script>

<style scoped>
.join-data-container {
  display: flex;
  background-color: white;
  overflow: hidden;
  flex-flow: wrap;
  height: 100%;
  width: 100%;
}
.join-data-no-results {
  width: 100%;
  background-color: #eee;
  padding: 8px;
  text-align: center;
}
.fake-search-input {
  position: relative;
  height: 38px;
  padding: 2px 2px;
  margin-bottom: 4px;
  background-color: #eee;
  border: 1px solid #ccc;
  border-radius: 0.2rem;
}
</style>
