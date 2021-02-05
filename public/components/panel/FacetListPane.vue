<template>
  <variable-facets
    enable-highlighting
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
    @search="onSearch"
  />
</template>

<script lang="ts">
import Vue from "vue";

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
import { overlayRouteEntry } from "../../util/routes";
import { Activity } from "../../util/userEvents";

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

    button(): (group: Group) => HTMLElement {
      return (group: Group) => {
        const variable = group.key;
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

        const onClick = () => {
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
          const args = {
            training: updatedTraining.join(","),
            task,
          };
          const entry = overlayRouteEntry(this.$route, args);
          this.$router.push(entry).catch((err) => console.warn(err));
          viewActions.updateSelectTrainingData(this.$store);
        };

        // create a button
        button.addEventListener("click", onClick);
        return button;
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
  },
});
</script>
