<template>
  <div class="view-container">
    <action-column
      :actions="activeActions"
      :current-action="currentAction"
      @set-active-pane="onSetActive"
    />

    <left-side-panel :panel-title="currentAction">
      <add-variable-pane v-if="activePane === 'add'" />
      <template v-else>
        <template v-if="hasNoVariables">
          <p v-if="activePane === 'selected'">Select a variable to explore.</p>
          <p v-else>All the variables of that type are selected.</p>
        </template>
        <facet-list-pane v-else :variables="activeVariables" />
      </template>
    </left-side-panel>

    <main class="content">
      <search-input class="mb-3" />
      <search-bar
        class="mb-3"
        :variables="variables"
        :filters="filters"
        :highlight="routeHighlight"
        @lex-query="updateFilterFromLexQuery"
      />

      <!-- Tabs to switch views -->
      <b-tabs v-model="activeView" pills class="mb-3">
        <b-tab
          v-for="(view, index) in activeViews"
          :key="index"
          :active="view === activeViews[activeView]"
          :title="capitalize(view)"
        />
      </b-tabs>

      <!-- <layer-selection v-if="isMultiBandImage" class="layer-select-dropdown" /> -->
      <section class="data-container">
        <div v-if="!hasData" v-html="spinnerHTML" />
        <component :is="viewComponent" :instance-name="instanceName" />
      </section>

      <p class="selection-data-size mt-2 mb-0">
        <data-size
          :current-size="numRows"
          :total="totalNumRows"
          @submit="onDataSizeSubmit"
        />
        <strong class="matching-color">matching</strong> samples of
        {{ totalNumRows }} to model<template v-if="selectionNumRows > 0">
          , {{ selectionNumRows }}
          <strong class="selected-color">selected</strong>
        </template>
      </p>
    </main>

    <status-sidebar />
    <status-panel />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { capitalize, isEmpty } from "lodash";

// Components
import ActionColumn, { Action } from "../components/layout/ActionColumn.vue";
import AddVariablePane from "../components/panel/AddVariablePane.vue";
import DataSize from "../components/buttons/DataSize.vue";
import FacetListPane from "../components/panel/FacetListPane.vue";
import LeftSidePanel from "../components/layout/LeftSidePanel.vue";
import ImageMosaic from "../components/ImageMosaic.vue";
import SearchBar from "../components/layout/SearchBar.vue";
import SearchInput from "../components/SearchInput.vue";
import SelectDataTable from "../components/SelectDataTable.vue";
import SelectGeoPlot from "../components/SelectGeoPlot.vue";
import SelectGraphView from "../components/SelectGraphView.vue";
import SelectTimeseriesView from "../components/SelectTimeseriesView.vue";
import StatusPanel from "../components/StatusPanel.vue";
import StatusSidebar from "../components/StatusSidebar.vue";

// Store
import { Highlight, RowSelection, Variable } from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import {
  DATA_EXPLORER_VAR_INSTANCE,
  ROUTE_PAGE_SUFFIX,
} from "../store/route/index";
import { getters as routeGetters } from "../store/route/module";
import { actions as viewActions } from "../store/view/module";

// Util
import { deepUpdateFiltersInRoute } from "../util/filters";
import { lexQueryToFiltersAndHighlight } from "../util/lex";
import { overlayRouteEntry } from "../util/routes";
import { getNumIncludedRows } from "../util/row";
import { spinnerHTML } from "../util/spinner";
import { META_TYPES } from "../util/types";
import {
  GEO_VIEW,
  GRAPH_VIEW,
  IMAGE_VIEW,
  TABLE_VIEW,
  TIMESERIES_VIEW,
  filterViews,
} from "../util/view";

