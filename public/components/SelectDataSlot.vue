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
  <div class="select-data-slot">
    <view-type-toggle
      v-model="viewTypeModel"
      has-tabs
      :variables="variables"
      :available-variables="trainingVariables"
      is-select-view
    >
      <b-nav-item
        class="font-weight-bold"
        :active="includedActive"
        @click="setIncludedActive(true)"
      >
        Samples to Model From
      </b-nav-item>
      <b-nav-item
        class="font-weight-bold mr-auto"
        :active="!includedActive"
        @click="setIncludedActive(false)"
      >
        Excluded Samples
      </b-nav-item>
    </view-type-toggle>

    <search-bar
      class="mb-3"
      :variables="allVariables"
      :filters="routeFilters"
      :highlights="routeHighlight"
      isSelectView
      @lex-query="updateFilterAndHighlightFromLexQuery"
    />

    <div class="table-title-container">
      <p v-if="!isGeoView" class="selection-data-slot-summary">
        <data-size
          :current-size="numItems"
          :total="numRows"
          @submit="onDataSizeSubmit"
        />
        <strong class="matching-color">matching</strong> samples of
        {{ numRows }} to model
        <template v-if="selectionNumRows > 0">
          , {{ selectionNumRows }}
          <strong class="selected-color">selected</strong>
        </template>
      </p>

      <layer-selection v-if="isMultiBandImage" class="layer-select-dropdown" />
      <b-button
        v-if="includedActive"
        class="select-data-action-exclude"
        variant="outline-secondary"
        :disabled="isExcludeDisabled"
        @click="onExcludeClick"
      >
        <i
          class="fa fa-minus-circle pr-1"
          :class="{
            'exclude-highlight': isFilteringHighlights,
            'exclude-selection': isFilteringSelection,
          }"
        />
        Exclude
      </b-button>
      <b-button
        v-if="!includedActive"
        variant="outline-secondary"
        :disabled="!isFilteringSelection"
        @click="onReincludeClick"
      >
        <i
          class="fa fa-plus-circle pr-1"
          :class="{ 'include-selection': isFilteringSelection }"
        />
        Reinclude
      </b-button>
    </div>

    <div class="select-data-container" :class="{ pending: !hasData }">
      <div v-if="!hasData" class="select-data-no-results">
        <div v-html="spinnerHTML" />
      </div>
      <component
        :is="viewComponent"
        :included-active="includedActive"
        :instance-name="instanceName"
        :dataset="dataset"
        :data-items="items"
        :data-fields="fields"
      />
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { spinnerHTML } from "../util/spinner";
import DataSize from "../components/buttons/DataSize.vue";
import SelectDataTable from "./SelectDataTable.vue";
import ImageMosaic from "./ImageMosaic.vue";
import SearchBar from "../components/layout/SearchBar.vue";
import SelectTimeseriesView from "./SelectTimeseriesView.vue";
import SelectGeoPlot from "./SelectGeoPlot.vue";
import SelectGraphView from "./SelectGraphView.vue";
import ViewTypeToggle from "./ViewTypeToggle.vue";
import LayerSelection from "./LayerSelection.vue";
import { overlayRouteEntry } from "../util/routes";
import {
  actions as datasetActions,
  getters as datasetGetters,
} from "../store/dataset/module";
import {
  TableRow,
  Variable,
  Highlight,
  RowSelection,
  TableColumn,
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import {
  Filter,
  addFilterToRoute,
  deepUpdateFiltersInRoute,
  EXCLUDE_FILTER,
  INCLUDE_FILTER,
} from "../util/filters";
import {
  clearHighlight,
  createFiltersFromHighlights,
  updateHighlight,
  UPDATE_ALL,
} from "../util/highlights";
import { lexQueryToFiltersAndHighlight } from "../util/lex";
import {
  clearRowSelection,
  getNumIncludedRows,
  getNumExcludedRows,
  createFilterFromRowSelection,
} from "../util/row";
import { actions as appActions } from "../store/app/module";
import { actions as viewActions } from "../store/view/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { Dictionary } from "lodash";

const GEO_VIEW = "geo";
const GRAPH_VIEW = "graph";
const IMAGE_VIEW = "image";
const TABLE_VIEW = "table";
const TIMESERIES_VIEW = "timeseries";

export default Vue.extend({
  name: "SelectDataSlot",

  components: {
    DataSize,
    ImageMosaic,
    LayerSelection,
    SearchBar,
    SelectDataTable,
    SelectGeoPlot,
    SelectGraphView,
    SelectTimeseriesView,
    ViewTypeToggle,
  },

  data() {
    return {
      instanceName: "select-data",
      viewTypeModel: TABLE_VIEW,
      GEO_VIEW: GEO_VIEW,
      GRAPH_VIEW: GRAPH_VIEW,
      IMAGE_VIEW: IMAGE_VIEW,
      TABLE_VIEW: TABLE_VIEW,
      TIMESERIES_VIEW: TIMESERIES_VIEW,
      includedActive: true,
    };
  },

  computed: {
    spinnerHTML(): string {
      return spinnerHTML();
    },

    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
    allVariables(): Variable[] {
      return datasetGetters.getAllVariables(this.$store);
    },
    availableVariables(): Variable[] {
      return routeGetters.getAvailableVariables(this.$store);
    },
    trainingVariables(): Variable[] {
      return routeGetters.getTrainingVariables(this.$store);
    },

    highlights(): Highlight[] {
      return routeGetters.getDecodedHighlights(this.$store);
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    numRows(): number {
      return this.hasData
        ? this.includedActive
          ? datasetGetters.getIncludedTableDataNumRows(this.$store)
          : datasetGetters.getExcludedTableDataNumRows(this.$store)
        : 0;
    },

    hasData(): boolean {
      return this.includedActive
        ? datasetGetters.hasIncludedTableData(this.$store)
        : datasetGetters.hasExcludedTableData(this.$store);
    },

    // extracts the table data from the store
    items(): TableRow[] {
      return this.hasData
        ? this.includedActive
          ? datasetGetters.getIncludedTableDataItems(this.$store)
          : datasetGetters.getExcludedTableDataItems(this.$store)
        : [];
    },

    numItems(): number {
      return this.hasData
        ? this.includedActive
          ? datasetGetters.getIncludedTableDataLength(this.$store)
          : datasetGetters.getExcludedTableDataLength(this.$store)
        : 0;
    },

    fields(): Dictionary<TableColumn> {
      return this.hasData
        ? this.includedActive
          ? datasetGetters.getIncludedTableDataFields(this.$store)
          : datasetGetters.getExcludedTableDataFields(this.$store)
        : {};
    },

    // return as filters for easier comparison in setting include/exclude state options.
    activeHighlights(): Filter[] {
      if (!this.highlights || this.highlights.length < 1) {
        return [] as Filter[];
      }
      if (this.includedActive) {
        return createFiltersFromHighlights(this.highlights, INCLUDE_FILTER);
      }
      return createFiltersFromHighlights(this.highlights, EXCLUDE_FILTER);
    },

    /* Check if the Active Filter is from an available feature. */
    areActiveHighlightsFromAnAvailableFeature(): boolean {
      if (this.activeHighlights.length < 1) {
        return false;
      }

      const activeHighlightNames = this.activeHighlights.map((v) => v.key);
      const availableVariablesNames = this.availableVariables.map((v) => v.key);
      return activeHighlightNames.reduce(
        (acc, afn) => acc || availableVariablesNames.includes(afn),
        false
      );
    },

    /* Disable the Exclude filter button. */
    isExcludeDisabled(): boolean {
      return (
        (!this.isFilteringHighlights && !this.isFilteringSelection) ||
        this.areActiveHighlightsFromAnAvailableFeature
      );
    },

    filters(): Filter[] {
      return routeGetters
        .getDecodedFilters(this.$store)
        .filter((f) => f.type !== "row");
    },

    routeFilters(): string {
      return routeGetters.getRouteFilters(this.$store);
    },

    routeHighlight(): string {
      return routeGetters.getRouteHighlight(this.$store);
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    selectionNumRows(): number {
      if (this.includedActive) {
        return getNumIncludedRows(this.rowSelection);
      } else {
        return getNumExcludedRows(this.rowSelection);
      }
    },

    isFilteringHighlights(): boolean {
      return (
        !this.isFilteringSelection &&
        this.highlights &&
        this.highlights.length > 0
      );
    },

    isFilteringSelection(): boolean {
      return !!this.rowSelection;
    },

    isMultiBandImage(): boolean {
      return routeGetters.isMultiBandImage(this.$store);
    },
    isGeoView(): boolean {
      return this.viewTypeModel === GEO_VIEW;
    },
    /* Select which component to display the data. */
    viewComponent(): string {
      if (this.viewTypeModel === GEO_VIEW) return "SelectGeoPlot";
      if (this.viewTypeModel === GRAPH_VIEW) return "SelectGraphView";
      if (this.viewTypeModel === IMAGE_VIEW) return "ImageMosaic";
      if (this.viewTypeModel === TIMESERIES_VIEW) return "SelectTimeseriesView";
      // Default is TABLE_VIEW
      return "SelectDataTable";
    },
    dataSize(): number {
      return routeGetters.getRouteDataSize(this.$store);
    },
  },

  methods: {
    onExcludeClick() {
      let filter = null;
      if (this.isFilteringHighlights) {
        filter = createFiltersFromHighlights(this.highlights, EXCLUDE_FILTER);
      } else {
        filter = createFilterFromRowSelection(
          this.rowSelection,
          EXCLUDE_FILTER
        );
      }

      addFilterToRoute(this.$router, filter);
      this.resetHighlightsOrRow();

      datasetActions.fetchVariableRankings(this.$store, {
        dataset: this.dataset,
        target: this.target,
      });

      appActions.logUserEvent(this.$store, {
        feature: Feature.FILTER_DATA,
        activity: Activity.DATA_PREPARATION,
        subActivity: SubActivity.DATA_TRANSFORMATION,
        details: { filter: filter },
      });
    },

    onReincludeClick() {
      let filter = null;
      if (this.isFilteringHighlights) {
        filter = createFiltersFromHighlights(this.highlights, INCLUDE_FILTER);
      } else {
        filter = createFilterFromRowSelection(
          this.rowSelection,
          INCLUDE_FILTER
        );
      }

      addFilterToRoute(this.$router, filter);
      this.resetHighlightsOrRow();

      datasetActions.fetchVariableRankings(this.$store, {
        dataset: this.dataset,
        target: this.target,
      });

      appActions.logUserEvent(this.$store, {
        feature: Feature.UNFILTER_DATA,
        activity: Activity.DATA_PREPARATION,
        subActivity: SubActivity.DATA_TRANSFORMATION,
        details: { filter: filter },
      });
    },

    setIncludedActive(val: boolean) {
      this.includedActive = val;
      const entry = overlayRouteEntry(this.$route, { include: `${val}` });
      this.$router.push(entry).catch((err) => console.warn(err));
    },

    /* When the user request to fetch a different size of data. */
    onDataSizeSubmit(dataSize: number) {
      if (this.dataSize !== dataSize) {
        const entry = overlayRouteEntry(this.$route, { dataSize });
        this.$router.push(entry).catch((err) => console.warn(err));
      }
    },

    resetHighlightsOrRow() {
      if (this.isFilteringHighlights) {
        clearHighlight(this.$router);
      } else {
        clearRowSelection(this.$router);
      }
    },

    updateFilterAndHighlightFromLexQuery(lexQuery) {
      const lqfh = lexQueryToFiltersAndHighlight(lexQuery, this.dataset);
      deepUpdateFiltersInRoute(this.$router, lqfh.filters);
      updateHighlight(this.$router, lqfh.highlights, UPDATE_ALL);
    },
  },
  watch: {
    dataSize() {
      viewActions.updateSelectTrainingData(this.$store);
    },
  },
});
</script>

<style scoped>
.select-data-container {
  display: flex;
  flex-flow: wrap;
  height: 100%;
  position: relative;
  width: 100%;
}

.select-data-no-results {
  position: absolute;
  display: block;
  top: 0;
  height: 100%;
  width: 100%;
  padding: 32px;
  text-align: center;
  opacity: 1;
  z-index: 1;
}

table tr {
  cursor: pointer;
}

.select-data-table .small-margin {
  margin-bottom: 0.5rem;
}

.select-data-action-exclude:not([disabled]) .include-highlight,
.select-data-action-exclude:not([disabled]) .exclude-highlight {
  color: var(--blue); /* #255dcc; */
}

.select-data-action-exclude:not([disabled]) .include-selection,
.select-data-action-exclude:not([disabled]) .exclude-selection {
  color: var(--red); /* #ff0067; */
}

.matching-color {
  color: var(--blue);
}
.selected-color {
  color: var(--red);
}

.fake-search-input {
  background-color: var(--gray-300);
  border: 1px solid var(--gray-500);
  border-radius: 0.2rem;
  display: flex;
  flex-wrap: wrap;
  min-height: 2.5rem;
  padding: 3px;
}

.pending {
  opacity: 0.5;
}

.table-title-container {
  display: flex;
  flex-direction: row;
  justify-content: flex-end;
  align-items: center;
  margin-bottom: 4px;
  margin-top: 6px;
}

.layer-select-dropdown {
  margin-right: 6px;
}

/* Make firsts element of this component unsquishable. */
.select-data-slot > *:not(:last-child) {
  flex-shrink: 0;
}

.selection-data-slot-summary {
  font-size: 90%;
  margin: auto auto -3px 0; /* Display against the table */
}

.select-data-slot .nav-link.active {
  border-top: 1px solid #ccc;
  border-right: 1px solid #ccc;
  border-bottom: 1px solid #e0e0e0;
  border-left: 1px solid #aaa;
  border-top-left-radius: 2px;
  border-top-left-radius: 0.125rem;
  border-top-right-radius: 2px;
  border-top-right-radius: 0.125rem;
  color: rgba(0, 0, 0, 1);
}

.select-data-slot .nav-item > a {
  color: rgba(0, 0, 0, 0.5);
}

.select-data-slot .nav-tabs .nav-link {
  padding: 0.5rem 0.75rem 1rem;
}
</style>
