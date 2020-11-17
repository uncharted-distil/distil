<template>
  <div class="view-container">
    <action-column>
      <template slot="actions">
        <b-button
          variant="light"
          title="Create a Timeseries variable"
          @click="onTimeseriesClick"
        >
          <i class="fa fa-area-chart" />
        </b-button>
        <b-button
          variant="light"
          title="Create a Geocoordinate variable"
          @click="onMapClick"
        >
          <i class="fa fa-globe" />
        </b-button>
      </template>
    </action-column>

    <left-side-panel panel-title="Select feature to infer below (target)">
      <variable-facets
        slot="content"
        enable-search
        enable-type-change
        enable-type-filtering
        ignore-highlights
        :facet-count="searchedActiveVariables.length"
        :html="button"
        :instance-name="instanceName"
        :log-activity="problemDefinition"
        :rows-per-page="numRowsPerPage"
        :summaries="summaries"
      />
    </left-side-panel>

    <main class="content">
      <create-solutions-form />
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
      <!-- <p class="selection-data-slot-summary">
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
      </p> -->
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
import { isEmpty } from "lodash";

// Components
import ActionColumn from "../components/layout/ActionColumn.vue";
import CreateSolutionsForm from "../components/CreateSolutionsForm.vue";
import LeftSidePanel from "../components/layout/LeftSidePanel.vue";
import ImageMosaic from "../components/ImageMosaic.vue";
import SelectDataTable from "../components/SelectDataTable.vue";
import SelectGeoPlot from "../components/SelectGeoPlot.vue";
import SelectGraphView from "../components/SelectGraphView.vue";
import SelectTimeseriesView from "../components/SelectTimeseriesView.vue";
import VariableFacets from "../components/facets/VariableFacets.vue";

// Store
import { actions as appActions } from "../store/app/module";
import {
  SummaryMode,
  TableRow,
  Variable,
  VariableSummary,
} from "../store/dataset/index";
import {
  actions as datasetActions,
  getters as datasetGetters,
} from "../store/dataset/module";
import {
  AVAILABLE_TARGET_VARS_INSTANCE,
  GROUPING_ROUTE,
  SELECT_TRAINING_ROUTE,
} from "../store/route/index";
import { getters as routeGetters } from "../store/route/module";
import { actions as viewActions } from "../store/view/module";

// Util
import {
  getVariableSummariesByState,
  NUM_PER_DATA_EXPLORER_PAGE,
  searchVariables,
} from "../util/data";
import { Group } from "../util/facets";
import {
  createRouteEntry,
  overlayRouteEntry,
  varModesToString,
} from "../util/routes";
import { spinnerHTML } from "../util/spinner";
import {
  GEOCOORDINATE_TYPE,
  isUnsupportedTargetVar,
  TIMESERIES_TYPE,
} from "../util/types";
import { Feature, Activity, SubActivity } from "../util/userEvents";

const GEO_VIEW = "geo";
const GRAPH_VIEW = "graph";
const IMAGE_VIEW = "image";
const TABLE_VIEW = "table";
const TIMESERIES_VIEW = "timeseries";

