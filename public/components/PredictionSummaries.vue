<template>
  <div class="prediction-summaries">
    <p class="nav-link font-weight-bold">
      Predictions for Model
    </p>
    <div v-for="summary in summaries" :key="summary.key">
      <div v-bind:class="active(summary.key)" @click="onClick(summary.key)">
        <div class="prediction-group-title">
          <p>{{ summary.dataset }}</p>
        </div>
        <div class="prediction-group-body">
          <facet-entry
            enable-highlighting
            :summary="summary"
            :highlight="highlight"
            :enabled-type-changes="[]"
            :row-selection="rowSelection"
            :instanceName="instanceName"
          >
          </facet-entry>
        </div>
      </div>
    </div>

    <!-- TODO: For show right now.-->
    <b-button block variant="primary">
      Export Predictions
    </b-button>

    <b-modal id="export" title="Export">
      <div class="check-message-container">
        <i class="fa fa-check-circle fa-3x check-icon"></i>
        <div>
          This action will export predictions and return to the application
          start screen.
        </div>
      </div>
    </b-modal>
  </div>
</template>

<script lang="ts">
import FacetEntry from "../components/FacetEntry";
import FileUploader from "../components/FileUploader";
import { getSolutionById } from "../util/solutions";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import { getters as predictionsGetters } from "../store/predictions/module";
import {
  actions as appActions,
  getters as appGetters
} from "../store/app/module";
import store from "../store/store";
import {
  EXPORT_SUCCESS_ROUTE,
  ROOT_ROUTE,
  PREDICTION_ROUTE
} from "../store/route/index";
import {
  Variable,
  TaskTypes,
  VariableSummary,
  Highlight,
  RowSelection
} from "../store/dataset/index";
import Vue from "vue";
import { Solution } from "../store/requests/index";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { createRouteEntry, overlayRouteEntry } from "../util/routes";
import { PREDICTION_UPLOAD } from "../util/uploads";
import { getPredictionResultSummary } from "../util/summaries";
import { sum } from "d3";
import { getPredictionsById } from "../util/predictions";

export default Vue.extend({
  name: "prediction-summaries",

  components: {
    FacetEntry,
    FileUploader
  },

  computed: {
    produceRequestId(): string {
      return routeGetters.getRouteProduceRequestId(this.$store);
    },

    fittedSolutionId(): string {
      return routeGetters.getRouteFittedSolutionID(this.$store);
    },

    instanceName(): string {
      return "predictions";
    },

    summaries(): VariableSummary[] {
      return requestGetters
        .getRelevantPredictions(this.$store)
        .map(p => getPredictionResultSummary(p.requestId))
        .filter(p => !!p);
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    }
  },

  methods: {
    onClick(key: string) {
      if (this.summaries && this.produceRequestId !== key) {
        appActions.logUserEvent(this.$store, {
          feature: Feature.SELECT_PREDICTIONS,
          activity: Activity.PREDICTION_ANALYSIS,
          subActivity: SubActivity.MODEL_PREDICTIONS,
          details: { requestId: key }
        });
        const routeEntry = overlayRouteEntry(this.$route, {
          produceRequestId: key,
          highlights: null
        });
        this.$router.push(routeEntry);
      }
    },

    active(summaryKey: string): string {
      return summaryKey === this.produceRequestId
        ? "prediction-group-selected prediction-group"
        : "prediction-group";
    },

    datasetByRequestId(requestId: string): string {
      return getPredictionsById(
        requestGetters.getRelevantPredictions(this.$store),
        requestId
      ).dataset;
    }
  }
});
</script>

<style>
.prediction-group {
  margin: 5px;
  padding: 10px;
  border-bottom-style: solid;
  border-bottom-color: lightgray;
  border-bottom-width: 1px;
}

.prediction-group-title {
  vertical-align: middle;
}

.prediction-group-body {
  padding: 4px 0;
}

.prediction-group-selected {
  padding: 9px;
  border-style: solid;
  border-color: #007bff;
  box-shadow: 0 0 10px #007bff;
  border-width: 1px;
  border-radius: 2px;
  padding-bottom: 10px;
}

.prediction-group:not(.prediction-group-selected):hover {
  padding: 9px;
  border-style: solid;
  border-color: lightgray;
  border-width: 1px;
  border-radius: 2px;
  padding-bottom: 10px;
}
</style>
