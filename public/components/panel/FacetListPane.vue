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
  <variable-facets
    enable-highlighting
    enable-search
    enable-type-change
    enable-type-filtering
    ignore-highlights
    :is-available-features="isSelectedView"
    :is-result-features="isResultView"
    :include="include"
    :enable-color-scales="enableColorScales"
    :facet-count="searchedActiveVariables.length"
    :html="buttons"
    :instance-name="instanceName"
    :log-activity="problemDefinition"
    :rows-per-page="numRowsPerPage"
    :pagination="
      searchedActiveVariables && searchedActiveVariables.length > numRowsPerPage
    "
    :summaries="activeSummaries"
    @search="onSearch"
  />
</template>

<script lang="ts">
import Vue from "vue";
import { isNil } from "lodash";

import VariableFacets from "../../components/facets/VariableFacets.vue";

import {
  SummaryMode,
  TaskTypes,
  Variable,
  VariableSummary,
} from "../../store/dataset/index";
import { actions as appActions } from "../../store/app/module";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../../store/dataset/module";
import {
  DATA_EXPLORER_VAR_INSTANCE,
  ROUTE_PAGE_SUFFIX,
} from "../../store/route/index";
import { getters as routeGetters } from "../../store/route/module";
import { actions as viewActions } from "../../store/view/module";

import { NUM_PER_PAGE, searchVariables } from "../../util/data";
import { Group } from "../../util/facets";
import {
  getRouteFacetPage,
  overlayRouteEntry,
  RouteArgs,
  varModesToString,
} from "../../util/routes";
import { Feature, Activity, SubActivity } from "../../util/userEvents";
import { isUnsupportedTargetVar } from "../../util/types";
import { ExplorerStateNames } from "../../util/dataExplorer";

