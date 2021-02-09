<template>
  <variable-facets
    enable-highlighting
    enable-search
    enable-type-change
    enable-type-filtering
    ignore-highlights
    :facet-count="searchedActiveVariables.length"
    :html="buttons"
    :instance-name="instanceName"
    :log-activity="problemDefinition"
    :rows-per-page="numRowsPerPage"
    :summaries="summaries"
    @search="onSearch"
  />
</template>

<script lang="ts">
import Vue from "vue";
import { isNil } from "lodash";

import VariableFacets from "../../components/facets/VariableFacets.vue";

import { Variable, VariableSummary } from "../../store/dataset/index";
import { getters as datasetGetters } from "../../store/dataset/module";
import { DATA_EXPLORER_VAR_INSTANCE } from "../../store/route/index";
import { getters as routeGetters } from "../../store/route/module";
import { actions as viewActions } from "../../store/view/module";

import {
  getVariableSummariesByState,
  NUM_PER_PAGE,
  searchVariables,
} from "../../util/data";
import { Group } from "../../util/facets";
import { overlayRouteEntry, RouteArgs } from "../../util/routes";
import { Activity } from "../../util/userEvents";
import { isUnsupportedTargetVar } from "../../util/types";

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
  },

  data() {
    return {
      instanceName: DATA_EXPLORER_VAR_INSTANCE,
      numRowsPerPage: NUM_PER_PAGE,
      search: "",
    };
  },

  computed: {
    varsPage(): number {
      return routeGetters.getRouteDataExplorerVarsPage(this.$store);
    },

    varsSearch(): string {
      return routeGetters.getRouteDataExplorerVarsSearch(this.$store);
    },

    buttons(): (group: Group) => HTMLElement {
      return (group: Group) => {
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
        container.className = "d-flex";
        container.append(exploreButton, modelButtons);
        return container;
      };
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

    summaries(): VariableSummary[] {
      const summaryDictionary = datasetGetters.getVariableSummariesDictionary(
        this.$store
      );

      const currentSummaries = getVariableSummariesByState(
        this.varsPage,
        this.numRowsPerPage,
        this.searchedActiveVariables,
        summaryDictionary
      );

      return currentSummaries;
    },

    explore(): string[] {
      return routeGetters.getExploreVariables(this.$store);
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

    updateTarget(variable: string): void {
      const args = {} as RouteArgs;
      // Is the variable the current target?
      if (this.isTarget(variable)) {
        // Remove the variable as target
        args.target = "";
      } else {
        // Or select it as target, but remove it from training
        args.target = variable;
        args.training = this.training.filter((v) => v !== variable).join(",");
      }
      this.updateRoute(args);
    },

    updateTraining(variable: string): void {
      const args = {} as RouteArgs;
      if (this.isTraining(variable)) {
        args.training = this.training.filter((v) => v !== variable).join(",");
      } else {
        args.training = this.training.concat([variable]).join(",");
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
