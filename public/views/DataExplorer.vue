<template>
  <div class="view-container">
    <action-column
      :actions="actions"
      :currentAction="currentAction"
      @set-active-pane="onSetActive"
    />

    <left-side-panel :panel-title="currentAction">
      <template slot="content">
        <add-variable-pane v-if="activePane === 'add'" />
        <facet-list-pane v-else />
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
          :currentSize="numRows"
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
        <div v-if="!hasData" v-html="spinnerHTML"></div>
        <component :is="viewComponent" :instance-name="instanceName" />
      </section>
    </main>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

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
import { RowSelection } from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import { DATA_EXPLORER_VAR_INSTANCE } from "../store/route/index";
import { getters as routeGetters } from "../store/route/module";
import { actions as viewActions } from "../store/view/module";

// Util
import { overlayRouteEntry } from "../util/routes";
import { getNumIncludedRows } from "../util/row";
import { spinnerHTML } from "../util/spinner";

const GEO_VIEW = "geo";
const GRAPH_VIEW = "graph";
const IMAGE_VIEW = "image";
const TABLE_VIEW = "table";
const TIMESERIES_VIEW = "timeseries";

const ACTIONS = [
  { name: "Selected Variables", icon: "eye", paneId: "selected" },
  { name: "Create Variable", icon: "plus", paneId: "add" },
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
      activePane: ACTIONS[0].paneId,
      instanceName: DATA_EXPLORER_VAR_INSTANCE,
      viewTypeModel: TABLE_VIEW,
    };
  },

  computed: {
    currentAction(): string {
      return (
        this.activePane &&
        this.actions.find((a) => a.paneId === this.activePane).name
      );
    },

    hasData(): boolean {
      return datasetGetters.hasIncludedTableData(this.$store);
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
