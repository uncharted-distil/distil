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
  <div class="select-training-view d-flex h-100">
    <!-- Status Panel and Sidebar -->
    <status-panel />
    <div class="sidebar-container d-flex flex-column h-100">
      <div class="padding-nav" />
      <status-sidebar />
    </div>

    <!-- Main -->
    <div class="container-fluid d-flex flex-column h-100 select-view">
      <!-- Spacer for the navigation. -->
      <div class="row flex-0-nav" />

      <!-- Header -->
      <header class="header row align-items-center justify-content-center">
        <div class="col-12 col-md-6 d-flex flex-column">
          <h5 class="header-title">
            Select Features That May Predict
            <strong>{{ targetLabel.toUpperCase() }}</strong>
          </h5>
          <p>
            Use interactive feature highlighting to analyze relationships or to
            exclude samples from the model. Features which appear to have
            stronger relation are listed&nbsp;first.
          </p>
        </div>

        <div class="col-12 col-md-6 d-flex flex-column">
          <div class="select-target-variables">
            <target-variable class="col-12 d-flex flex-column" />
          </div>
        </div>
      </header>

      <!-- Content -->
      <div class="row flex-1 pb-3">
        <available-training-variables
          is-available-features
          class="col-12 col-md-3 d-flex h-100"
          :variables="availableVariables"
          :summaries="availableSummaries"
          :instance-name="availableInstance"
          title="Available Features"
          subtitle="features available"
          group-btn-title="Add All"
          btn-title="Add"
          @var-change="addVar"
          @group-change="addAll"
        />
        <available-training-variables
          check-geo-type
          class="col-12 col-md-3 nopadding d-flex h-100"
          title="Features to Model"
          group-btn-title="Remove All"
          btn-title="Remove"
          subtitle="features selected"
          :variables="trainingVariables"
          :summaries="trainingVariableSummaries"
          :instance-name="trainingInstance"
          @var-change="removeVar"
          @group-change="removeAll"
        />

        <div class="col-12 col-md-6 d-flex flex-column h-100">
          <select-data-slot class="flex-1 d-flex flex-column pb-1 pt-2" />
          <create-solutions-form class="select-create-solutions" />
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { datasetActions, datasetGetters, viewActions } from "../store";
import { getters as routeGetters } from "../store/route/module";
import StatusPanel from "../components/StatusPanel.vue";
import StatusSidebar from "../components/StatusSidebar.vue";
import CreateSolutionsForm from "../components/CreateSolutionsForm.vue";
import SelectDataSlot from "../components/SelectDataSlot.vue";
import AvailableTrainingVariables, {
  GroupChangeParams,
} from "../components/AvailableTrainingVariables.vue";
import { Variable, VariableSummary } from "../store/dataset/index";
import TargetVariable from "../components/TargetVariable.vue";
import { DataMode } from "../store/dataset";
import { overlayRouteEntry } from "../util/routes";
import {
  searchVariables,
  NUM_PER_PAGE,
  getVariableSummariesByState,
  addVariableToTraining,
  removeVariableFromTraining,
} from "../util/data";
import {
  TRAINING_VARS_INSTANCE,
  AVAILABLE_TRAINING_VARS_INSTANCE,
} from "../store/route/index";
import { Dictionary } from "../util/dict";
import { Route } from "vue-router";
import { Group } from "../util/facets";

