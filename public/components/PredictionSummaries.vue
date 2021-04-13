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
  <div class="prediction-summaries mh-100">
    <p class="nav-link font-weight-bold">Predictions for Dataset</p>
    <div v-for="meta in metaSummaries" :key="meta.summary.key">
      <div :class="active(meta.summary.key)" @click="onClick(meta.summary.key)">
        <header class="prediction-group-title" :title="meta.summary.dataset">
          {{ meta.summary.dataset }}
        </header>
        <div class="prediction-group-body">
          <!-- we need the new facets in here-->
          <prediction-group
            :confidenceSummary="meta.confidence"
            :predictedSummary="meta.summary"
            :rankingSummary="meta.rank"
            :highlights="highlights"
            @categorical-click="onCategoricalClick"
            @numerical-click="onNumericalClick"
            @range-change="onRangeChange"
          />
        </div>
      </div>
    </div>

    <b-button v-b-modal.save> Create Dataset </b-button>

    <b-modal id="save" title="Create Dataset" @ok="createDataset">
      <div class="check-message-container d-flex justify-content-around">
        <i class="fa fa-file-text-o fa-3x" aria-hidden="true"></i>
        <div>
          <b-form-input
            v-model="newDatasetName"
            placeholder="Enter dataset name to use for new dataset"
          />
          <b-form-checkbox v-model="includeAllFeatures" class="pt-2">
            Include data not used in model
          </b-form-checkbox>
        </div>
      </div>
    </b-modal>

    <b-button variant="primary" class="float-right mt-2" v-b-modal.export>
      Export Predictions
    </b-button>

    <b-modal id="export" title="Export" @ok="savePredictions">
      <div class="check-message-container d-flex justify-content-around">
        <i class="fa fa-file-text-o fa-3x" aria-hidden="true"></i>
        <div>
          <b-form-input
            v-model="saveFileName"
            placeholder="Enter name to save as"
          ></b-form-input>
          <b-form-select
            v-model="selectedFormat"
            name="model-scoring"
            size="sm"
          >
            <b-form-select-option
              v-for="format in formats"
              :key="format"
              :value="format"
            >
              {{ format }}
            </b-form-select-option>
          </b-form-select>
        </div>
      </div>
    </b-modal>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import FileUploader from "../components/FileUploader.vue";
import FacetNumerical from "../components/facets/FacetNumerical.vue";
import FacetCategorical from "../components/facets/FacetCategorical.vue";
import PredictionGroup from "./PredictionGroup.vue";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import { actions as predictionActions } from "../store/predictions/module";
import { actions as appActions } from "../store/app/module";

import {
  VariableSummary,
  Highlight,
  RowSelection,
} from "../store/dataset/index";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { getFacetByType } from "../util/facets";
import { overlayRouteEntry } from "../util/routes";
import {
  getPredictionResultSummary,
  getIDFromKey,
  getPredictionConfidenceSummary,
  getPredictionRankSummary,
} from "../util/summaries";
import { getPredictionsById } from "../util/predictions";
import { updateHighlight, clearHighlight } from "../util/highlights";
import { Dictionary } from "vue-router/types/router";

