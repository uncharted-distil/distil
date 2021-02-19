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
      <search-bar
        :variables="allVariables"
        :filters="filters"
        :highlight="routeHighlight"
        @lex-query="updateFilterAndHighlightFromLexQuery"
      />

      <!-- Tabs to switch views -->

      <div class="d-flex flex-row align-items-end mt-2">
        <div class="flex-grow-1 mr-2">
          <b-tabs v-model="activeView" class="tab-container">
            <b-tab
              v-for="(view, index) in activeViews"
              :key="index"
              :active="view === activeViews[activeView]"
              :title="capitalize(view)"
            />
          </b-tabs>
        </div>
        <b-button
          v-if="includedActive"
          class="select-data-action-exclude align-self-center"
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
      <!-- <layer-selection v-if="isMultiBandImage" class="layer-select-dropdown" /> -->
      <section class="data-container">
        <div v-if="!hasData" v-html="spinnerHTML" />
        <component
          :is="viewComponent"
          :instance-name="instanceName"
          :included-active="includedActive"
        />
      </section>

      <footer
        class="d-flex align-items-end d-flex justify-content-between mt-1 mb-0"
      >
        <div class="flex-grow-1">
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
        </div>
        <b-button-toolbar>
          <b-button-group class="ml-2 mt-1">
            <b-button
              variant="primary"
              :disabled="includedActive"
              @click="setIncludedActive"
              >Included</b-button
            >
            <b-button
              variant="secondary"
              :disabled="!includedActive"
              @click="setExcludedActive"
              >Excluded</b-button
            >
          </b-button-group>
        </b-button-toolbar>
        <create-solutions-form v-if="isCreateModelPossible" class="ml-2" />
      </footer>
    </main>

    <status-sidebar />
    <status-panel />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { capitalize, isEmpty, isNil } from "lodash";

// Components
import ActionColumn, { Action } from "../components/layout/ActionColumn.vue";
import AddVariablePane from "../components/panel/AddVariablePane.vue";
import CreateSolutionsForm from "../components/CreateSolutionsForm.vue";
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
import { actions as appActions } from "../store/app/module";
import { Highlight, RowSelection, Variable } from "../store/dataset/index";
import {
  actions as datasetActions,
  getters as datasetGetters,
} from "../store/dataset/module";
import {
  DATA_EXPLORER_VAR_INSTANCE,
  ROUTE_PAGE_SUFFIX,
} from "../store/route/index";
import { getters as routeGetters } from "../store/route/module";
import { actions as viewActions } from "../store/view/module";

// Util
import {
  Filter,
  addFilterToRoute,
  deepUpdateFiltersInRoute,
  EXCLUDE_FILTER,
  INCLUDE_FILTER,
} from "../util/filters";
import {
  clearHighlight,
  createFilterFromHighlight,
  updateHighlight,
} from "../util/highlights";
import { lexQueryToFiltersAndHighlight } from "../util/lex";
import { overlayRouteEntry } from "../util/routes";
import {
  clearRowSelection,
  getNumIncludedRows,
  getNumExcludedRows,
  createFilterFromRowSelection,
} from "../util/row";
import { spinnerHTML } from "../util/spinner";
import { META_TYPES } from "../util/types";
import { Feature, Activity, SubActivity } from "../util/userEvents";
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
    CreateSolutionsForm,
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

    /* All variables, only used for lex as we need to parse the hidden variables from groupings */
    allVariables(): Variable[] {
      return datasetGetters.getAllVariables(this.$store);
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

    isCreateModelPossible(): boolean {
      // check that we have some target and training variables.
      return !isNil(this.target) && !isEmpty(this.training);
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

    includedActive(): boolean {
      return routeGetters.getRouteInclude(this.$store);
    },

    /* Disable the Exclude filter button. */
    isExcludeDisabled(): boolean {
      return !this.isFilteringHighlights && !this.isFilteringSelection;
    },

    isFilteringHighlights(): boolean {
      return !!this.highlight;
    },

    isFilteringSelection(): boolean {
      return !!this.rowSelection;
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
      const variables = Array.from(datasetGetters.getVariables(this.$store));
      variables.sort((a, b) => {
        // If their ranking are identical or do not exist
        // sort by importance
        if (a?.ranking === b?.ranking) {
          return b.importance - a.importance;

          // otherwise by ranking
        } else {
          return b.ranking - a.ranking;
        }
      });
      return variables;
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

    updateFilterAndHighlightFromLexQuery(lexQuery) {
      const lqfh = lexQueryToFiltersAndHighlight(lexQuery, this.dataset);
      deepUpdateFiltersInRoute(this.$router, lqfh.filters);
      updateHighlight(this.$router, lqfh.highlight);
    },

    /* When the user request to fetch a different size of data. */
    onDataSizeSubmit(dataSize: number) {
      this.updateRoute({ dataSize });
      viewActions.updateDataExplorerData(this.$store);
    },

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
        target: this.target.key,
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
        target: this.target.key,
      });

      appActions.logUserEvent(this.$store, {
        feature: Feature.UNFILTER_DATA,
        activity: Activity.DATA_PREPARATION,
        subActivity: SubActivity.DATA_TRANSFORMATION,
        details: { filter: filter },
      });
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

    setIncludedActive() {
      const entry = overlayRouteEntry(this.$route, {
        include: "true",
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },

    setExcludedActive() {
      const entry = overlayRouteEntry(this.$route, {
        include: "false",
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },

    updateRoute(args) {
      const entry = overlayRouteEntry(this.$route, args);
      this.$router.push(entry).catch((err) => console.warn(err));
    },

    preSelectTopVariables(number = 5): void {
      // if explore is already filled let's skip
      if (!isEmpty(this.explore)) return;

      // get the top 5 variables
      const top5Variables = [...this.variables]
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
  padding-bottom: 1rem;
  padding-top: 1rem;
}

/* Add padding to all elements but the tabs and data */
.content > *:not(.data-container),
.content > *:not(.tab-container) {
  padding-left: 1rem;
  padding-right: 1rem;
}

.tab-container,
.data-container {
  border-bottom: 1px solid var(--border-color);
}

.data-container {
  background-color: var(--white);
  display: flex;
  flex-flow: wrap;
  height: 100%;
  padding: 1rem;
  position: relative;
  width: 100%;
}
</style>
<style>
.view-container .tab-container ul.nav-tabs {
  border: none;
  margin-bottom: -1px;
}

.view-container .tab-container a.nav-link {
  border: 1px solid transparent;
  border-bottom-color: var(--border-color);
  border-top-width: 3px;
  color: var(--color-text-second);
  margin-bottom: 0;
}

.view-container .tab-container a.nav-link.active {
  background-color: var(--white);
  border-color: var(--border-color);
  border-top-color: var(--primary);
  border-bottom-width: 0;
  border-top-left-radius: 0.25rem;
  border-top-right-radius: 0.25rem;
  color: var(--primary);
  margin-bottom: -1px;
}

.select-data-action-exclude:not([disabled]) .include-highlight,
.select-data-action-exclude:not([disabled]) .exclude-highlight {
  color: var(--blue); /* #255dcc; */
}

.select-data-action-exclude:not([disabled]) .include-selection,
.select-data-action-exclude:not([disabled]) .exclude-selection {
  color: var(--red); /* #ff0067; */
}
</style>
