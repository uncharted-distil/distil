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
        <facet-list-pane
          v-else
          :variables="activeVariables"
          :enable-color-scales="geoVarExists"
        />
      </template>
    </left-side-panel>

    <main class="content">
      <search-bar
        :variables="allVariables"
        :filters="filters"
        :highlights="routeHighlight"
        handle-updates
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
          v-if="include"
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
          v-if="!include"
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
          :included-active="include"
          :dataset="dataset"
          :data-fields="fields"
          :timeseries-info="timeseries"
          :data-items="items"
          :baseline-items="baselineItems"
          :summaries="summaries"
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
              :disabled="include"
              @click="setIncludedActive"
            >
              Included
            </b-button>
            <b-button
              variant="secondary"
              :disabled="!include"
              @click="setExcludedActive"
            >
              Excluded
            </b-button>
          </b-button-group>
        </b-button-toolbar>
        <create-solutions-form
          v-if="isCreateModelPossible"
          class="ml-2"
          @create-model="onModelCreation"
        />
      </footer>
    </main>
    <left-side-panel
      v-if="activePane === 'outcome'"
      :panel-title="currentAction"
    >
      <add-variable-pane v-if="activePane === 'add'" />
      <template v-else>
        <template v-if="hasNoVariables">
          <p v-if="activePane === 'selected'">Select a variable to explore.</p>
          <p v-else>No Outcome Variables available.</p>
        </template>
        <facet-list-pane
          v-else
          :variables="activeVariables"
          :enable-color-scales="geoVarExists"
        />
      </template>
    </left-side-panel>
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
import SelectDataTable from "../components/SelectDataTable.vue";
import GeoPlot from "../components/GeoPlot.vue";
import SelectGraphView from "../components/SelectGraphView.vue";
import SelectTimeseriesView from "../components/SelectTimeseriesView.vue";
import StatusPanel from "../components/StatusPanel.vue";
import StatusSidebar from "../components/StatusSidebar.vue";

// Store
import {
  appActions,
  viewActions,
  datasetActions,
  datasetGetters,
  requestActions,
} from "../store";
import {
  DataMode,
  Highlight,
  RowSelection,
  TableColumn,
  TableRow,
  TimeSeries,
  Variable,
  VariableSummary,
} from "../store/dataset/index";
import {
  DATA_EXPLORER_VAR_INSTANCE,
  ROUTE_PAGE_SUFFIX,
} from "../store/route/index";
import { getters as routeGetters } from "../store/route/module";

// Util
import {
  addFilterToRoute,
  EXCLUDE_FILTER,
  INCLUDE_FILTER,
} from "../util/filters";
import {
  clearHighlight,
  createFiltersFromHighlights,
} from "../util/highlights";
import { overlayRouteEntry, varModesToString } from "../util/routes";
import {
  clearRowSelection,
  getNumIncludedRows,
  createFilterFromRowSelection,
} from "../util/row";
import { spinnerHTML } from "../util/spinner";
import { isGeoLocatedType, META_TYPES } from "../util/types";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import {
  GEO_VIEW,
  GRAPH_VIEW,
  IMAGE_VIEW,
  TABLE_VIEW,
  TIMESERIES_VIEW,
  filterViews,
} from "../util/view";
import { Dictionary } from "vue-router/types/router";
import {
  BaseState,
  ResultViewState,
  SelectViewState,
} from "../util/state/AppStateWrapper";
import { SolutionRequestMsg } from "../store/requests/actions";
import { Solution } from "../store/requests";

