<template>
  <div
    class="training-variables"
    v-bind:class="{ included: includedActive, excluded: !includedActive }"
  >
    <p class="nav-link font-weight-bold">
      Features to Model
      <i class="float-right fa fa-angle-right fa-lg"></i>
    </p>
    <variable-facets
      ref="facets"
      enable-highlighting
      enable-search
      enable-type-change
      :facetCount="trainingVariables.length"
      :html="html"
      :isAvailableFeatures="false"
      :isFeaturesToModel="true"
      :log-activity="logActivity"
      :instance-name="instanceName"
      :pagination="trainingVariables.length > numRowsPerPage"
      :rows-per-page="numRowsPerPage"
      :summaries="trainingVariableSummaries"
    >
      <div class="available-variables-menu">
        <div>
          {{ subtitle }}
        </div>
        <div v-if="isAllTrainingVariablesRemovable">
          <b-button size="sm" variant="outline-secondary" @click="removeAll"
            >Remove All</b-button
          >
        </div>
      </div>
      <div v-if="trainingVariableSummaries.length === 0">
        <i class="no-selections-icon fa fa-arrow-circle-left"></i>
      </div>
    </variable-facets>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import VariableFacets from "./facets/VariableFacets.vue";
import { Variable, VariableSummary, Highlight } from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";
import { TRAINING_VARS_INSTANCE } from "../store/route/index";
import { Group } from "../util/facets";
import { NUM_PER_PAGE, getVariableSummariesByState } from "../util/data";
import { overlayRouteEntry } from "../util/routes";
import { removeFiltersByName } from "../util/filters";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";

export default Vue.extend({
  name: "training-variables",

  components: {
    VariableFacets,
  },

  data() {
    return {
      logActivity: Activity.DATA_PREPARATION,
    };
  },

  computed: {
    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    numRowsPerPage(): number {
      return NUM_PER_PAGE;
    },
    includedActive(): boolean {
      return routeGetters.getRouteInclude(this.$store);
    },
    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },
    trainingVariables(): Variable[] {
      return routeGetters.getTrainingVariables(this.$store);
    },
    trainingVariableSummaries(): VariableSummary[] {
      const pageIndex = routeGetters.getRouteTrainingVarsPage(this.$store);
      const include = routeGetters.getRouteInclude(this.$store);
      const summaryDictionary = include
        ? datasetGetters.getIncludedVariableSummariesDictionary(this.$store)
        : datasetGetters.getExcludedVariableSummariesDictionary(this.$store);

      const currentSummaries = getVariableSummariesByState(
        pageIndex,
        this.numRowsPerPage,
        this.trainingVariables,
        summaryDictionary
      );

      return currentSummaries;
    },

    /**
     * Check if all the training variables are removable.
     * @return {Boolean}
     */
    isAllTrainingVariablesRemovable(): boolean {
      // Fetch the variables in the timeseries grouping.
      const timeseriesGrouping = datasetGetters.getTimeseriesGroupingVariables(
        this.$store
      );

      // Filter them out of the available training variables.
      const trainingVariables = Array.from(this.trainingVariables).filter(
        (variable) => !timeseriesGrouping.includes(variable.colName)
      );

      // The variables can be removed.
      return trainingVariables.length > 0;
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
    subtitle(): string {
      return `${this.trainingVariables.length} features selected`;
    },
    instanceName(): string {
      return TRAINING_VARS_INSTANCE;
    },

    html(): (Group) => HTMLDivElement {
      return (group: Group) => {
        // exclude remove button if the var is an id / sub-id of the
        // timeseries grouping.

        const targetVar = this.variables.find((v) => {
          return v.colName === this.target;
        });

        if (targetVar?.grouping) {
          let isGroupingID = false;
          if (targetVar.grouping.subIds.length > 0) {
            isGroupingID = !!targetVar.grouping.subIds.find((v) => {
              return v === group.colName;
            });
          } else {
            isGroupingID = targetVar.grouping.idCol === group.key;
          }
          if (isGroupingID) {
            return;
          }
        }

        const container = document.createElement("div");
        const remove = document.createElement("button");
        remove.className += "btn btn-sm btn-outline-secondary mr-1 mb-2";
        remove.innerHTML = "Remove";

        remove.addEventListener("click", async () => {
          appActions.logUserEvent(this.$store, {
            feature: Feature.REMOVE_FEATURE,
            activity: Activity.DATA_PREPARATION,
            subActivity: SubActivity.DATA_TRANSFORMATION,
            details: { feature: group.colName },
          });

          const training = routeGetters.getDecodedTrainingVariableNames(
            this.$store
          );
          training.splice(training.indexOf(group.colName), 1);

          // update task based on the current training data
          const taskResponse = await datasetActions.fetchTask(this.$store, {
            dataset: routeGetters.getRouteDataset(this.$store),
            targetName: routeGetters.getRouteTargetVariable(this.$store),
            variableNames: training,
          });

          const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
            training: training.join(","),
            task: taskResponse.data.task.join(","),
          });
          this.$router.push(entry);
          removeFiltersByName(this.$router, group.colName);
        });
        container.appendChild(remove);
        return container;
      };
    },
  },

  methods: {
    removeAll() {
      appActions.logUserEvent(this.$store, {
        feature: Feature.REMOVE_ALL_FEATURES,
        activity: Activity.DATA_PREPARATION,
        subActivity: SubActivity.DATA_TRANSFORMATION,
        details: {},
      });

      // Fetch the variables in the timeseries grouping.
      const timeseriesGrouping = datasetGetters.getTimeseriesGroupingVariables(
        this.$store
      );

      // Retain only variables used in group on remove all since they can't
      // be actually be removed without ungrouping
      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        training: timeseriesGrouping.join(","),
        trainingVarsPage: 1,
      });
      this.$router.push(entry);
    },
  },
});
</script>

<style>
.training-variables {
  display: flex;
  flex-direction: column;
}
.no-selections-icon {
  color: #32cd32;
  font-size: 46px;
}
.training-variables-menu {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
  line-height: 30px;
}

.training-variables /deep/ .variable-facets-wrapper {
  height: 100%;
}
</style>