export default Vue.extend({
  name: "SelectTrainingView",

  components: {
    CreateSolutionsForm,
    SelectDataSlot,
    AvailableTrainingVariables,
    TargetVariable,
    StatusPanel,
    StatusSidebar,
  },

  data() {
    return {
      binningIntervalModel: null,
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    trainingStr(): string {
      return routeGetters.getRouteTrainingVariables(this.$store);
    },
    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },
    // Always use the label from the target summary facet as the displayed target name to ensure compund
    // variables like a time series display the same name for the target as the value being predicted.
    targetLabel(): string {
      const summaries = routeGetters.getTargetVariableSummaries(this.$store);
      if (summaries.length > 0) {
        const summary = summaries[0];
        return summary.label;
      }
      return this.target;
    },
    filters(): string {
      return (
        routeGetters.getRouteHighlight(this.$store) +
        routeGetters.getRouteFilters(this.$store)
      );
    },
    highlightString(): string {
      return routeGetters.getRouteHighlight(this.$store);
    },
    ranking(): boolean {
      return routeGetters.getRouteIsTrainingVariablesRanked(this.$store);
    },
    availableTrainingVarsPage(): number {
      return routeGetters.getRouteAvailableTrainingVarsPage(this.$store);
    },
    trainingVarsPage(): number {
      return routeGetters.getRouteTrainingVarsPage(this.$store);
    },
    availableInstance(): string {
      return AVAILABLE_TRAINING_VARS_INSTANCE;
    },
    availableTrainingVarsSearch(): string {
      return routeGetters.getRouteAvailableTrainingVarsSearch(this.$store);
    },
    trainingInstance(): string {
      return TRAINING_VARS_INSTANCE;
    },
    trainingVarsSearch(): string {
      return routeGetters.getRouteTrainingVarsSearch(this.$store);
    },
    availableSummaries(): VariableSummary[] {
      const pageIndex = routeGetters.getRouteAvailableTrainingVarsPage(
        this.$store
      );
      const currentSummaries = getVariableSummariesByState(
        pageIndex,
        NUM_PER_PAGE,
        this.availableVariables,
        this.variableDict
      );

      return currentSummaries;
    },
    availableVariables(): Variable[] {
      return searchVariables(
        routeGetters.getAvailableVariables(this.$store),
        this.availableTrainingVarsSearch
      );
    },
    include(): boolean {
      return routeGetters.getRouteInclude(this.$store);
    },
    variableDict(): Dictionary<Dictionary<VariableSummary>> {
      return this.include
        ? datasetGetters.getIncludedVariableSummariesDictionary(this.$store)
        : datasetGetters.getExcludedVariableSummariesDictionary(this.$store);
    },
    trainingVariables(): Variable[] {
      return searchVariables(
        routeGetters.getTrainingVariables(this.$store),
        this.trainingVarsSearch
      );
    },
    trainingVariableSummaries(): VariableSummary[] {
      const pageIndex = routeGetters.getRouteTrainingVarsPage(this.$store);
      const currentSummaries = getVariableSummariesByState(
        pageIndex,
        NUM_PER_PAGE,
        this.trainingVariables,
        this.variableDict
      );

      return currentSummaries;
    },
  },
  watch: {
    trainingStr() {
      viewActions.updateSelectTrainingData(this.$store);
    },
    filters() {
      viewActions.updateSelectTrainingData(this.$store);
    },
    availableTrainingVarsPage() {
      viewActions.updateSelectVariables(this.$store);
    },
    trainingVarsPage() {
      viewActions.updateSelectVariables(this.$store);
    },
    availableTrainingVarsSearch() {
      viewActions.updateSelectVariables(this.$store);
    },
    trainingVarsSearch() {
      viewActions.updateSelectVariables(this.$store);
    },
    dataset() {
      viewActions.fetchSelectTrainingData(this.$store, true);
    },
    ranking() {
      viewActions.updateSelectTrainingData(this.$store);
    },

    $route(to: Route, from: Route) {
      const dataModeOld = from.query.dataMode as string;
      if (
        routeGetters.getDataMode(this.$store) === DataMode.Cluster &&
        dataModeOld !== DataMode.Cluster
      ) {
        viewActions.updateSelectTrainingData(this.$store);
        const clusterEntry = overlayRouteEntry(this.$route, {
          clustering: "1",
        });
        this.$router.push(clusterEntry).catch((err) => console.warn(err));
      }
    },
  },

  beforeMount() {
    viewActions.fetchSelectTrainingData(this.$store, false);
    viewActions.updateHighlight(this.$store);
  },
  methods: {
    async addVar(group: Group) {
      this.$router;
      addVariableToTraining(group, this.$router);
    },
    async addAll(params: GroupChangeParams) {
      // update task based on the current training data
      const taskResponse = await datasetActions.fetchTask(this.$store, {
        dataset: params.dataset,
        targetName: params.targetName,
        variableNames: params.variableNames,
      });
      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        training: params.variableNames.join(","),
        availableTrainingVarsPage: 1,
        task: taskResponse.data.task.join(","),
      });

      this.$router.push(entry).catch((err) => console.warn(err));
    },
    async removeVar(group: Group) {
      removeVariableFromTraining(group, this.$router);
    },
    async removeAll() {
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

<style>
.select-training-view {
  flex-direction: row-reverse;
}
.select-target-variables {
  min-width: 500px;
  max-width: 600px !important;
  align-self: flex-end;
}
.select-target-variables .facet-sparkline-container {
  height: 30px !important;
}
</style>
