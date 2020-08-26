<template>
  <div
    class="available-training-variables"
    v-bind:class="{ included: includedActive, excluded: !includedActive }"
  >
    <p class="nav-link font-weight-bold">
      Available Features
      <i class="float-right fa fa-angle-right fa-lg"></i>
    </p>
    <variable-facets
      ref="facets"
      enable-highlighting
      enable-search
      enable-type-change
      :instance-name="instanceName"
      :rows-per-page="numRowsPerPage"
      :summaries="availableVariableSummaries"
      :html="html"
      :isAvailableFeatures="true"
      :isFeaturesToModel="false"
      :pagination="variables.length > numRowsPerPage"
    >
      <div class="available-variables-menu">
        <div>
          {{ subtitle }}
        </div>
        <div v-if="availableVariableSummaries.length > 0">
          <b-button size="sm" variant="outline-secondary" @click="addAll"
            >Add All</b-button
          >
        </div>
      </div>
    </variable-facets>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { overlayRouteEntry } from "../util/routes";
import { Variable, VariableSummary, Task } from "../store/dataset/index";
import {
  actions as datasetActions,
  getters as datasetGetters,
} from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { filterSummariesByDataset, NUM_PER_PAGE } from "../util/data";
import { AVAILABLE_TRAINING_VARS_INSTANCE } from "../store/route/index";
import { Group } from "../util/facets";
import VariableFacets from "./facets/VariableFacets.vue";
import { Dictionary } from "vue-router/types/router";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";

export default Vue.extend({
  name: "available-training-variables",

  components: {
    VariableFacets,
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    includedActive(): boolean {
      return routeGetters.getRouteInclude(this.$store);
    },
    availableVariableSummaries(): VariableSummary[] {
      const summaries = routeGetters.getAvailableVariableSummaries(this.$store);
      return filterSummariesByDataset(summaries, this.dataset);
    },
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
    subtitle(): string {
      return `${this.availableVariableSummaries.length} features available`;
    },
    numRowsPerPage(): number {
      return NUM_PER_PAGE;
    },
    instanceName(): string {
      return AVAILABLE_TRAINING_VARS_INSTANCE;
    },
    html(): (group: Group) => HTMLDivElement {
      return (group: Group) => {
        const container = document.createElement("div");
        const trainingElem = document.createElement("button");
        trainingElem.className += "btn btn-sm btn-outline-secondary mb-2";
        trainingElem.innerHTML = "Add";
        trainingElem.addEventListener("click", async () => {
          // log UI event on server
          appActions.logUserEvent(this.$store, {
            feature: Feature.ADD_FEATURE,
            activity: Activity.DATA_PREPARATION,
            subActivity: SubActivity.DATA_TRANSFORMATION,
            details: { feature: group.colName },
          });

          // get an updated view of the training data list
          const training = routeGetters
            .getDecodedTrainingVariableNames(this.$store)
            .concat([group.colName]);

          // update task based on the current training data
          const taskResponse = await datasetActions.fetchTask(this.$store, {
            dataset: routeGetters.getRouteDataset(this.$store),
            targetName: routeGetters.getRouteTargetVariable(this.$store),
            variableNames: training,
          });

          // update route with training data
          const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
            training: training.join(","),
            task: taskResponse.data.task.join(","),
          });
          this.$router.push(entry);
        });
        container.appendChild(trainingElem);
        return container;
      };
    },
  },

  methods: {
    addAll() {
      // log UI event on server
      appActions.logUserEvent(this.$store, {
        feature: Feature.ADD_ALL_FEATURES,
        activity: Activity.DATA_PREPARATION,
        subActivity: SubActivity.DATA_TRANSFORMATION,
        details: {},
      });

      const facets = this.$refs.facets as any;
      const training = routeGetters.getDecodedTrainingVariableNames(
        this.$store,
      );
      facets.availableVariables().forEach((variable) => {
        training.push(variable);
      });
      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        training: training.join(","),
      });
      this.$router.push(entry);
    },
  },
});
</script>

<style>
.available-training-variables {
  display: flex;
  flex-direction: column;
}
.available-variables-menu {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
  line-height: 30px;
}

.available-training-variables /deep/ .variable-facets-wrapper {
  height: 100%;
}
</style>
