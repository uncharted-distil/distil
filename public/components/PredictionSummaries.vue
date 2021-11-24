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
    <p v-if="includeTitle" class="nav-link font-weight-bold">
      Predictions for Dataset
    </p>
    <div class="prediction-group-container overflow-auto">
      <div v-for="meta in metaSummaries" :key="meta.summary.key">
        <div
          :class="active(meta.summary.key)"
          @click="onClick(meta.summary.key)"
        >
          <header class="prediction-group-title" :title="meta.summary.dataset">
            {{ meta.summary.dataset }}
            <div
              class="pull-right pl-2 solution-button"
              @click.stop="onCollapseClick(meta.prediction.requestId)"
            >
              <i
                class="fa"
                :class="{
                  'fa-angle-down': !openSolution.has(meta.prediction.requestId),
                  'fa-angle-up': openSolution.has(meta.prediction.requestId),
                }"
              />
            </div>
          </header>
          <div class="prediction-group-datetime">
            {{ predictionTimestamp(meta.summary.dataset) }}
          </div>
          <div class="prediction-group-body">
            <!-- we need the new facets in here-->
            <prediction-group
              :confidence-summary="meta.confidence"
              :predicted-summary="meta.summary"
              :ranking-summary="meta.rank"
              :highlights="highlights"
              :prediction="meta.prediction"
              @categorical-click="onCategoricalClick"
              @numerical-click="onNumericalClick"
              @range-change="onRangeChange"
            />
          </div>
        </div>
      </div>
    </div>
    <b-button v-if="includeFooter" v-b-modal.save class="mt-3">
      Create Dataset
    </b-button>

    <b-modal id="save" title="Create Dataset" @ok="createDataset">
      <form ref="createDatasetForm">
        <b-form-group
          label="Dataset Name"
          label-for="model-name-input"
          invalid-feedback="Dataset Name is Required"
          :state="datasetModelNameState"
        >
          <b-form-input
            id="model-name-input"
            v-model="newDatasetName"
            placeholder="Enter dataset name to use for new dataset"
            :state="datasetModelNameState"
          />
        </b-form-group>
        <b-form-group>
          <b-form-checkbox v-model="includeAllFeatures" class="pt-2">
            Include data not used in model
          </b-form-checkbox>
        </b-form-group>
      </form>
    </b-modal>

    <b-button
      v-if="includeFooter"
      v-b-modal.export
      variant="primary"
      class="float-right mt-3"
    >
      Export Predictions
    </b-button>

    <b-modal id="export" title="Export" @ok="savePredictions">
      <form ref="exportPredictionsForm">
        <b-form-group
          label="Export File Name"
          label-for="export-name-input"
          invalid-feedback="File Name is Required"
        >
          <b-form-input
            id="export-name-input"
            v-model="saveFileName"
            placeholder="Enter name to save as"
            :state="datasetExportNameState"
          />
        </b-form-group>
        <b-form-group label="Export File Type" label-for="export-type-input">
          <b-form-select
            id="export-type-input"
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
        </b-form-group>
      </form>
    </b-modal>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
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
import {
  updateHighlight,
  clearHighlight,
  UPDATE_FOR_KEY,
} from "../util/highlights";
import { reviseOpenSolutions } from "../util/solutions";
import moment from "moment";
import { EventList } from "../util/events";

export default Vue.extend({
  name: "PredictionSummaries",

  components: {
    PredictionGroup,
  },
  props: {
    includeFooter: { type: Boolean as () => boolean, default: false },
    includeTitle: { type: Boolean as () => boolean, default: false },
    isBusy: { type: Boolean as () => boolean, default: false },
  },
  data() {
    return {
      saveFileName: "",
      newDatasetName: "",
      includeAllFeatures: true,
      selectedFormat: "csv",
      formats: ["csv", "geojson"],
      datasetModelNameState: false,
      datasetExportNameState: null,
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
          prediction: p,
        };
        if (
          !meta.rank &&
          !meta.confidence &&
          !meta.summary &&
          !meta.prediction &&
          !meta.summary?.key
        ) {
          return;
        }
        result.push(meta);
      });
      return result.sort((a, b) => {
        return (
          moment(b.prediction.timestamp).unix() -
          moment(a.prediction.timestamp).unix()
        );
      });
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

    openSolution(): Map<string, boolean> {
      return new Map(
        routeGetters.getRouteOpenSolutions(this.$store).map((s) => {
          return [s, true];
        })
      );
    },
  },

  watch: {
    newDatasetName() {
      if (this.newDatasetName !== null && this.newDatasetName.length > 0) {
        this.datasetModelNameState = true;
      } else {
        this.datasetModelNameState = false;
      }
    },
  },

  methods: {
    getFacetByType: getFacetByType,
    onCollapseClick(requestId: string) {
      reviseOpenSolutions(requestId, this.$route, this.$router);
      if (this.openSolution.has(requestId)) {
        this.$emit(EventList.SUMMARIES.FETCH_SUMMARY_PREDICTION, requestId);
      }
    },

    onClick(key: string) {
      // Note that the key is of the form <requestId>:predicted and so needs to be parsed.
      const requestId = getIDFromKey(key);
      if (
        this.summaries &&
        this.produceRequestId !== requestId &&
        !this.isBusy
      ) {
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
          colorScaleVariable: "",
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
      let highlight = this.highlights.find((h) => {
        return h.key === args.key;
      });
      if (
        args.key &&
        args.value &&
        Array.isArray(args.value) &&
        args.value.length > 0
      ) {
        if (
          this.summaries &&
          this.produceRequestId !== getIDFromKey(args.key)
        ) {
          this.onClick(args.key);
        }
        highlight = highlight ?? {
          context: args.context,
          dataset: args.dataset,
          key: args.key,
          value: [],
        };
        highlight.value = args.value;
        updateHighlight(this.$router, highlight, UPDATE_FOR_KEY);
      } else {
        clearHighlight(this.$router, highlight.key);
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
        updateHighlight(
          this.$router,
          {
            context: args.context,
            dataset: args.dataset,
            key: args.key,
            value: args.value,
          },
          UPDATE_FOR_KEY
        );
      } else {
        clearHighlight(this.$router, args.key);
      }
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.PREDICTION_ANALYSIS,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: args.key, value: args.value },
      });
      this.$emit(EventList.FACETS.RANGE_CHANGE_EVENT, args.key, args.value);
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

    async savePredictions(bvModalEvt) {
      if (!this.saveFileName) {
        bvModalEvt.preventDefault();
        this.datasetExportNameState = false;
        return;
      }
      this.datasetExportNameState = true;

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

    async createDataset(bvModalEvt) {
      if (!this.newDatasetName) {
        bvModalEvt.preventDefault();
        this.datasetModelNameState = false;
        return;
      }
      this.datasetModelNameState = true;

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

    predictionTimestamp(datasetName: string): string {
      const timestamp = requestGetters
        .getRelevantPredictions(this.$store)
        .find((p) => p.dataset === datasetName).timestamp;
      return new Date(Date.parse(timestamp)).toLocaleString(
        navigator.language,
        {
          day: "numeric",
          month: "long",
          weekday: "short",
          hour: "numeric",
          minute: "numeric",
          timeZoneName: "short",
        }
      );
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
.prediction-group-container {
  max-height: 87%;
}

.prediction-group-datetime {
  font-size: 75%;
  color: var(--color-text-second);
}
</style>