export default Vue.extend({
  name: "DataExplorer",

  components: {
    ActionColumn,
    CreateSolutionsForm,
    LeftSidePanel,
    ImageMosaic,
    SelectDataTable,
    SelectGeoPlot,
    SelectGraphView,
    SelectTimeseriesView,
    VariableFacets,
  },

  data() {
    return {
      instanceName: AVAILABLE_TARGET_VARS_INSTANCE,
      numRowsPerPage: NUM_PER_DATA_EXPLORER_PAGE,
      viewTypeModel: TABLE_VIEW,
    };
  },

  computed: {
    availableTargetVarsPage(): number {
      return routeGetters.getRouteAvailableTargetVarsPage(this.$store);
    },

    availableTargetVarsSearch(): string {
      return routeGetters.getRouteAvailableTargetVarsSearch(this.$store);
    },

    groupedFeatures(): string[] {
      // Fetch the grouped features.
      const groupedFeatures = datasetGetters
        .getGroupings(this.$store)
        .filter((group) => Array.isArray(group.grouping.subIds))
        .map((group) => group.grouping.subIds)
        .flat();
      return groupedFeatures;
    },

    hasData(): boolean {
      return datasetGetters.hasIncludedTableData(this.$store);
    },

    button(): (group: Group) => HTMLElement {
      return (group: Group) => {
        const variable = group.colName;
        const training = routeGetters.getDecodedTrainingVariableNames(
          this.$store
        );
        const isInTraining = training.includes(variable);

        // create a button
        const button = document.createElement("button");
        button.className = "btn btn-sm";
        button.className += isInTraining
          ? " btn-outline-secondary"
          : " btn-primary";
        button.textContent = isInTraining ? "Hide" : "Display";

        const onClick = async () => {
          const route = routeGetters.getRoute(this.$store);
          const task = routeGetters.getRouteTask(this.$store);
          const training = routeGetters.getDecodedTrainingVariableNames(
            this.$store
          );
          const updatedTraining = isInTraining
            ? // Remove the variable from the exploration
              training.filter((v) => v !== variable)
            : // Add the variable to the exploration
              training.concat([variable]);

          // update route with training data
          const entry = overlayRouteEntry(route, {
            training: updatedTraining.join(","),
            task,
          });
          this.$router.push(entry).catch((err) => console.warn(err));

          // update store
          viewActions.updateSelectTrainingData(this.$store);
        };

        // create a button
        button.addEventListener("click", onClick);
        return button;
      };
    },

    problemDefinition(): string {
      return Activity.PROBLEM_DEFINITION;
    },

    searchedActiveVariables(): Variable[] {
      // remove variables used in groupedFeature;
      const activeVariables = this.variables.filter(
        (v) => !this.groupedFeatures.includes(v.colName)
      );

      return searchVariables(activeVariables, this.availableTargetVarsSearch);
    },

    summaries(): VariableSummary[] {
      const pageIndex = routeGetters.getRouteAvailableTargetVarsPage(
        this.$store
      );

      const summaryDictionary = datasetGetters.getVariableSummariesDictionary(
        this.$store
      );

      const currentSummaries = getVariableSummariesByState(
        pageIndex,
        this.numRowsPerPage,
        this.searchedActiveVariables,
        summaryDictionary
      );

      return currentSummaries;
    },

    unsupportedTargets(): Set<string> {
      return new Set(
        this.variables
          .filter((v) => isUnsupportedTargetVar(v.colName, v.colType))
          .map((v) => v.colName)
      );
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
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

  watch: {
    availableTargetVarsPage() {
      viewActions.fetchDataExplorerData(this.$store);
    },

    availableTargetVarsSearch() {
      viewActions.fetchDataExplorerData(this.$store);
    },
  },

  async beforeMount() {
    // Fill up the store
    await viewActions.fetchDataExplorerData(this.$store);

    // If there is no training selected, display the first summary key
    const currentTraining = routeGetters.getTrainingVariables(this.$store);
    const firstSummary = this.summaries?.[0]?.key;
    if (isEmpty(currentTraining) && !!firstSummary) {
      const args = { training: firstSummary };
      const currentRoute = routeGetters.getRoute(this.$store);
      const entry = overlayRouteEntry(currentRoute, args);
      this.$router.push(entry).catch((err) => console.warn(err));
    }

    viewActions.updateSelectTrainingData(this.$store);
  },

  methods: {
    groupingClick(type) {
      const entry = createRouteEntry(GROUPING_ROUTE, {
        dataset: routeGetters.getRouteDataset(this.$store),
        groupingType: type,
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },

    onMapClick() {
      this.groupingClick(GEOCOORDINATE_TYPE);
    },

    onTimeseriesClick() {
      this.groupingClick(TIMESERIES_TYPE);
    },

    spinnerHTML,
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
</style>
