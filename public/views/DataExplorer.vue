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
        enable-search
        enable-type-change
        enable-type-filtering
        :facet-count="searchedActiveVariables.length"
        :html="html"
        ignore-highlights
        :instance-name="instanceName"
        :log-activity="problemDefinition"
        :pagination="searchedActiveVariables.length > numRowsPerPage"
        :rows-per-page="numRowsPerPage"
        :summaries="summaries"
      />
    </left-side-panel>
    <main class="content">
      <create-solutions-form />
      <select-data-slot />
    </main>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

// Components
import ActionColumn from "../components/layout/ActionColumn.vue";
import CreateSolutionsForm from "../components/CreateSolutionsForm.vue";
import LeftSidePanel from "../components/layout/LeftSidePanel.vue";
import SelectDataSlot from "../components/SelectDataSlot.vue";
import VariableFacets from "../components/facets/VariableFacets.vue";

// Store
import { actions as appActions } from "../store/app/module";
import { SummaryMode, Variable, VariableSummary } from "../store/dataset/index";
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
  NUM_PER_TARGET_PAGE,
  searchVariables,
} from "../util/data";
import { Group } from "../util/facets";
import { createRouteEntry, varModesToString } from "../util/routes";
import {
  GEOCOORDINATE_TYPE,
  isUnsupportedTargetVar,
  TIMESERIES_TYPE,
} from "../util/types";
import { Feature, Activity, SubActivity } from "../util/userEvents";

export default Vue.extend({
  name: "DataExplorer",

  components: {
    ActionColumn,
    CreateSolutionsForm,
    LeftSidePanel,
    SelectDataSlot,
    VariableFacets,
  },

  data() {
    return {
      instanceName: AVAILABLE_TARGET_VARS_INSTANCE,
      numRowsPerPage: NUM_PER_TARGET_PAGE,
    };
  },

  computed: {
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

    html(): (group: Group) => HTMLDivElement {
      return (group: Group) => {
        const container = document.createElement("div");
        const targetElem = document.createElement("button");

        const unsupported = this.unsupportedTargets.has(group.colName);
        targetElem.className += "btn btn-sm btn-success mb-2";
        if (unsupported) {
          targetElem.className += " disabled";
        }

        targetElem.innerHTML = "Select Target";
        if (!unsupported) {
          // only add listener on supported target types
          targetElem.addEventListener("click", () => {
            const target = group.colName;
            // remove from training
            const training = routeGetters.getDecodedTrainingVariableNames(
              this.$store
            );
            const index = training.indexOf(target);
            if (index !== -1) {
              training.splice(index, 1);
            }

            const v = this.variables.find((v) => {
              return v.colName === group.colName;
            });
            if (v && v.grouping) {
              if (v.grouping.subIds.length > 0) {
                v.grouping.subIds.forEach((subId) => {
                  const exists = training.find((t) => {
                    return t === subId;
                  });
                  if (!exists) {
                    training.push(subId);
                  }
                });
              } else {
                const exists = training.find((t) => {
                  return t === v.grouping.idCol;
                });
                if (!exists) {
                  training.push(v.grouping.idCol);
                }
              }
            }

            // kick off the fetch task and wait for the result - when we've got it, update the route with info
            const dataset = routeGetters.getRouteDataset(this.$store);
            datasetActions
              .fetchTask(this.$store, {
                dataset: dataset,
                targetName: group.colName,
                variableNames: [],
              })
              .then((response) => {
                const task = response.data.task.join(",");

                const varModesMap = routeGetters.getDecodedVarModes(
                  this.$store
                );
                if (task.includes("timeseries")) {
                  training.forEach((v) => {
                    if (v !== group.colName) {
                      varModesMap.set(v, SummaryMode.Timeseries);
                    }
                  });
                }
                const varModesStr = varModesToString(varModesMap);

                const routeArgs = {
                  target: group.colName,
                  dataset: dataset,
                  filters: routeGetters.getRouteFilters(this.$store),
                  training: training.join(","),
                  task: task,
                  varModes: varModesStr,
                };

                appActions.logUserEvent(this.$store, {
                  feature: Feature.SELECT_TARGET,
                  activity: Activity.PROBLEM_DEFINITION,
                  subActivity: SubActivity.PROBLEM_SPECIFICATION,
                  details: { target: group.colName },
                });

                const entry = createRouteEntry(
                  SELECT_TRAINING_ROUTE,
                  routeArgs
                );
                this.$router.push(entry).catch((err) => console.warn(err));
              })
              .catch((error) => {
                console.error(error);
              });
          });
        }
        container.appendChild(targetElem);
        return container;
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
  },

  beforeMount() {
    viewActions.fetchSelectTargetData(this.$store, true);
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
</style>
