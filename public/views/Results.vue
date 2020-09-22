<template>
  <div class="container-fluid d-flex flex-column h-100 results-view">
    <div class="row flex-0-nav"></div>
    <div class="row align-items-center justify-content-left bg-white">
      <h5 class="header-label">
        Check Models: Review results to understand model performance
      </h5>
    </div>

    <div class="row flex-1 pb-3">
      <div
        class="variable-summaries col-12 col-md-3 border-gray-right results-variable-summaries"
      >
        <p class="nav-link font-weight-bold">Feature Summaries</p>
        <variable-facets
          class="h-100"
          enable-search
          enable-highlighting
          :facetCount="trainingVariables.length"
          instance-name="resultTrainingVars"
          is-result-features
          :log-activity="logActivity"
          model-selection
          :pagination="trainingVariables.length > rowsPerPage"
          :summaries="trainingSummariesByImportance"
        >
        </variable-facets>
      </div>

      <results-comparison
        class="col-12 col-md-6 results-result-comparison"
      ></results-comparison>
      <result-summaries
        class="col-12 col-md-3 border-gray-left results-result-summaries"
      ></result-summaries>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import VariableFacets from "../components/facets/VariableFacets.vue";
import ResultsComparison from "../components/ResultsComparison";
import ResultSummaries from "../components/ResultSummaries";
import ResultTargetVariable from "../components/ResultTargetVariable";
import { Variable, VariableSummary } from "../store/dataset/index";
import { actions as viewActions } from "../store/view/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as resultGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import {
  NUM_PER_PAGE,
  getVariableSummariesByState,
  searchVariables,
  sortSolutionSummariesByImportance,
} from "../util/data";
import { Feature, Activity } from "../util/userEvents";

export default Vue.extend({
  name: "results-view",

  components: {
    VariableFacets,
    ResultTargetVariable,
    ResultsComparison,
    ResultSummaries,
  },

  data() {
    return {
      logActivity: Activity.MODEL_SELECTION,
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },
    // Always use the label from the target summary facet as the displayed target name to ensure compund
    // variables like a time series display the same name for the target as the value being predicted.
    targetLabel(): string {
      const summary = resultGetters.getTargetSummary(this.$store);
      if (summary !== null) {
        return summary.label;
      }
      return this.target;
    },
    targetType(): string {
      const variables = datasetGetters.getVariablesMap(this.$store);
      if (variables && variables[this.target]) {
        return variables[this.target].colType;
      }
      return "";
    },
    resultTrainingVarsSearch(): string {
      return routeGetters.getRouteResultTrainingVarsSearch(this.$store);
    },
    trainingVariables(): Variable[] {
      return searchVariables(
        requestGetters.getActiveSolutionTrainingVariables(this.$store),
        this.resultTrainingVarsSearch
      );
    },
    trainingSummaries(): VariableSummary[] {
      const summaryDictionary = resultGetters.getTrainingSummariesDictionary(
        this.$store
      );

      return getVariableSummariesByState(
        this.resultTrainingVarsPage,
        this.rowsPerPage,
        this.trainingVariables,
        summaryDictionary
      );
    },
    trainingSummariesByImportance(): VariableSummary[] {
      return sortSolutionSummariesByImportance(
        this.trainingSummaries,
        this.trainingVariables,
        this.solutionId
      );
    },
    solutionId(): string {
      return routeGetters.getRouteSolutionId(this.$store);
    },
    highlightString(): string {
      return routeGetters.getRouteHighlight(this.$store);
    },
    resultTrainingVarsPage(): number {
      return routeGetters.getRouteResultTrainingVarsPage(this.$store);
    },
    rowsPerPage(): number {
      return NUM_PER_PAGE;
    },
  },

  beforeMount() {
    viewActions.fetchResultsData(this.$store);
  },

  watch: {
    highlightString() {
      viewActions.updateResultsSolution(this.$store);
    },
    solutionId() {
      viewActions.updateResultsSolution(this.$store);
    },
    resultTrainingVarsPage() {
      viewActions.updateResultsSummaries(this.$store);
    },
    resultTrainingVarsSearch() {
      viewActions.updateResultsSummaries(this.$store);
    },
  },
});
</script>

<style>
.variable-summaries {
  display: flex;
  flex-direction: column;
}
.variable-summaries .facets-group {
  /* for the spinners, this isn't needed on other views because of the buttoms that create the space */
  padding-bottom: 20px;
}
.results-view .nav-link {
  padding: 1rem 0 0.25rem 0;
  border-bottom: 1px solid #e0e0e0;
  color: rgba(0, 0, 0, 0.87);
}
.header-label {
  padding: 1rem 0 0.5rem 0;
  margin-left: 200px;
  font-weight: bold;
}
.results-view .table td {
  text-align: left;
  padding: 0px;
}
.results-view .table td > div {
  text-align: left;
  padding: 0.3rem;
  overflow: hidden;
  text-overflow: ellipsis;
  min-height: 1.875rem;
}
.result-facets {
  margin-bottom: 12px;
}
.results-variable-summaries,
.results-result-comparison,
.results-result-summaries,
.results-variable-summaries /deep/ .variable-facets-wrapper {
  height: 100%;
}
@media (max-width: 767px) {
  .results-variable-summaries,
  .results-result-comparison,
  .results-result-summaries {
    height: unset;
  }
}
</style>
