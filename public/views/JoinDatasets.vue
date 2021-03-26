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
  <div class="container-fluid d-flex flex-column h-100 join-view">
    <div class="row flex-0-nav"></div>

    <div class="row align-items-center justify-content-center bg-white">
      <div class="col-12 col-md-10">
        <h5 class="header-label">
          Select Features To Join {{ topDatasetName }} with
          {{ bottomDatasetName }}
        </h5>
      </div>
    </div>

    <div class="row flex-1 pb-3 h-100">
      <div class="h-100 col-md-3">
        <div class="h-50">
          <variable-facets
            enable-search
            enable-type-change
            enable-highlighting
            :instance-name="instanceName"
            :rows-per-page="numRowsPerPage"
            :summaries="topVariableSummaries"
          />
        </div>
        <div class="h-50">
          <variable-facets
            enable-search
            enable-type-change
            enable-highlighting
            :instance-name="instanceName"
            :rows-per-page="numRowsPerPage"
            :summaries="bottomVariableSummaries"
          />
        </div>
      </div>
      <div class="col-12 col-md-9 d-flex flex-column h-100">
        <div class="row flex-1 pb-3">
          <join-data-slot
            class="col-12 pt-2 h-100"
            :dataset="topDataset"
            :items="topDatasetItems"
            :fields="topDatasetFields"
            :numRows="topDatasetNumRows"
            :hasData="topDatasetHasData"
            :selected-column="topColumn"
            :other-selected-column="bottomColumn"
            instance-name="join-dataset-top"
            @col-clicked="onTopColumnClicked"
          />
        </div>
        <div class="row flex-1 pb-3">
          <join-data-slot
            class="col-12 pt-2 h-100"
            :dataset="bottomDataset"
            :items="bottomDatasetItems"
            :fields="bottomDatasetFields"
            :numRows="bottomDatasetNumRows"
            :hasData="bottomDatasetHasData"
            :selected-column="bottomColumn"
            :other-selected-column="topColumn"
            instance-name="join-dataset-bottom"
            @col-clicked="onBottomColumnClicked"
          />
        </div>
        <div class="row pb-5">
          <div
            class="join-accuracy-slider col-12 d-flex flex-column align-items-center"
          >
            <div class="join-accuracy-label">Join Accuracy</div>
            <vue-slider
              :min="0"
              :max="1"
              :interval="0.01"
              :value="joinAccuracy"
              :lazy="true"
              width="100px"
              tooltip-dir="bottom"
              @callback="onJoinAccuracyChanged"
            />
          </div>
        </div>
        <div class="row">
          <div class="col-12">
            <join-datasets-form
              class="select-create-solutions"
              :dataset-id-a="topDataset"
              :dataset-id-b="bottomDataset"
              :dataset-a-column="topColumn"
              :dataset-b-column="bottomColumn"
              :join-accuracy="joinAccuracy"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import vueSlider from "vue-slider-component";
import JoinDatasetsForm from "../components/JoinDatasetsForm.vue";
import JoinDataSlot from "../components/JoinDataSlot.vue";
import VariableFacets from "../components/facets/VariableFacets.vue";
import { overlayRouteEntry } from "../util/routes";
import { Dictionary } from "../util/dict";
import {
  VariableSummary,
  Variable,
  TableData,
  TableColumn,
  TableRow,
  Dataset,
} from "../store/dataset/index";
import {
  NUM_PER_PAGE,
  getTableDataItems,
  getTableDataFields,
} from "../util/data";
import { JOINED_VARS_INSTANCE } from "../store/route/index";
import { actions as viewActions } from "../store/view/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { getVariableSummaries } from "../util/join";

