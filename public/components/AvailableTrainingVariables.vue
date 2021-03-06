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
  <div
    class="available-training-variables"
    :class="{ included: includedActive, excluded: !includedActive }"
  >
    <p class="nav-link font-weight-bold">
      Available Features
      <i class="float-right fa fa-angle-right fa-lg" />
    </p>
    <variable-facets
      ref="facets"
      enable-highlighting
      enable-search
      enable-type-change
      :facetCount="availableVariables && availableVariables.length"
      :html="html"
      :isAvailableFeatures="true"
      :isFeaturesToModel="false"
      :instance-name="instanceName"
      :pagination="
        availableVariables && availableVariables.length > numRowsPerPage
      "
      :rows-per-page="numRowsPerPage"
      :summaries="availableVariableSummaries"
    >
      <div
        class="d-flex flex-row justify-content-between align-items-center my-2 mx-1"
      >
        <div>
          {{ subtitle }}
        </div>
        <div v-if="availableVariableSummaries.length > 0">
          <b-button size="sm" variant="outline-secondary" @click="addAll">
            Add All
          </b-button>
        </div>
      </div>
    </variable-facets>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { overlayRouteEntry } from "../util/routes";
import { Variable, VariableSummary } from "../store/dataset/index";
import {
  actions as datasetActions,
  getters as datasetGetters,
} from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import {
  getComposedVariableKey,
  getVariableSummariesByState,
  NUM_PER_PAGE,
  searchVariables,
} from "../util/data";
import { AVAILABLE_TRAINING_VARS_INSTANCE } from "../store/route/index";
import { Group } from "../util/facets";
import VariableFacets from "./facets/VariableFacets.vue";
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

    availableTrainingVarsSearch(): string {
      return routeGetters.getRouteAvailableTrainingVarsSearch(this.$store);
    },

    availableVariableSummaries(): VariableSummary[] {
      const pageIndex = routeGetters.getRouteAvailableTrainingVarsPage(
        this.$store
      );
      const include = routeGetters.getRouteInclude(this.$store);
      const summaryDictionary = include
        ? datasetGetters.getIncludedVariableSummariesDictionary(this.$store)
        : datasetGetters.getExcludedVariableSummariesDictionary(this.$store);

      const currentSummaries = getVariableSummariesByState(
        pageIndex,
        this.numRowsPerPage,
        this.availableVariables,
        summaryDictionary
      );

      return currentSummaries;
    },

    availableVariables(): Variable[] {
      return searchVariables(
        routeGetters.getAvailableVariables(this.$store),
        this.availableTrainingVarsSearch
      );
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    subtitle(): string {
      return `${this.availableVariables.length} features available`;
    },

    numRowsPerPage(): number {
      return NUM_PER_PAGE;
    },

    instanceName(): string {
      return AVAILABLE_TRAINING_VARS_INSTANCE;
    },

    isTimeseries(): boolean {
      return routeGetters.isTimeseries(this.$store);
    },

    targetVariable(): Variable {
      return routeGetters.getTargetVariable(this.$store);
    },

    html(): (group: Group) => HTMLElement {
      return (group: Group) => {
        const variable = group.key;
        const trainingElem = document.createElement("button");
        trainingElem.className += "btn btn-sm btn-outline-secondary mb-2";
        trainingElem.textContent = "Add";

        // In the case of a categorical variable with a timeserie selected.
        const isCategorical: boolean = group.type === "categorical";
        if (this.isTimeseries && isCategorical) {
          // Change the meaning of the button as this action is different than the default one.
          trainingElem.textContent = "Add to Timeseries";
        }

        trainingElem.addEventListener("click", async () => {
          // log UI event on server
          appActions.logUserEvent(this.$store, {
            feature: Feature.ADD_FEATURE,
            activity: Activity.DATA_PREPARATION,
            subActivity: SubActivity.DATA_TRANSFORMATION,
            details: { feature: group.key },
          });

          const dataset = routeGetters.getRouteDataset(this.$store);
          const targetName = routeGetters.getRouteTargetVariable(this.$store);

          // get an updated view of the training data list
          const training = routeGetters
            .getDecodedTrainingVariableNames(this.$store)
            .concat([group.key]);

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

            // Simply duplicate its grouping information and add the new variable
            const grouping = JSON.parse(JSON.stringify(currentGrouping));
            grouping.subIds.push(variable);
            grouping.idCol = getComposedVariableKey(grouping.subIds);

            // Request to update the timeserie grouping
            await datasetActions.updateGrouping(this.$store, {
              variable: targetName,
              grouping,
            });
          }

          this.$router.push(entry).catch((err) => console.warn(err));
        });

        return trainingElem;
      };
    },
  },

  methods: {
    async addAll() {
      // log UI event on server
      appActions.logUserEvent(this.$store, {
        feature: Feature.ADD_ALL_FEATURES,
        activity: Activity.DATA_PREPARATION,
        subActivity: SubActivity.DATA_TRANSFORMATION,
        details: {},
      });

      const training = routeGetters.getDecodedTrainingVariableNames(
        this.$store
      );

      this.availableVariables.forEach((variable) => {
        training.push(variable.key);
      });
      const dataset = routeGetters.getRouteDataset(this.$store);
      const targetName = routeGetters.getRouteTargetVariable(this.$store);
      // update task based on the current training data
      const taskResponse = await datasetActions.fetchTask(this.$store, {
        dataset,
        targetName,
        variableNames: training,
      });
      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        training: training.join(","),
        availableTrainingVarsPage: 1,
        task: taskResponse.data.task.join(","),
      });

      this.$router.push(entry).catch((err) => console.warn(err));
    },
  },
});
</script>

<style scoped>
.available-training-variables {
  display: flex;
  flex-direction: column;
}

.available-training-variables /deep/ .variable-facets-wrapper {
  height: 100%;
}
</style>
