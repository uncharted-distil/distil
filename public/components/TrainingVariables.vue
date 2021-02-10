<template>
  <div
    class="training-variables"
    :class="{ included: includedActive, excluded: !includedActive }"
  >
    <p class="nav-link font-weight-bold">
      Features to Model
      <i class="float-right fa fa-angle-right fa-lg" />
    </p>
    <variable-facets
      ref="facets"
      enable-highlighting
      enable-search
      enable-type-change
      :facet-count="trainingVariables.length"
      :html="html"
      :is-available-features="false"
      :is-features-to-model="true"
      :log-activity="logActivity"
      :instance-name="instanceName"
      :pagination="trainingVariables.length > numRowsPerPage"
      :rows-per-page="numRowsPerPage"
      :summaries="trainingVariableSummaries"
    >
      <div
        class="d-flex flex-row justify-content-between align-items-center my-2 mx-1"
      >
        <div>
          {{ subtitle }}
        </div>
        <div v-if="isAllTrainingVariablesRemovable">
          <b-button size="sm" variant="outline-secondary" @click="removeAll">
            Remove All
          </b-button>
        </div>
      </div>
      <div v-if="trainingVariableSummaries.length === 0">
        <i class="no-selections-icon fa fa-arrow-circle-left" />
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
import {
  getComposedVariableKey,
  NUM_PER_PAGE,
  getVariableSummariesByState,
  searchVariables,
  filterHiddenVariables,
} from "../util/data";
import { overlayRouteEntry } from "../util/routes";
import { removeFiltersByName } from "../util/filters";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";

export default Vue.extend({
  name: "TrainingVariables",

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
    trainingVarsSearch(): string {
      return routeGetters.getRouteTrainingVarsSearch(this.$store);
    },
    trainingVariables(): Variable[] {
      const searchVars = searchVariables(
        routeGetters.getTrainingVariables(this.$store),
        this.trainingVarsSearch
      );
      return filterHiddenVariables(this.variables, searchVars);
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

    /* Check if all the training variables are removable. */
    isAllTrainingVariablesRemovable(): boolean {
      // Fetch the variables in the timeseries grouping.
      const timeseriesGrouping = datasetGetters.getTimeseriesGroupingVariables(
        this.$store
      );

      // Filter them out of the available training variables.
      const trainingVariables = Array.from(this.trainingVariables).filter(
        (variable) => !timeseriesGrouping.includes(variable.key)
      );

      // The variables can be removed.
      return trainingVariables.length > 0;
    },

    isTimeseries(): boolean {
      return routeGetters.isTimeseries(this.$store);
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

    html(): (Group) => HTMLElement {
      return (group: Group) => {
        // get the target variable information
        const targetVar = this.variables.find((v) => {
          return v.key === this.target;
        });

        // the target variable contains grouping information (Timeseries/Geocoordinate)
        if (!!targetVar?.grouping) {
          let hideRemoveButton = false;
          const seriesIds = targetVar.grouping.subIds;

          // Check there is any series IDs
          if (seriesIds.length > 0) {
            // Make sure to show the button for all of them.
            if (seriesIds.length !== 1) {
              hideRemoveButton = !seriesIds.some((v) => v === group.key);
            }
            // unless there is only one series ID, then we hide the remove button.
            else {
              hideRemoveButton = true;
            }
          } else {
            hideRemoveButton = targetVar.grouping.idCol === group.groupKey;
          }

          // Hide the remove button
          if (hideRemoveButton) return;
        }

        // Create the remove button
        const removeBtn = document.createElement("button");
        removeBtn.className += "btn btn-sm btn-outline-secondary mr-1 mb-2";
        removeBtn.textContent = "Remove";

        // Is the variable of categorical type
        const isCategorical: boolean = group.type === "categorical";

        if (this.isTimeseries && isCategorical) {
          // Change the meaning of the button as this action is different than the default one.
          removeBtn.textContent = "Remove from Timeseries";
        }

        removeBtn.addEventListener("click", async () => {
          // log UI event on server
          appActions.logUserEvent(this.$store, {
            feature: Feature.REMOVE_FEATURE,
            activity: Activity.DATA_PREPARATION,
            subActivity: SubActivity.DATA_TRANSFORMATION,
            details: { feature: group.key },
          });

          const dataset = routeGetters.getRouteDataset(this.$store);
          const targetName = routeGetters.getRouteTargetVariable(this.$store);

          // get an updated view of the training data list
          const training = routeGetters.getDecodedTrainingVariableNames(
            this.$store
          );
          training.splice(training.indexOf(group.key), 1);

          // update task based on the current training data
          const taskResponse = await datasetActions.fetchTask(this.$store, {
            dataset,
            targetName,
            variableNames: training,
          });

          // update route with training data
          const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
            training: training.join(","),
            task: taskResponse.data.task.join(","),
          });

          if (this.isTimeseries && isCategorical) {
            // Fetch the information of the timeseries grouping
            const currentGrouping = datasetGetters
              .getGroupings(this.$store)
              .find((v) => v.key === targetName)?.grouping;

            // Simply duplicate its grouping information and remove the series ID
            const grouping = JSON.parse(JSON.stringify(currentGrouping));
            grouping.subIds = grouping.subIds.filter(
              (subId) => subId !== group.key
            );
            grouping.idCol = getComposedVariableKey(grouping.subIds);

            // Request to update the timeseries grouping without this series ID
            await datasetActions.updateGrouping(this.$store, {
              variable: targetName,
              grouping,
            });
          }

          this.$router.push(entry).catch((err) => console.warn(err));
          removeFiltersByName(this.$router, group.key);
        });

        return removeBtn;
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
      this.$router.push(entry).catch((err) => console.warn(err));
    },
  },
});
</script>

<style scoped>
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