export default Vue.extend({
  name: "prediction-summaries",

  components: {
    FacetNumerical,
    FacetCategorical,
    FileUploader,
    PredictionGroup,
  },

  data() {
    return {
      saveFileName: "",
      newDatasetName: "",
      includeAllFeatures: false,
      selectedFormat: "",
      formats: ["csv", "geojson"],
    };
  },

  computed: {
    produceRequestId(): string {
      return routeGetters.getRouteProduceRequestId(this.$store);
    },

    instanceName(): string {
      return requestGetters.getActivePredictions(this.$store).feature;
    },
    metaSummaries(): Array<{
      rank: VariableSummary;
      confidence: VariableSummary;
      summary: VariableSummary;
    }> {
      const result = [];
      requestGetters.getRelevantPredictions(this.$store).forEach((p) => {
        const meta = {
          rank: getPredictionRankSummary(p.resultId),
          confidence: getPredictionConfidenceSummary(p.resultId),
          summary: getPredictionResultSummary(p.requestId),
        };
        if (!meta.rank && !meta.confidence && !meta.summary) {
          return;
        }
        result.push(meta);
      });
      return result;
    },
    summaries(): VariableSummary[] {
      // get the list of variable summaries, sorting by timestamp
      return requestGetters
        .getRelevantPredictions(this.$store)
        .map((p) => getPredictionResultSummary(p.requestId))
        .filter((p) => !!p);
    },

    highlights(): Highlight[] {
      return routeGetters.getDecodedHighlights(this.$store);
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },
  },

  methods: {
    getFacetByType: getFacetByType,

    onClick(key: string) {
      // Note that the key is of the form <requestId>:predicted and so needs to be parsed.
      const requestId = getIDFromKey(key);
      if (this.summaries && this.produceRequestId !== requestId) {
        appActions.logUserEvent(this.$store, {
          feature: Feature.SELECT_PREDICTIONS,
          activity: Activity.PREDICTION_ANALYSIS,
          subActivity: SubActivity.MODEL_PREDICTIONS,
          details: { requestId: key },
        });
        const dataset = getPredictionsById(
          requestGetters.getPredictions(this.$store),
          requestId
        ).dataset;
        const entry = overlayRouteEntry(this.$route, {
          produceRequestId: requestId,
          highlights: null,
          predictionsDataset: dataset,
        });
        this.$router.push(entry).catch((err) => console.warn(err));
      }
    },

    onCategoricalClick(args: {
      context: string;
      key: string;
      value: string;
      dataset: string;
    }) {
      if (args.key && args.value) {
        // If this isn't the currently selected prediction set, first update it.
        // Note that the key is of the form <requestId>:predicted and so needs to be
        // parsed.
        if (
          this.summaries &&
          this.produceRequestId !== getIDFromKey(args.key)
        ) {
          this.onClick(args.key);
        }

        // extract the var name from the key
        updateHighlight(this.$router, {
          context: args.context,
          dataset: args.dataset,
          key: args.key,
          value: args.value,
        });
      } else {
        clearHighlight(this.$router, args.key);
      }
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.PREDICTION_ANALYSIS,
        subActivity: SubActivity.MODEL_PREDICTIONS,
        details: { key: args.key, value: args.value },
      });
    },

    onNumericalClick(args: {
      context: string;
      key: string;
      value: { from: number; to: number };
      dataset: string;
    }) {
      const uniqueHighlight = this.highlights.reduce(
        (acc, highlight) => highlight.key !== args.key || acc,
        false
      );
      if (uniqueHighlight) {
        // If this isn't the currently selected prediction set, first update it.
        // Note that the key is of the form <requestId>:predicted and so needs to be
        // parsed.
        if (
          this.summaries &&
          this.produceRequestId !== getIDFromKey(args.key)
        ) {
          this.onClick(args.key);
        }
        if (args.key && args.value) {
          updateHighlight(this.$router, {
            context: args.context,
            dataset: args.dataset,
            key: args.key,
            value: args.value,
          });
        } else {
          clearHighlight(this.$router, args.key);
        }
        appActions.logUserEvent(this.$store, {
          feature: Feature.CHANGE_HIGHLIGHT,
          activity: Activity.PREDICTION_ANALYSIS,
          subActivity: SubActivity.MODEL_EXPLANATION,
          details: { key: args.key, value: args.value },
        });
      }
    },

    onRangeChange(args: {
      context: string;
      key: string;
      value: { from: { label: string[] }; to: { label: string[] } };
      dataset: string;
    }) {
      if (args.key && args.value) {
        updateHighlight(this.$router, {
          context: args.context,
          dataset: args.dataset,
          key: args.key,
          value: args.value,
        });
      } else {
        clearHighlight(this.$router, args.key);
      }
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.PREDICTION_ANALYSIS,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: args.key, value: args.value },
      });
      this.$emit("range-change", args.key, args.value);
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
    },

    async savePredictions() {
      let dataStr = await predictionActions.fetchExportData(this.$store, {
        produceRequestId: this.produceRequestId,
        format: this.selectedFormat,
      });
      if (!dataStr) {
        console.error("No Data");
        return;
      }

      let dataType = "text/csv";
      let extension = "csv";
      if (this.selectedFormat == "geojson") {
        dataType = "application/json";
        extension = "json";
        dataStr = JSON.stringify(dataStr);
      }

      const hiddenElement = document.createElement("a");
      const fileName =
        this.saveFileName === "" ? "predictions" : this.saveFileName;
      hiddenElement.href =
        `data:${dataType};charset=utf-8,` + encodeURI(dataStr);
      hiddenElement.target = "_blank";
      hiddenElement.download = `${fileName}.${extension}`;
      hiddenElement.click();
    },

    async createDataset() {
      const err = await predictionActions.createDataset(this.$store, {
        produceRequestId: this.produceRequestId,
        newDatasetName: this.newDatasetName,
        includeDatasetFeatures: this.includeAllFeatures,
      });
      const location = "b-toaster-bottom-right";
      if (err) {
        this.$bvToast.toast(err.message, {
          title: "Error creating dataset ${this.newDatasetName}",
          solid: true,
          appendToast: true,
          variant: "danger",
          toaster: location,
        });
        return;
      }
      this.$bvToast.toast(`Success`, {
        title: `Success creating dataset ${this.newDatasetName}`,
        solid: true,
        appendToast: true,
        variant: "success",
        toaster: location,
      });
    },
  },
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
  color: var(--color-text-base);
  overflow: hidden;
  padding: 0.25rem 0 0.25rem;
  text-overflow: ellipsis;
}

.prediction-group-body {
  padding: 4px 0;
}

.prediction-group-selected {
  padding: 9px;
  border-style: solid;
  border-color: var(--blue);
  box-shadow: 0 0 10px var(--blue);
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
