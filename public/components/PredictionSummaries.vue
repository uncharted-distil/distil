<template>
  <div class="prediction-summaries">
    <p class="nav-link font-weight-bold">
      Predictions for Dataset
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
            :key="summary.key"
            :highlight="highlight"
            :enabled-type-changes="[]"
            :row-selection="rowSelection"
            :instanceName="instanceName"
            @facet-click="onCategoricalClick"
            @numerical-click="onNumericalClick"
            @range-change="onRangeChange"
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
import { getPredictionResultSummary, getIDFromKey } from "../util/summaries";
import { sum } from "d3";
import { getPredictionsById } from "../util/predictions";
import { updateHighlight, clearHighlight } from "../util/highlights";
import moment from "moment";
import _ from "lodash";

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

    instanceName(): string {
      return requestGetters.getActivePredictions(this.$store).feature;
    },

    summaries(): VariableSummary[] {
      // get the list of variable summaries, sorting by timestamp
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
      // Note that the key is of the form <requestId>:predicted and so needs to be
      // parsed.
      const requestId = getIDFromKey(key);
      if (this.summaries && this.produceRequestId !== requestId) {
        appActions.logUserEvent(this.$store, {
          feature: Feature.SELECT_PREDICTIONS,
          activity: Activity.PREDICTION_ANALYSIS,
          subActivity: SubActivity.MODEL_PREDICTIONS,
          details: { requestId: key }
        });
        const dataset = getPredictionsById(
          requestGetters.getPredictions(this.$store),
          requestId
        ).dataset;
        const routeEntry = overlayRouteEntry(this.$route, {
          produceRequestId: requestId,
          highlights: null,
          predictionsDataset: dataset
        });
        this.$router.push(routeEntry);
      }
    },

    onCategoricalClick(
      context: string,
      key: string,
      value: string,
      dataset: string
    ) {
      if (key && value) {
        // If this isn't the currently selected prediction set, first update it.
        // Note that the key is of the form <requestId>:predicted and so needs to be
        // parsed.
        if (this.summaries && this.produceRequestId !== getIDFromKey(key)) {
          this.onClick(key);
        }

        // extract the var name from the key
        updateHighlight(this.$router, {
          context: context,
          dataset: dataset,
          key: key,
          value: value
        });
      } else {
        clearHighlight(this.$router);
      }
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.PREDICTION_ANALYSIS,
        subActivity: SubActivity.MODEL_PREDICTIONS,
        details: { key: key, value: value }
      });
    },

    onNumericalClick(
      context: string,
      key: string,
      value: { from: number; to: number },
      dataset: string
    ) {
      if (!this.highlight || this.highlight.key !== key) {
        // If this isn't the currently selected prediction set, first update it.
        // Note that the key is of the form <requestId>:predicted and so needs to be
        // parsed.
        if (this.summaries && this.produceRequestId !== getIDFromKey(key)) {
          this.onClick(key);
        }
        updateHighlight(this.$router, {
          context: context,
          dataset: dataset,
          key: key,
          value: value
        });
        appActions.logUserEvent(this.$store, {
          feature: Feature.CHANGE_HIGHLIGHT,
          activity: Activity.PREDICTION_ANALYSIS,
          subActivity: SubActivity.MODEL_EXPLANATION,
          details: { key: key, value: value }
        });
      }
    },

    onRangeChange(
      context: string,
      key: string,
      value: { from: { label: string[] }; to: { label: string[] } },
      dataset: string
    ) {
      updateHighlight(this.$router, {
        context: context,
        dataset: dataset,
        key: key,
        value: value
      });
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.PREDICTION_ANALYSIS,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: key, value: value }
      });
      this.$emit("range-change", key, value);
    },

    active(summaryKey: string): string {
      const predictions = getPredictionsById(
        requestGetters.getRelevantPredictions(this.$store),
        this.produceRequestId
      );
      return summaryKey === predictions.predictedKey
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
.prediction-summaries {
  overflow-x: hidden;
  overflow-y: auto;
}

.prediction-summaries .facets-facet-base {
  overflow: visible;
}

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