<template>
  <div class="container-fluid d-flex flex-column h-100 results-view">
    <div class="row flex-0-nav"></div>
    <div class="row flex-1 pb-3">
      <div
        class="variable-summaries col-12 col-md-3 border-gray-right results-variable-summaries"
      >
        <p class="nav-link font-weight-bold">Feature Summaries</p>
        <variable-facets
          class="h-100"
          enable-search
          enable-highlighting
          model-selection
          instance-name="resultTrainingVars"
          :summaries="trainingSummaries"
          :log-activity="logActivity"
        >
        </variable-facets>
      </div>

      <results-comparison
        class="col-12 col-md-6 results-result-comparison"
      ></results-comparison>
      <result-summaries
        class="col-12 col-md-3 border-gray-left results-result-summaries"
        :isPrediction="true"
      ></result-summaries>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import VariableFacets from "../components/facets/VariableFacets.vue";
import ResultsComparison from "../components/ResultsComparison";
import ResultSummaries from "../components/ResultSummaries";
import { VariableSummary } from "../store/dataset/index";
import { actions as viewActions } from "../store/view/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as resultGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { Feature, Activity } from "../util/userEvents";

export default Vue.extend({
  name: "results-view",

  components: {
    VariableFacets,
    ResultsComparison,
    ResultSummaries
  },

  data() {
    return {
      logActivity: Activity.PREDICTION_ANALYSIS
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    trainingSummaries(): VariableSummary[] {
      return resultGetters.getTrainingSummaries(this.$store);
    },
    solutionId(): string {
      return routeGetters.getRouteSolutionId(this.$store);
    },
    produceRequestId(): string {
      return routeGetters.getRouteProduceRequestId(this.$store);
    },
    highlightString(): string {
      return routeGetters.getRouteHighlight(this.$store);
    },
    resultTrainingVarsPage(): number {
      return routeGetters.getRouteResultTrainingVarsPage(this.$store);
    }
  },

  beforeMount() {
    viewActions.fetchPredictionsData(this.$store);
  },

  watch: {
    highlightString() {
      viewActions.updatePrediction(this.$store);
    },
    solutionId() {
      viewActions.updatePrediction(this.$store);
    },
    produceRequestId() {
      viewActions.updatePrediction(this.$store);
    },
    resultTrainingVarsPage() {
      viewActions.updatePrediction(this.$store);
    }
  }
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
}
.result-facets {
  margin-bottom: 12px;
}
.results-variable-summaries,
.results-result-comparison,
.results-result-summaries {
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
