<template>
  <div class="select-data-slot">
    <view-type-toggle
      has-tabs
      v-model="viewTypeModel"
      :variables="variables"
      :available-variables="trainingVariables"
    >
      <b-nav-item
        class="font-weight-bold"
        @click="setIncludedActive"
        :active="includedActive"
        >Samples to Model From</b-nav-item
      >
      <b-nav-item
        class="font-weight-bold mr-auto"
        @click="setExcludedActive"
        :active="!includedActive"
        >Excluded Samples</b-nav-item
      >
    </view-type-toggle>

    <div class="fake-search-input">
      <filter-badge v-if="activeFilter" active-filter :filter="activeFilter" />
      <filter-badge
        v-for="(filter, index) in filters"
        :key="index"
        :filter="filter"
      />
    </div>

    <div class="table-title-container">
      <p class="selection-data-slot-summary">
        <data-size
          :currentSize="numRows"
          :total="numRows"
          @submit="onDataSizeSubmit"
        />
        <strong class="matching-color">matching</strong> samples of
        {{ numRows }} to model<template v-if="selectionNumRows > 0"
          >, {{ selectionNumRows }}
          <strong class="selected-color">selected</strong>
        </template>
      </p>

      <layer-selection v-if="isMultiBandImage" class="layer-select-dropdown" />
      <b-button
        class="select-data-action-exclude"
        v-if="includedActive"
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
        ></i>
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
        ></i>
        Reinclude
      </b-button>
    </div>

    <div class="select-data-container" :class="{ pending: !hasData }">
      <div class="select-data-no-results" v-if="!hasData">
        <div v-html="spinnerHTML"></div>
      </div>
      <component
        :is="viewComponent"
        :included-active="includedActive"
        :instance-name="instanceName"
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
import SelectTimeseriesView from "./SelectTimeseriesView.vue";
import SelectGeoPlot from "./SelectGeoPlot.vue";
import SelectGraphView from "./SelectGraphView.vue";
import FilterBadge from "./FilterBadge.vue";
import ViewTypeToggle from "./ViewTypeToggle.vue";
import LayerSelection from "./LayerSelection.vue";
import { overlayRouteEntry } from "../util/routes";
import {
  actions as datasetActions,
  getters as datasetGetters,
} from "../store/dataset/module";
import {
  TableRow,
  D3M_INDEX_FIELD,
  Variable,
  Highlight,
  RowSelection,
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import {
  Filter,
  FilterParams,
  addFilterToRoute,
  EXCLUDE_FILTER,
  INCLUDE_FILTER,
} from "../util/filters";
import { clearHighlight, createFilterFromHighlight } from "../util/highlights";
import {
  addRowSelection,
  removeRowSelection,
  clearRowSelection,
  isRowSelected,
  getNumIncludedRows,
  getNumExcludedRows,
  createFilterFromRowSelection,
} from "../util/row";
import { actions as appActions } from "../store/app/module";
import { actions as viewActions } from "../store/view/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";

const GEO_VIEW = "geo";
const GRAPH_VIEW = "graph";
const IMAGE_VIEW = "image";
const TABLE_VIEW = "table";
const TIMESERIES_VIEW = "timeseries";

export default Vue.extend({
  name: "select-data-slot",

  components: {
    DataSize,
    FilterBadge,
    ImageMosaic,
    LayerSelection,
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

    availableVariables(): Variable[] {
      return routeGetters.getAvailableVariables(this.$store);
    },
    trainingVariables(): Variable[] {
      return routeGetters.getTrainingVariables(this.$store);
    },
    includedActive(): boolean {
      return routeGetters.getRouteInclude(this.$store);
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
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

    activeFilter(): Filter {
      if (!this.highlight || !this.highlight.value) {
        return null;
      }
      if (this.includedActive) {
        return createFilterFromHighlight(this.highlight, INCLUDE_FILTER);
      }
      return createFilterFromHighlight(this.highlight, EXCLUDE_FILTER);
    },

    /* Check if the Active Filter is from an available feature. */
    isActiveFilterFromAnAvailableFeature(): Boolean {
      if (!this.activeFilter) {
        return false;
      }

      const activeFilterName = this.activeFilter.key;
      const availableVariablesNames = this.availableVariables.map(
        (v) => v.colName
      );

      return availableVariablesNames.includes(activeFilterName);
    },

    /* Disable the Exclude filter button. */
    isExcludeDisabled(): Boolean {
      return (
        (!this.isFilteringHighlights && !this.isFilteringSelection) ||
        this.isActiveFilterFromAnAvailableFeature
      );
    },

    filters(): Filter[] {
      return routeGetters
        .getDecodedFilters(this.$store)
        .filter((f) => f.type !== "row");
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
      return !this.isFilteringSelection && !!this.highlight;
    },

    isFilteringSelection(): boolean {
      return !!this.rowSelection;
    },

    isMultiBandImage(): boolean {
      return routeGetters.isMultiBandImage(this.$store);
    },

    /* Select which component to display the data. */
    viewComponent() {
      if (this.viewTypeModel === GEO_VIEW) return "SelectGeoPlot";
      if (this.viewTypeModel === GRAPH_VIEW) return "SelectGraphView";
      if (this.viewTypeModel === IMAGE_VIEW) return "ImageMosaic";
      if (this.viewTypeModel === TABLE_VIEW) return "SelectDataTable";
      if (this.viewTypeModel === TIMESERIES_VIEW) return "SelectTimeseriesView";
    },
  },

  methods: {
    onExcludeClick() {
      let filter = null;
      if (this.isFilteringHighlights) {
        filter = createFilterFromHighlight(this.highlight, EXCLUDE_FILTER);
      } else {
        filter = createFilterFromRowSelection(
          this.rowSelection,
          EXCLUDE_FILTER
        );
      }

      addFilterToRoute(this.$router, filter);

      if (this.isFilteringHighlights) {
        clearHighlight(this.$router);
      } else {
        clearRowSelection(this.$router);
      }

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
        filter = createFilterFromHighlight(this.highlight, INCLUDE_FILTER);
      } else {
        filter = createFilterFromRowSelection(
          this.rowSelection,
          INCLUDE_FILTER
        );
      }

      addFilterToRoute(this.$router, filter);

      if (this.isFilteringHighlights) {
        clearHighlight(this.$router);
      } else {
        clearRowSelection(this.$router);
      }

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

    setIncludedActive() {
      const entry = overlayRouteEntry(this.$route, {
        include: "true",
      });
      this.$router.push(entry).catch((err) => console.warn(err));

      clearRowSelection(this.$router);
    },

    setExcludedActive() {
      const entry = overlayRouteEntry(this.$route, {
        include: "false",
      });
      this.$router.push(entry).catch((err) => console.warn(err));

      clearRowSelection(this.$router);
    },

    /* When the user request to fetch a different size of data. */
    onDataSizeSubmit(dataSize: number) {
      const entry = overlayRouteEntry(this.$route, { dataSize });
      this.$router.push(entry).catch((err) => console.warn(err));
      viewActions.updateSelectTrainingData(this.$store);
    },
  },
  watch: {
    numRows(newVal: number) {
      const entry = overlayRouteEntry(this.$route, { dataSize: newVal });
      this.$router.push(entry).catch((err) => console.warn(err));
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

.fake-search-input > .filter-badge {
  margin: 2px;
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