export default Vue.extend({
  name: "join-datasets",

  components: {
    JoinDatasetsForm,
    JoinDataSlot,
    VariableFacets,
    vueSlider,
  },

  computed: {
    joinDatasets(): string[] {
      return routeGetters.getRouteJoinDatasets(this.$store);
    },
    variableSummaries(): Map<string, VariableSummary[]> {
      const result = new Map<string, VariableSummary[]>();
      this.joinDatasets.forEach((d) => {
        result.set(d, getVariableSummaries(this.$store, d));
      });
      return result;
    },
    topVariableSummaries(): VariableSummary[] {
      if (!this.variableSummaries.has(this.topDataset)) {
        return [];
      }
      return this.variableSummaries.get(this.topDataset);
    },
    bottomVariableSummaries(): VariableSummary[] {
      if (!this.variableSummaries.has(this.bottomDataset)) {
        return [];
      }
      return this.variableSummaries.get(this.bottomDataset);
    },
    numRowsPerPage(): number {
      return NUM_PER_PAGE;
    },
    instanceName(): string {
      return JOINED_VARS_INSTANCE;
    },
    highlightString(): string {
      return routeGetters.getRouteHighlight(this.$store);
    },
    joinedVarsPage(): number {
      return routeGetters.getRouteJoinDatasetsVarsPage(this.$store);
    },
    joinDatasetsTableData(): Dictionary<TableData> {
      return datasetGetters.getJoinDatasetsTableData(this.$store);
    },
    topColumn(): TableColumn {
      const colKey = routeGetters.getJoinDatasetColumnA(this.$store);
      return colKey ? this.topDatasetFields[colKey] : null;
    },
    joinAccuracy(): number {
      return routeGetters.getJoinAccuracy(this.$store);
    },
    topDataset(): string {
      return this.joinDatasets.length >= 1 ? this.joinDatasets[0] : null;
    },
    topDatasetItem(): Dataset {
      const datasets = datasetGetters.getDatasets(this.$store);
      return datasets.find((item) => item.id === this.topDataset);
    },
    topDatasetName(): string {
      return this.topDatasetItem ? this.topDatasetItem.name.toUpperCase() : "";
    },
    topDatasetTableData(): TableData {
      return this.topDataset
        ? this.joinDatasetsTableData[this.topDataset]
        : null;
    },
    topDatasetItems(): TableRow[] {
      return getTableDataItems(this.topDatasetTableData);
    },
    topDatasetFields(): Dictionary<TableColumn> {
      return getTableDataFields(this.topDatasetTableData);
    },
    topDatasetNumRows(): number {
      return this.topDatasetTableData ? this.topDatasetTableData.numRows : 0;
    },
    topDatasetHasData(): boolean {
      return !!this.topDatasetTableData;
    },
    bottomColumn(): TableColumn {
      const colKey = routeGetters.getJoinDatasetColumnB(this.$store);
      return colKey ? this.bottomDatasetFields[colKey] : null;
    },
    bottomDataset(): string {
      return this.joinDatasets.length >= 2 ? this.joinDatasets[1] : null;
    },
    bottomDatasetItem(): Dataset {
      const datasets = datasetGetters.getDatasets(this.$store);
      return datasets.find((item) => item.id === this.bottomDataset);
    },
    bottomDatasetName(): string {
      return this.bottomDatasetItem
        ? this.bottomDatasetItem.name.toUpperCase()
        : "";
    },
    bottomDatasetTableData(): TableData {
      return this.bottomDataset
        ? this.joinDatasetsTableData[this.bottomDataset]
        : null;
    },
    bottomDatasetItems(): TableRow[] {
      return getTableDataItems(this.bottomDatasetTableData);
    },
    bottomDatasetFields(): Dictionary<TableColumn> {
      return getTableDataFields(this.bottomDatasetTableData);
    },
    bottomDatasetNumRows(): number {
      return this.bottomDatasetTableData
        ? this.bottomDatasetTableData.numRows
        : 0;
    },
    bottomDatasetHasData(): boolean {
      return !!this.bottomDatasetTableData;
    },
  },

  watch: {
    highlightString() {
      viewActions.updateJoinDatasetsData(this.$store);
    },
    joinedVarsPage() {
      viewActions.updateJoinDatasetsData(this.$store);
    },
  },

  beforeMount() {
    viewActions.fetchJoinDatasetsData(this.$store);
  },

  beforeDestroy() {
    // Entering join view mutates variables and variable sumaries data. Clear them when exiting
    viewActions.clearJoinDatasetsData(this.$store);
  },

  methods: {
    onTopColumnClicked(column) {
      const route = {
        // clear top and bottom column
        joinColumnA: null,
        joinColumnSuggestions: null,
        baseColumnSuggestions: null,
      };
      if (column) {
        route.joinColumnA = column.key;
        const suggestVars = this.variableSuggestions(
          column.type,
          this.bottomVariableSummaries
        );
        if (!this.bottomColumn) {
          route.joinColumnSuggestions = suggestVars;
        }
      } else {
        if (this.bottomColumn) {
          const suggestVars = this.variableSuggestions(
            this.bottomColumn.type,
            this.topVariableSummaries
          );
          route.baseColumnSuggestions = suggestVars;
        }
      }
      const entry = overlayRouteEntry(this.$route, route);
      this.$router.push(entry).catch((err) => console.warn(err));
    },
    onBottomColumnClicked(column) {
      const route = {
        // clear top and bottom column
        joinColumnB: null,
        joinColumnSuggestions: null,
        baseColumnSuggestions: null,
      };
      let suggestVars = [];
      if (column) {
        suggestVars = this.variableSuggestions(
          column.type,
          this.topVariableSummaries
        );
        route.joinColumnB = column.key;
        if (!this.topColumn) {
          route.baseColumnSuggestions = suggestVars;
        }
      } else {
        if (this.topColumn) {
          const suggestVars = this.variableSuggestions(
            this.topColumn.type,
            this.bottomVariableSummaries
          );
          route.joinColumnSuggestions = suggestVars;
        }
      }
      const entry = overlayRouteEntry(this.$route, route);
      this.$router.push(entry).catch((err) => console.warn(err));
    },
    variableSuggestions(type: string, variables: VariableSummary[]): string[] {
      const result = [];
      variables.forEach((v) => {
        if (v.type === type) {
          result.push(v.key);
        }
      });
      return result;
    },
    onJoinAccuracyChanged(value: number) {
      const entry = overlayRouteEntry(this.$route, {
        joinAccuracy: value.toString(),
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
  },
});
</script>

<style>
.join-view .nav-link {
  padding: 1rem 0 0.25rem 0;
  border-bottom: 1px solid #e0e0e0;
  color: rgba(0, 0, 0, 0.87);
}
.header-label {
  padding: 1rem 0 0.5rem 0;
  font-weight: bold;
}
</style>