const ACTIONS = [
  { name: "Create New Variable", icon: "fa fa-plus", paneId: "add" },
  { name: "All Variables", icon: "fa fa-database", paneId: "available" },
  { name: "Text Variables", icon: "fa fa-font", paneId: "text" },
  {
    name: "Categorical Variables",
    icon: "fa fa-align-left",
    paneId: "categorical",
  },
  { name: "Number Variables", icon: "fa fa-bar-chart", paneId: "number" },
  { name: "Time Variables", icon: "fa fa-clock-o", paneId: "time" },
  { name: "Location Variables", icon: "fa fa-map-o", paneId: "location" },
  { name: "Image Variables", icon: "fa fa-image", paneId: "image" },
  { name: "Unknown Variables", icon: "fa fa-question", paneId: "unknown" },
  { name: "Target Variable", icon: "fa fa-crosshairs", paneId: "target" },
  { name: "Training Variables", icon: "fa fa-asterisk", paneId: "training" },
  { name: "Outcome Variables", icon: "fas fa-poll", paneId: "outcome" },
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
    SelectDataTable,
    GeoPlot,
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
      include: true,
      state: new SelectViewState(),
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
      return this.state.getLexBarVariables();
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

    highlights(): Highlight[] {
      return routeGetters.getDecodedHighlights(this.$store);
    },

    isCreateModelPossible(): boolean {
      // check that we have some target and training variables.
      return !isNil(this.target) && !isEmpty(this.training);
    },
    timeseries(): Dictionary<TimeSeries> {
      return datasetGetters.getTimeseries(this.$store);
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
    fields(): Dictionary<TableColumn> {
      return this.state.getFields(this.include);
    },
    /* Disable the Exclude filter button. */
    isExcludeDisabled(): boolean {
      return !this.isFilteringHighlights && !this.isFilteringSelection;
    },

    isFilteringHighlights(): boolean {
      return this.highlights && this.highlights.length > 0;
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
      const variables = this.state.getVariables();
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
        } else if (action.paneId === "outcome") {
          variables[action.paneId] = [];
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
    geoVarExists(): boolean {
      const varSums = this.state.getVariableSummaries(this.include);
      return varSums.some((v) => {
        return isGeoLocatedType(v.type);
      });
    },
    viewComponent() {
      const viewType = this.activeViews[this.activeView] as string;
      if (viewType === GEO_VIEW) return "GeoPlot";
      if (viewType === GRAPH_VIEW) return "SelectGraphView";
      if (viewType === IMAGE_VIEW) return "ImageMosaic";
      if (viewType === TABLE_VIEW) return "SelectDataTable";
      if (viewType === TIMESERIES_VIEW) return "SelectTimeseriesView";

      // Default is TABLE_VIEW
      return "SelectDataTable";
    },
    items(): TableRow[] {
      return this.state.getData(this.include);
    },
    baselineItems(): TableRow[] {
      return this.state.getMapBaseline();
    },
    summaries(): VariableSummary[] {
      return this.state.getVariableSummaries(this.include);
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
    geoVarExists() {
      const route = routeGetters.getRoute(this.$store);
      const entry = overlayRouteEntry(route, { hasGeoData: this.geoVarExists });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
  },

  async beforeMount() {
    // First get the dataset informations
    await viewActions.fetchDataExplorerData(this.$store, [] as Variable[]);
    this.state.fetchMapBaseline();
    // Pre-select the top 5 variables by importance
    this.preSelectTopVariables();

    // Update the explore data
    viewActions.updateDataExplorerData(this.$store);
  },

  methods: {
    capitalize,

    /* When the user request to fetch a different size of data. */
    onDataSizeSubmit(dataSize: number) {
      this.updateRoute({ dataSize });
      viewActions.updateDataExplorerData(this.$store);
    },
    onModelCreation(solutionRequestMsg: SolutionRequestMsg) {
      // handle solutionRequestMsg
      requestActions
        .createSolutionRequest(this.$store, solutionRequestMsg)
        .then(async (res: Solution) => {
          const dataMode = routeGetters.getDataMode(this.$store);
          const dataModeDefault = dataMode ? dataMode : DataMode.Default;
          // transition to result screen
          const entry = overlayRouteEntry(this.$route, {
            dataset: routeGetters.getRouteDataset(this.$store),
            target: routeGetters.getRouteTargetVariable(this.$store),
            solutionId: res.solutionId,
            task: routeGetters.getRouteTask(this.$store),
            dataMode: dataModeDefault,
            varModes: varModesToString(
              routeGetters.getDecodedVarModes(this.$store)
            ),
            modelLimit: routeGetters.getModelLimit(this.$store),
            modelTimeLimit: routeGetters.getModelTimeLimit(this.$store),
            modelQuality: routeGetters.getModelQuality(this.$store),
          });
          this.$router.push(entry).catch((err) => console.warn(err));
          this.setState(new ResultViewState());
          await this.state.fetchVariables();
        })
        .catch((err) => {
          console.error(err);
        });
      return;
    },
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
    setState(state: BaseState) {
      this.state = state;
    },
    setIncludedActive() {
      this.include = true;
    },

    setExcludedActive() {
      this.include = false;
    },

    updateRoute(args) {
      const entry = overlayRouteEntry(this.$route, args);
      this.$router.push(entry).catch((err) => console.warn(err));
    },

    resetHighlightsOrRow() {
      if (this.isFilteringHighlights) {
        clearHighlight(this.$router);
      } else {
        clearRowSelection(this.$router);
      }
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