const ACTIONS = [
  { name: "Create New Variable", icon: "plus", paneId: "add" },
  { name: "All Variables", icon: "database", paneId: "available" },
  { name: "Text Variables", icon: "font", paneId: "text" },
  { name: "Categorical Variables", icon: "align-left", paneId: "categorical" },
  { name: "Number Variables", icon: "bar-chart", paneId: "number" },
  { name: "Time Variables", icon: "clock-o", paneId: "time" },
  { name: "Location Variables", icon: "map-o", paneId: "location" },
  { name: "Image Variables", icon: "image", paneId: "image" },
  { name: "Unknown Variables", icon: "question", paneId: "unknown" },
  { name: "Target Variable", icon: "crosshairs", paneId: "target" },
  { name: "Training Variable", icon: "asterisk", paneId: "training" },
] as Action[];

export default Vue.extend({
  name: "DataExplorer",

  components: {
    ActionColumn,
    AddVariablePane,
    DataSize,
    FacetListPane,
    LeftSidePanel,
    ImageMosaic,
    SearchBar,
    SearchInput,
    SelectDataTable,
    SelectGeoPlot,
    SelectGraphView,
    SelectTimeseriesView,
    StatusPanel,
    StatusSidebar,
  },

  data() {
    return {
      actions: ACTIONS,
      activePane: "available",
      activeView: 0, // TABLE_VIEW
      instanceName: DATA_EXPLORER_VAR_INSTANCE,
      metaTypes: Object.keys(META_TYPES),
    };
  },

  computed: {
    /* Actions displayed on the Action column */
    activeActions(): Action[] {
      return this.availableActions.map((action) => {
        const count = this.variablesPerActions[action.paneId]?.length;
        return count ? { ...action, count } : action;
      });
    },

    /* Variables displayed on the Facet Panel */
    activeVariables(): Variable[] {
      return this.variablesPerActions[this.activePane] ?? [];
    },

    activeViews(): string[] {
      return filterViews(this.variables);
    },

    /* Actions available based on the variables meta types */
    availableActions(): Action[] {
      // Remove the inactive MetaTypes
      return this.actions.filter(
        (action) => !this.inactiveMetaTypes.includes(action.paneId)
      );
    },

    currentAction(): string {
      return (
        this.activePane &&
        this.actions.find((a) => a.paneId === this.activePane).name
      );
    },

    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    explore(): string[] {
      return routeGetters.getExploreVariables(this.$store);
    },

    filters(): string {
      return routeGetters.getRouteFilters(this.$store);
    },

    hasData(): boolean {
      return datasetGetters.hasIncludedTableData(this.$store);
    },

    hasNoVariables(): boolean {
      return isEmpty(this.activeVariables);
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    routeHighlight(): string {
      return routeGetters.getRouteHighlight(this.$store);
    },

    inactiveMetaTypes(): string[] {
      // Go trough each meta type
      return this.metaTypes.map((metaType) => {
        // test if some variables types...
        const typeNotInMetaTypes = !this.variablesTypes.some((t) =>
          // ...is in that meta type
          META_TYPES[metaType].includes(t)
        );
        if (typeNotInMetaTypes) return metaType;
      });
    },

    numRows(): number {
      return this.hasData
        ? datasetGetters.getIncludedTableDataLength(this.$store)
        : 0;
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    selectionNumRows(): number {
      return getNumIncludedRows(this.rowSelection);
    },

    spinnerHTML,

    target(): Variable {
      return routeGetters.getTargetVariable(this.$store);
    },

    totalNumRows(): number {
      return this.hasData
        ? datasetGetters.getIncludedTableDataNumRows(this.$store)
        : 0;
    },

    training(): string[] {
      return routeGetters.getDecodedTrainingVariableNames(this.$store);
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    variablesPerActions() {
      const variables = {};
      this.availableActions.forEach((action) => {
        if (action.paneId === "add") variables[action.paneId] = null;
        else if (action.paneId === "available") {
          variables[action.paneId] = this.variables;
        } else if (action.paneId === "target") {
          variables[action.paneId] = this.target ? [this.target] : [];
        } else if (action.paneId === "training") {
          variables[action.paneId] = this.variables.filter((variable) =>
            this.training.includes(variable.key)
          );
        } else {
          variables[action.paneId] = this.variables.filter((variable) =>
            META_TYPES[action.paneId].includes(variable.colType)
          );
        }
      });

      return variables;
    },

    variablesTypes(): string[] {
      return [...new Set(this.variables.map((v) => v.colType))];
    },

    viewComponent() {
      const viewType = this.activeViews[this.activeView] as string;
      if (viewType === GEO_VIEW) return "SelectGeoPlot";
      if (viewType === GRAPH_VIEW) return "SelectGraphView";
      if (viewType === IMAGE_VIEW) return "ImageMosaic";
      if (viewType === TABLE_VIEW) return "SelectDataTable";
      if (viewType === TIMESERIES_VIEW) return "SelectTimeseriesView";

      // Default is TABLE_VIEW
      return "SelectDataTable";
    },
  },

  // Update either the summaries or explore data on user interaction.
  watch: {
    activeVariables(n, o) {
      if (n === o) return;
      viewActions.fetchDataExplorerData(this.$store, this.activeVariables);
    },

    filters(n, o) {
      if (n === o) return;
      viewActions.updateDataExplorerData(this.$store);
    },

    highlight(n, o) {
      if (n === o) return;
      viewActions.updateDataExplorerData(this.$store);
    },

    explore(n, o) {
      if (n === o) return;
      viewActions.updateDataExplorerData(this.$store);
    },
  },

  async beforeMount() {
    // First get the dataset informations
    await viewActions.fetchDataExplorerData(this.$store, [] as Variable[]);

    // Pre-select the top 5 variables by importance
    this.preSelectTopVariables();

    // Update the explore data
    viewActions.updateDataExplorerData(this.$store);
  },

  methods: {
    capitalize,

    updateFilterFromLexQuery(lexQuery) {
      const lqfh = lexQueryToFiltersAndHighlight(lexQuery, this.dataset);
      deepUpdateFiltersInRoute(this.$router, lqfh.filters);
    },

    /* When the user request to fetch a different size of data. */
    onDataSizeSubmit(dataSize: number) {
      this.updateRoute({ dataSize });
      viewActions.updateDataExplorerData(this.$store);
    },

    onSetActive(actionName: string): void {
      if (actionName === this.activePane) return;

      let activePane = "available"; // default
      if (actionName !== "") {
        activePane = this.actions.find((a) => a.name === actionName).paneId;
      }
      this.activePane = activePane;

      // update the selected pane, and reset the page var to 1
      this.updateRoute({
        pane: activePane,
        [`${DATA_EXPLORER_VAR_INSTANCE}${ROUTE_PAGE_SUFFIX}`]: 1,
      });
    },

    updateRoute(args) {
      const entry = overlayRouteEntry(this.$route, args);
      this.$router.push(entry).catch((err) => console.warn(err));
    },

    preSelectTopVariables(number = 5): void {
      // if explore is already filled let's skip
      if (!isEmpty(this.explore)) return;

      // get the top 5 variables sorted by importance
      const top5Variables = [...this.variables]
        .sort((a, b) => b.importance - a.importance)
        .slice(0, number)
        .map((variable) => variable.key)
        .join(",");

      // Update the route with the top 5 variable as training
      this.updateRoute({ explore: top5Variables });
    },
  },
});
</script>

<style scoped>
.view-container {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  flex-grow: 1;
  height: var(--content-full-height);
  margin-top: var(--navbar-outer-height);
  overflow: hidden;
}

/* Make some elements of a container unsquishable. */
.view-container > *:not(.content),
.content > *:not(.data-container) {
  flex-shrink: 0;
}

.content {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
  padding: 1rem;
}

.data-container {
  display: flex;
  flex-flow: wrap;
  height: 100%;
  position: relative;
  width: 100%;
}
</style>
