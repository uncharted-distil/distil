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
            :dataset-name="topDataset"
            :instance-name="topFacetName"
            :rows-per-page="numRowsPerPage"
            :summaries="topVariableSummaries"
            :type-change-event="OnTypeChangeEvent"
          />
        </div>
        <div class="h-50">
          <variable-facets
            enable-search
            enable-type-change
            enable-highlighting
            :dataset-name="bottomDataset"
            :instance-name="bottomFacetName"
            :rows-per-page="numRowsPerPage"
            :summaries="bottomVariableSummaries"
            :type-change-event="OnTypeChangeEvent"
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
            :num-rows="topDatasetNumRows"
            :has-data="topDatasetHasData"
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
            :num-rows="bottomDatasetNumRows"
            :has-data="bottomDatasetHasData"
            :selected-column="bottomColumn"
            :other-selected-column="topColumn"
            instance-name="join-dataset-bottom"
            @col-clicked="onBottomColumnClicked"
          />
        </div>
        <div class="row">
          <div class="col-12">
            <join-datasets-form
              class="select-create-solutions"
              :dataset-id-a="topDataset"
              :dataset-id-b="bottomDataset"
              :dataset-a-column="topColumn"
              :dataset-b-column="bottomColumn"
              :dataset-a-fields="topDatasetFields"
              :dataset-b-fields="bottomDatasetFields"
              :join-accuracy="joinAccuracies"
              :join-absolute="joinAbsolute"
              @swap-datasets="swapDatasets"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
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
  searchVariables,
} from "../util/data";
import { TOP_VARS_INSTANCE, BOTTOM_VARS_INSTANCE } from "../store/route/index";
import { actions as viewActions } from "../store/view/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { getVariableSummaries, swapJoinView } from "../util/join";
import { EventList } from "../util/events";

export default Vue.extend({
  name: "join-datasets",

  components: {
    JoinDatasetsForm,
    JoinDataSlot,
    VariableFacets,
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
      const topVarMap = new Map(
        this.topVariables.map((tv) => {
          return [tv.key, tv];
        })
      );
      return this.variableSummaries.get(this.topDataset).filter((vs) => {
        return topVarMap.has(vs.key);
      });
    },
    bottomVariableSummaries(): VariableSummary[] {
      if (!this.variableSummaries.has(this.bottomDataset)) {
        return [];
      }
      const bottomVarMap = new Map(
        this.bottomVariables.map((bv) => {
          return [bv.key, bv];
        })
      );
      return this.variableSummaries.get(this.bottomDataset).filter((vs) => {
        return bottomVarMap.has(vs.key);
      });
    },
    variables(): Map<string, Variable[]> {
      const variables = datasetGetters.getVariables(this.$store);
      const result = new Map<string, Variable[]>();
      this.joinDatasets.forEach((jd) => {
        result.set(
          jd,
          variables.filter((v) => {
            return v.datasetName === jd;
          })
        );
      });
      return result;
    },
    topVariables(): Variable[] {
      return searchVariables(
        this.variables.get(this.topDataset),
        this.joinTopVarsSearch
      );
    },
    bottomVariables(): Variable[] {
      return searchVariables(
        this.variables.get(this.bottomDataset),
        this.joinBottomVarsSearch
      );
    },
    joinBottomVarsSearch(): string {
      return routeGetters.getRouteBottomVarsSearch(this.$store);
    },
    joinTopVarsSearch(): string {
      return routeGetters.getRouteTopVarsSearch(this.$store);
    },
    numRowsPerPage(): number {
      return NUM_PER_PAGE;
    },
    bottomFacetName(): string {
      return BOTTOM_VARS_INSTANCE;
    },
    topFacetName(): string {
      return TOP_VARS_INSTANCE;
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
    joinAccuracies(): number[] {
      const info = routeGetters.getJoinInfo(this.$store);
      if (!info) {
        return [];
      }
      return info.map((i) => {
        return i.accuracy;
      });
    },
    joinAbsolute(): boolean[] {
      const info = routeGetters.getJoinInfo(this.$store);
      if (!info) {
        return [];
      }
      return info.map((i) => {
        return i.absolute;
      });
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
    OnTypeChangeEvent(): string {
      return EventList.JOIN.JOIN_TYPE_CHANGE;
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
    // listen for type change event
    this.$eventBus.$on(this.OnTypeChangeEvent, this.onTypeChange);
  },

  beforeDestroy() {
    // Entering join view mutates variables and variable sumaries data. Clear them when exiting
    viewActions.clearJoinDatasetsData(this.$store);
    // remove type change event listener
    this.$eventBus.$off(this.OnTypeChangeEvent);
  },

  methods: {
    async onTypeChange() {
      await viewActions.updateJoinDatasetsData(this.$store);
    },
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
          this.bottomDatasetFields
        );
        if (!this.bottomColumn) {
          route.joinColumnSuggestions = suggestVars;
        }
      } else {
        if (this.bottomColumn) {
          const suggestVars = this.variableSuggestions(
            this.bottomColumn.type,
            this.topDatasetFields
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
          this.topDatasetFields
        );
        route.joinColumnB = column.key;
        if (!this.topColumn) {
          route.baseColumnSuggestions = suggestVars;
        }
      } else {
        if (this.topColumn) {
          const suggestVars = this.variableSuggestions(
            this.topColumn.type,
            this.bottomDatasetFields
          );
          route.joinColumnSuggestions = suggestVars;
        }
      }
      const entry = overlayRouteEntry(this.$route, route);
      this.$router.push(entry).catch((err) => console.warn(err));
    },
    variableSuggestions(
      type: string,
      dataFields: Dictionary<TableColumn>
    ): string[] {
      const result = [];
      for (const value in dataFields) {
        if (dataFields[value].type === type) {
          result.push(dataFields[value].key);
        }
      }
      return result;
    },
    swapDatasets() {
      if (this.joinDatasets.length >= 2) {
        swapJoinView(this.$router);
      }
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