export default Vue.extend({
  name: "FacetListPane",

  components: {
    VariableFacets,
  },

  props: {
    variables: {
      type: Array as () => Variable[],
      default: () => [] as Variable[],
    },
    summaries: {
      type: Array as () => VariableSummary[],
      default: () => [] as VariableSummary[],
    },
    enableColorScales: { type: Boolean as () => boolean, default: false },
    include: { type: Boolean as () => boolean, default: true },
    enableFooter: { type: Boolean as () => boolean, default: false },
  },

  data() {
    return {
      instanceName: DATA_EXPLORER_VAR_INSTANCE,
      numRowsPerPage: NUM_PER_PAGE,
      search: "",
    };
  },

  computed: {
    isSelectedView(): boolean {
      return (
        routeGetters.getDataExplorerState(this.$store) ===
        ExplorerStateNames.SELECT_VIEW
      );
    },
    isResultView(): boolean {
      return (
        routeGetters.getDataExplorerState(this.$store) ===
        ExplorerStateNames.RESULT_VIEW
      );
    },
    buttons(): (group: Group) => HTMLElement | null {
      return !this.enableFooter
        ? null
        : (group: Group) => {
            const variable = group.key;

            // Display and Hide variables in the Data Explorer.
            const exploreButton = this.displayButton(variable);

            // Add/Remove a variable as training.
            const trainingButton = this.trainingButton(variable);

            // Add/Remove a variable as target.
            const targetButton = this.targetButton(variable);

            // List of model creation buttons to be added.
            const buttons = [targetButton, trainingButton].filter((b) => !!b);
            const modelButtons = document.createElement("div");
            modelButtons.className = "btn-group ml-auto";
            modelButtons.append(...buttons);

            // Container to display the buttons with flex.
            const container = document.createElement("div");
            container.className = "d-flex justify-content-between w-100";
            container.append(exploreButton, modelButtons);
            return container;
          };
    },
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    explore(): string[] {
      return routeGetters.getExploreVariables(this.$store);
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

    problemDefinition(): string {
      return Activity.PROBLEM_DEFINITION;
    },

    searchedActiveVariables(): Variable[] {
      // remove variables used in groupedFeature;
      const activeVariables = this.variables.filter(
        (v) => !this.groupedFeatures.includes(v.key)
      );
      return searchVariables(activeVariables, this.search);
    },

    activeSummaries(): VariableSummary[] {
      const searchedMap = new Map(
        this.searchedActiveVariables.map((v) => {
          return [v.key, true];
        })
      );
      const pageId = DATA_EXPLORER_VAR_INSTANCE + ROUTE_PAGE_SUFFIX;
      const page = getRouteFacetPage(pageId, this.$route);
      const begin = (page - 1) * this.numRowsPerPage;
      const currentSummaries = this.summaries.filter((s) => {
        return searchedMap.has(s.key);
      });
      const end = Math.min(page * this.numRowsPerPage, currentSummaries.length);
      return currentSummaries.slice(begin, end);
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    training(): string[] {
      return routeGetters.getDecodedTrainingVariableNames(this.$store);
    },

    unsupportedTargets(): Set<string> {
      return new Set(
        this.variables
          .filter((v) => isUnsupportedTargetVar(v.key, v.colType))
          .map((v) => v.key)
      );
    },

    varsPage(): number {
      return routeGetters.getRouteDataExplorerVarsPage(this.$store);
    },

    varsSearch(): string {
      return routeGetters.getRouteDataExplorerVarsSearch(this.$store);
    },
  },

  watch: {
    varsPage() {
      viewActions.fetchDataExplorerData(this.$store, this.variables);
    },

    varsSearch() {
      viewActions.fetchDataExplorerData(this.$store, this.variables);
    },
  },

  methods: {
    onSearch(term): void {
      this.search = term;
    },

    displayButton(variable: string): HTMLElement {
      const isInExplore = this.isExplore(variable);
      const button = document.createElement("button");
      button.className = "btn btn-sm";
      button.className += isInExplore ? " btn-outline-primary" : " btn-primary";
      button.textContent = isInExplore ? "Hide" : "Display";
      button.addEventListener("click", () => this.updateExplore(variable));
      return button;
    },

    trainingButton(variable: string): HTMLElement {
      // To do not allow selection as training if the variable is the target.
      if (this.isTarget(variable)) return;

      const isTraining = this.isTraining(variable);

      const button = document.createElement("button");
      button.className = "btn btn-sm";
      button.className += isTraining
        ? " btn-outline-secondary"
        : " btn-secondary";
      button.textContent = isTraining ? "Remove Training" : "Select Training";
      button.addEventListener("click", () => this.updateTraining(variable));
      return button;
    },

    targetButton(variable: string): HTMLElement {
      // Check if the variable is an unsupported to be a target.
      const isUnsupported = this.unsupportedTargets.has(variable);
      if (isUnsupported) return;

      const isTarget = this.isTarget(variable);

      // Only display the button if no target has been selected,
      // or this variable is the target and needs to be unselected.
      if (!isNil(this.target) && !isTarget) return;

      const button = document.createElement("button");
      button.className = "btn btn-sm";
      button.className += isTarget
        ? " btn-outline-secondary"
        : " btn-secondary";
      button.textContent = isTarget ? "Remove Target" : "Select Target";
      button.addEventListener("click", () => this.updateTarget(variable));
      return button;
    },

    updateRoute(args: RouteArgs) {
      const entry = overlayRouteEntry(this.$route, args);
      this.$router.push(entry).catch((err) => console.warn(err));
    },

    isTarget(variable: string): boolean {
      return variable === this.target;
    },

    isTraining(variable: string): boolean {
      return this.training?.includes(variable) ?? false;
    },

    isExplore(variable: string): boolean {
      return this.explore.includes(variable);
    },

    async updateTarget(target: string): Promise<void> {
      // Is the variable the current target?
      if (this.isTarget(target)) {
        // Remove the variable as target
        this.updateRoute({ target: null, task: null });
        return;
      }

      const args = {} as RouteArgs;
      args.target = target;

      // Filter it out of the training
      const training = this.training.filter((v) => v !== target);

      // Get Variables Grouping and check if our target is one of them
      const groupings = datasetGetters.getGroupings(this.$store);
      const targetGrouping = groupings?.find((g) => g.key === target)?.grouping;
      if (!!targetGrouping) {
        if (targetGrouping.subIds.length > 0) {
          targetGrouping.subIds.forEach((subId) => {
            if (!training.find((t) => t === subId)) {
              training.push(subId);
            }
          });
        } else {
          if (!training.find((t) => t === targetGrouping.idCol)) {
            training.push(targetGrouping.idCol);
          }
        }
      }

      // Get the var modes
      const varModesMap = routeGetters.getDecodedVarModes(this.$store);
      args.varModes = varModesToString(varModesMap);

      // Fetch the task
      try {
        const response = await datasetActions.fetchTask(this.$store, {
          dataset: this.dataset,
          targetName: target,
          variableNames: [],
        });
        args.task = response.data.task.join(",") ?? "";

        // Update the training variable
        if (args.task.includes("timeseries")) {
          training.forEach((variable) => {
            if (variable !== target) {
              varModesMap.set(variable, SummaryMode.Timeseries);
            }
          });
        }
      } catch (error) {
        console.log(error);
      }

      // Make the list of training variables' name a string.
      args.training = training.join(",");

      appActions.logUserEvent(this.$store, {
        feature: Feature.SELECT_TARGET,
        activity: Activity.PROBLEM_DEFINITION,
        subActivity: SubActivity.PROBLEM_SPECIFICATION,
        details: { target },
      });

      this.updateRoute(args);

      datasetActions.fetchVariableRankings(this.$store, {
        dataset: this.dataset,
        target,
      });
    },

    async updateTraining(variable: string): Promise<void> {
      const args = {} as RouteArgs;
      if (this.isTraining(variable)) {
        args.training = this.training.filter((v) => v !== variable).join(",");
      } else {
        const training = this.training.concat([variable]);
        args.training = training.join(",");
        const taskResponse = await datasetActions.fetchTask(this.$store, {
          dataset: this.dataset,
          targetName: this.target,
          variableNames: training,
        });
        const task = taskResponse.data.task.join(",");
        args.task = task;
        if (task.includes(TaskTypes.REMOTE_SENSING)) {
          const available = routeGetters.getAvailableVariables(this.$store);
          const varModesMap = routeGetters.getDecodedVarModes(this.$store);
          training.forEach((v) => {
            varModesMap.set(v, SummaryMode.MultiBandImage);
          });

          available.forEach((v) => {
            varModesMap.set(v.key, SummaryMode.MultiBandImage);
          });

          varModesMap.set(
            routeGetters.getRouteTargetVariable(this.$store),
            SummaryMode.MultiBandImage
          );
          const varModesStr = varModesToString(varModesMap);
          args.varModes = varModesStr;
        }
      }

      this.updateRoute(args);
    },

    updateExplore(variable: string): void {
      const args = {} as RouteArgs;
      if (this.isExplore(variable)) {
        args.explore = this.explore.filter((v) => v !== variable).join(",");
      } else {
        args.explore = this.explore.concat([variable]).join(",");
      }
      this.updateRoute(args);
    },
  },
});
</script>
