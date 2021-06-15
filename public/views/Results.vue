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
  <div class="container-fluid d-flex flex-column h-100 results-view">
    <div class="row flex-0-nav"></div>
    <div class="row align-items-center justify-content-left bg-white">
      <h5 class="header-label">
        Check Models: Review results to understand model performance
      </h5>
    </div>

    <div class="row flex-1 pb-3">
      <div
        class="variable-summaries col-12 col-md-3 border-gray-right results-variable-summaries h-100"
      >
        <p class="nav-link font-weight-bold">Feature Summaries</p>
        <variable-facets
          class="h-100"
          enable-search
          enable-highlighting
          enable-color-scales
          :facetCount="trainingVariables.length"
          instance-name="resultTrainingVars"
          is-result-features
          :log-activity="logActivity"
          model-selection
          :pagination="trainingVariables.length > rowsPerPage"
          :summaries="trainingSummariesByImportance"
          :rows-per-page="rowsPerPage"
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
import ResultsComparison from "../components/ResultsComparison.vue";
import ResultSummaries from "../components/ResultSummaries.vue";
import { Variable, VariableSummary } from "../store/dataset/index";
import { actions as viewActions } from "../store/view/module";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";
import { getters as resultGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import {
  NUM_PER_PAGE,
  getVariableSummariesByState,
  searchVariables,
  filterArrayByPage,
  shouldRunMi,
  getAllVariablesSummaries,
} from "../util/data";
import { Activity } from "../util/userEvents";
import { isGeoLocatedType } from "../util/types";
import { overlayRouteEntry } from "../util/routes";

export default Vue.extend({
  name: "results-view",

  components: {
    VariableFacets,
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
    geoVarExists(): boolean {
      const varSums = getAllVariablesSummaries(
        requestGetters.getActiveSolutionTrainingVariables(this.$store),
        resultGetters.getTrainingSummariesDictionary(this.$store)
      );
      return varSums.some((v) => {
        return isGeoLocatedType(v.type);
      });
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
      const summaryDictionary = resultGetters.getTrainingSummariesDictionary(
        this.$store
      );
      const trainingSummaries = getVariableSummariesByState(
        this.resultTrainingVarsPage,
        this.trainingVariables.length,
        this.trainingVariables,
        summaryDictionary,
        true
      );

      return filterArrayByPage(
        this.resultTrainingVarsPage,
        this.rowsPerPage,
        trainingSummaries
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
    openSolutions(): string[] {
      return routeGetters.getRouteOpenSolutions(this.$store);
    },
  },

  async beforeMount() {
    await viewActions.fetchResultsData(this.$store);
    viewActions.updateResultBaseline(this.$store);
  },
  mounted() {
    this.runMi();
  },
  methods: {
    // checks if MI should be available, if it should and isnt run MI
    async runMi() {
      if (shouldRunMi(this.dataset)) {
        await datasetActions.fetchVariables(this.$store, {
          dataset: this.dataset,
        });
        await datasetActions.fetchVariableRankings(this.$store, {
          dataset: this.dataset,
          target: this.target,
        });
      }
    },
  },
  watch: {
    openSolutions(requestIds: string[]) {
      viewActions.updateResultSummaries(this.$store, { requestIds });
    },
    geoVarExists() {
      const route = routeGetters.getRoute(this.$store);
      const entry = overlayRouteEntry(route, { hasGeoData: this.geoVarExists });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
    highlightString() {
      viewActions.updateResultsSolution(this.$store);
    },
    solutionId() {
      viewActions.updateResultsSolution(this.$store);
      viewActions.updateResultBaseline(this.$store);
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
