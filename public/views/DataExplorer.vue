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
      <!-- <div class="fake-search-input">
        <filter-badge
          v-if="activeFilter"
          active-filter
          :filter="activeFilter"
        />
        <filter-badge
          v-for="(filter, index) in filters"
          :key="index"
          :filter="filter"
        />
      </div> -->
      <p class="selection-data-size">
        <data-size
          :current-size="numRows"
          :total="totalNumRows"
          @submit="onDataSizeSubmit"
        />
        <strong class="matching-color">matching</strong> samples of
        {{ totalNumRows }} to model<template v-if="selectionNumRows > 0"
          >, {{ selectionNumRows }}
          <strong class="selected-color">selected</strong>
        </template>
      </p>
      <!-- <layer-selection v-if="isMultiBandImage" class="layer-select-dropdown" /> -->

      <section class="data-container">
        <div v-if="!hasData" v-html="spinnerHTML" />
        <component :is="viewComponent" :instance-name="instanceName" />
      </section>
    </main>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { isEmpty } from "lodash";

// Components
import ActionColumn, { Action } from "../components/layout/ActionColumn.vue";
import AddVariablePane from "../components/layout/AddVariablePane.vue";
import DataSize from "../components/buttons/DataSize.vue";
import FacetListPane from "../components/layout/FacetListPane.vue";
import LeftSidePanel from "../components/layout/LeftSidePanel.vue";
import ImageMosaic from "../components/ImageMosaic.vue";
import SelectDataTable from "../components/SelectDataTable.vue";
import SelectGeoPlot from "../components/SelectGeoPlot.vue";
import SelectGraphView from "../components/SelectGraphView.vue";
import SelectTimeseriesView from "../components/SelectTimeseriesView.vue";

// Store
import { RowSelection, Variable } from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import { DATA_EXPLORER_VAR_INSTANCE } from "../store/route/index";
import { getters as routeGetters } from "../store/route/module";
import { actions as viewActions } from "../store/view/module";

// Util
import { overlayRouteEntry } from "../util/routes";
import { getNumIncludedRows } from "../util/row";
import { spinnerHTML } from "../util/spinner";
import { META_TYPES } from "../util/types";

const GEO_VIEW = "geo";
const GRAPH_VIEW = "graph";
const IMAGE_VIEW = "image";
const TABLE_VIEW = "table";
const TIMESERIES_VIEW = "timeseries";

const ACTIONS = [
  { name: "All Variables", icon: "database", paneId: "available" },
  { name: "Text Variables", icon: "font", paneId: "text" },
  { name: "Categorical Variables", icon: "align-left", paneId: "categorical" },
  { name: "Number Variables", icon: "bar-chart", paneId: "number" },
  { name: "Time Variables", icon: "clock-o", paneId: "time" },
  { name: "Location Variables", icon: "map-o", paneId: "location" },
  { name: "Image Variables", icon: "image", paneId: "image" },
  { name: "Unknown Variables", icon: "question", paneId: "unknown" },
  { name: "Selected Variables", icon: "check", paneId: "selected" },
  { name: "Create New Variable", icon: "plus", paneId: "add" },
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
    SelectDataTable,
    SelectGeoPlot,
    SelectGraphView,
    SelectTimeseriesView,
  },

  data() {
    return {
      actions: ACTIONS,
      activePane: "selected",
      instanceName: DATA_EXPLORER_VAR_INSTANCE,
      metaTypes: Object.keys(META_TYPES),
      viewTypeModel: TABLE_VIEW,
    };
  },

  computed: {
    /* Variables displayed on the Facet Panel */
    activeVariables(): Variable[] {
      return this.availableVariables[this.activePane];
    },

    availableVariables(): any {
      const cleanTraining = this.training.map((t) => t.toLowerCase());

      const selectedVariables = this.variables.filter((v) =>
        cleanTraining.includes(v.colName.toLowerCase())
      );
      const nonSelectedVariables = this.variables.filter(
        (v) => !cleanTraining.includes(v.colName.toLowerCase())
      );

      const variables = {};
      this.activeActions.forEach((action) => {
        if (action.paneId === "add") variables[action.paneId] = null;
        else if (action.paneId === "selected") {
          variables[action.paneId] = selectedVariables;
        } else if (action.paneId === "available") {
          variables[action.paneId] = nonSelectedVariables;
        } else {
          variables[action.paneId] = nonSelectedVariables.filter((variable) =>
            META_TYPES[action.paneId].includes(variable.colType)
          );
        }
      });

      return variables;
    },

    /* Actions displayed on the Action column */
    activeActions(): Action[] {
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

    hasData(): boolean {
      return datasetGetters.hasIncludedTableData(this.$store);
    },

    hasNoVariables(): boolean {
      return isEmpty(this.activeVariables);
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

    totalNumRows(): number {
      return this.hasData
        ? datasetGetters.getIncludedTableDataNumRows(this.$store)
        : 0;
    },

    training(): string[] {
      return routeGetters.getDecodedTrainingVariableNames(this.$store);
    },

    variablesTypes(): string[] {
      return [...new Set(this.variables.map((v) => v.colType))];
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    viewComponent() {
      if (this.viewTypeModel === GEO_VIEW) return "SelectGeoPlot";
      if (this.viewTypeModel === GRAPH_VIEW) return "SelectGraphView";
      if (this.viewTypeModel === IMAGE_VIEW) return "ImageMosaic";
      if (this.viewTypeModel === TABLE_VIEW) return "SelectDataTable";
      if (this.viewTypeModel === TIMESERIES_VIEW) return "SelectTimeseriesView";
    },
  },

  async beforeMount() {
    // Fill up the store
    await viewActions.fetchDataExplorerData(this.$store);

    // Update the training data
    viewActions.updateSelectTrainingData(this.$store);
  },

  methods: {
    /* When the user request to fetch a different size of data. */
    onDataSizeSubmit(dataSize: number) {
      this.updateRoute({ dataSize });
    },

    onSetActive(actionName: string): void {
      let activePane = "";
      if (actionName !== "") {
        activePane = this.actions.find((a) => a.name === actionName).paneId;
      }
      this.activePane = activePane;
    },

    updateRoute(args) {
      const entry = overlayRouteEntry(this.$route, args);
      this.$router.push(entry).catch((err) => console.warn(err));
      viewActions.updateSelectTrainingData(this.$store);
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

.view-container .content {
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

.matching-color {
  color: var(--blue);
}
.selected-color {
  color: var(--red);
}
</style>
