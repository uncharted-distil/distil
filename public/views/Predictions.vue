<template>
  <div class="predictions-data-view d-flex h-100">
    <status-panel></status-panel>
    <div class="sidebar-container d-flex flex-column h-100">
      <div class="padding-nav"></div>
      <status-sidebar></status-sidebar>
    </div>
    <div class="container-fluid d-flex flex-column h-100 predictions-view">
      <div class="row flex-0-nav"></div>
      <div class="row flex-1 pb-3">
        <div
          class="variable-summaries col-12 col-md-3 border-gray-right predictions-variable-summaries"
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

        <predictions-data-slot
          class="col-12 col-md-6 d-flex flex-column predictions-predictions-data"
        ></predictions-data-slot>

        <prediction-summaries
          class="col-12 col-md-3 border-gray-left predictions-predictions-summaries"
        ></prediction-summaries>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import VariableFacets from "../components/facets/VariableFacets.vue";
import PredictionsDataSlot from "../components/PredictionsDataSlot";
import PredictionSummaries from "../components/PredictionSummaries";
import StatusPanel from "../components/StatusPanel";
import StatusSidebar from "../components/StatusSidebar";
import { VariableSummary } from "../store/dataset/index";
import { actions as viewActions } from "../store/view/module";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";
import { getters as resultGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as predictionGetters } from "../store/predictions/module";
import { Feature, Activity } from "../util/userEvents";

export default Vue.extend({
  name: "predictions-view",

  components: {
    VariableFacets,
    PredictionsDataSlot,
    PredictionSummaries,
    StatusPanel,
    StatusSidebar,
  },

  data() {
    return {
      logActivity: Activity.PREDICTION_ANALYSIS,
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    trainingSummaries(): VariableSummary[] {
      return predictionGetters.getTrainingSummaries(this.$store);
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
  },

  beforeMount() {
    viewActions.fetchPredictionsData(this.$store);
    datasetActions.fetchClusters(this.$store, { dataset: this.dataset });
  },

  watch: {
    produceRequestId() {
      viewActions.updatePrediction(this.$store);
    },
    highlightString() {
      viewActions.updatePrediction(this.$store);
    },
    trainingVarsPage() {
      viewActions.updatePrediction(this.$store);
    },
  },
});
</script>

<style>
.predictions-data-view {
  flex-direction: row-reverse;
}

.predictions-view .nav-link {
  padding: 1rem 0 0.25rem 0;
  border-bottom: 1px solid #e0e0e0;
  color: rgba(0, 0, 0, 0.87);
}

.predictions-view .table td {
  text-align: left;
  padding: 0px;
}
.predictions-view .table td > div {
  text-align: left;
  padding: 0.3rem;
  overflow: hidden;
  text-overflow: ellipsis;
  min-height: 1.875rem;
}

.variable-summaries {
  display: flex;
  flex-direction: column;
}

.variable-summaries .facets-group {
  /* for the spinners, this isn't needed on other views because of the buttoms that create the space */
  padding-bottom: 20px;
}

.predictions-variable-summaries,
.predictions-predictions-data,
.predictions-predictions-summaries {
  height: 100%;
}
@media (max-width: 767px) {
  .predictions-variable-summaries,
  .predictions-predictions-data,
  .predictions-predictions-summaries {
    height: unset;
  }
}
</style>
