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
  <div :class="displayClass">
    <sparkline-svg
      :timeseries="timeseries"
      :timeseries-extrema="timeseriesExtrema"
      :forecast="forecast"
      :forecast-extrema="forecastExtrema"
      :highlight-range="highlightRange"
      :join-forecast="!!predictionsId"
      :is-date-time="isDateTime()"
    />
    <i class="fa fa-plus zoom-sparkline-icon" @click.stop="onClick" />
    <b-modal
      id="sparkline-zoom-modal"
      hide-footer
      :title="timeseriesId"
      :visible="zoomSparkline"
      @hide="hideModal"
    >
      <div v-if="forecast" class="sparkline-legend">
        <div class="sparkline-legend-historical" />
        <div class="sparkline-legend-label">Historical</div>
        <div class="sparkline-legend-missing" />
        <div class="sparkline-legend-label">Missing</div>
        <div class="sparkline-legend-predicted" />
        <div class="sparkline-legend-label">Predicted</div>
        <div class="sparkline-legend-variability" />
        <div class="sparkline-legend-label">Variability</div>
        <div class="sparkline-legend-scoring" />
        <div class="sparkline-legend-label">Scoring</div>
      </div>
      <sparkline-chart
        v-if="zoomSparkline"
        :timeseries="timeseries"
        :forecast="forecast"
        :highlight-range="highlightRange"
        :x-axis-title="xCol"
        :y-axis-title="yCol"
        :x-axis-date-time="isDateTime()"
        :join-forecast="!!predictionsId"
      />
    </b-modal>
  </div>
</template>

<script lang="ts">
import * as d3 from "d3";
import Vue from "vue";
import SparklineChart from "../components/SparklineChart.vue";
import SparklineSvg from "../components/SparklineSvg.vue";
import {
  TimeSeries,
  TimeseriesExtrema,
  TimeSeriesValue,
} from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as resultsGetters } from "../store/results/module";
import { getters as predictionsGetters } from "../store/predictions/module";
import { Dictionary } from "../util/dict";
export default Vue.extend({
  name: "SparklinePreview",

  components: {
    SparklineSvg,
    SparklineChart,
  },

  props: {
    facetView: Boolean as () => boolean,
    truthDataset: String as () => string,
    forecastDataset: String as () => string,
    xCol: String as () => string,
    yCol: String as () => string,
    variableKey: { type: String as () => string, default: "" },
    timeseriesId: { type: String as () => string, default: "" },
    solutionId: String as () => string,
    predictionsId: String as () => string,
    includeForecast: Boolean as () => boolean,
    uniqueTrail: { type: String as () => string, default: "" },
    getTimeseries: {
      type: Function,
      default: null,
    },
  },
  data() {
    return {
      zoomSparkline: false,
      isVisible: false,
      hasRequested: false,
    };
  },
  computed: {
    displayClass(): string {
      return this.facetView
        ? "facet-sparkline-preview-container"
        : "sparkline-preview-container";
    },
    timeseriesUniqueId(): string {
      return this.variableKey + this.timeseriesId + this.uniqueTrail;
    },
    timeseries(): TimeSeriesValue[] {
      if (this.getTimeseries != null) {
        const timeseries = this.getTimeseries() as TimeSeries;
        if (!timeseries) {
          return null;
        }
        return timeseries.timeseriesData[this.timeseriesUniqueId];
      }
      if (this.predictionsId) {
        const timeseries = predictionsGetters.getPredictedTimeseries(
          this.$store
        );
        const predictions = timeseries[this.predictionsId];
        if (!predictions) {
          return null;
        }
        return predictions.timeseriesData[this.timeseriesUniqueId];
      }
      if (this.solutionId) {
        const timeseries = resultsGetters.getPredictedTimeseries(this.$store);
        const solutions = timeseries[this.solutionId];
        if (!solutions) {
          return null;
        }
        return solutions.timeseriesData[this.timeseriesUniqueId];
      }

      const timeseries = datasetGetters.getTimeseries(this.$store);
      const datasets = timeseries[this.truthDataset];
      if (!datasets) {
        return null;
      }
      return datasets.timeseriesData[this.timeseriesUniqueId];
    },

    forecast(): TimeSeriesValue[] {
      if (this.predictionsId && this.includeForecast) {
        const forecasts = predictionsGetters.getPredictedForecasts(this.$store);
        const predictions = forecasts[this.predictionsId];
        if (
          !predictions ||
          !predictions.forecastData[this.timeseriesUniqueId]
        ) {
          return null;
        }
        return predictions.forecastData[this.timeseriesUniqueId];
      }

      if (this.solutionId && this.includeForecast) {
        const forecasts = resultsGetters.getPredictedForecasts(this.$store);
        const solutions = forecasts[this.solutionId];
        if (!solutions || !solutions.forecastData[this.timeseriesUniqueId]) {
          return null;
        }
        return solutions.forecastData[this.timeseriesUniqueId];
      }

      return null;
    },

    highlightRange(): number[] {
      if (this.solutionId && this.includeForecast) {
        const forecasts = resultsGetters.getPredictedForecasts(this.$store);
        const solutions = forecasts[this.solutionId];
        if (!solutions || !solutions.forecastRange[this.timeseriesUniqueId]) {
          return null;
        }
        return solutions.forecastRange[this.timeseriesUniqueId];
      }
      return null;
    },

    timeseriesExtrema(): TimeseriesExtrema {
      if (!this.timeseries) {
        return null;
      }
      return {
        x: {
          min: d3.min(this.timeseries, (d) => d.time),
          max: d3.max(this.timeseries, (d) => d.time),
        },
        y: {
          min: d3.min(this.timeseries, (d) => d.value),
          max: d3.max(this.timeseries, (d) => d.value),
        },
      };
    },
    forecastExtrema(): TimeseriesExtrema {
      if (!this.forecast) {
        return null;
      }
      return {
        x: {
          min: d3.min(this.forecast, (d) => d.time),
          max: d3.max(this.forecast, (d) => d.time),
        },
        y: {
          min: d3.min(this.forecast, (d) => d.value),
          max: d3.max(this.forecast, (d) => d.value),
        },
      };
    },
  },
  methods: {
    isDateTime(): boolean {
      if (this.solutionId) {
        const timeseries = resultsGetters.getPredictedTimeseries(this.$store);
        const solutions = timeseries[this.solutionId];
        if (!solutions) {
          return null;
        }
        return solutions.isDateTime[this.timeseriesUniqueId];
      } else if (this.predictionsId) {
        const timeseries = predictionsGetters.getPredictedTimeseries(
          this.$store
        );
        const datasets = timeseries[this.truthDataset];
        if (!datasets) {
          return null;
        }
        return datasets.isDateTime[this.timeseriesUniqueId];
      } else {
        const timeseries = datasetGetters.getTimeseries(this.$store);
        const datasets = timeseries[this.truthDataset];
        if (!datasets) {
          return null;
        }
        return datasets.isDateTime[this.timeseriesUniqueId];
      }
    },
    onClick() {
      this.zoomSparkline = true;
    },
    hideModal() {
      this.zoomSparkline = false;
    },
  },
});
</script>

<style>
.zoom-sparkline-icon {
  position: absolute;
  right: 4px;
  top: 4px;
  color: #666;
  visibility: hidden;
}

.sparkline-preview-container {
  position: relative;
  min-width: 400px;
  max-width: 500px !important;
  min-height: 45px;
}

.sparkline-preview-container:hover .zoom-sparkline-icon {
  visibility: visible;
}

.facet-sparkline-preview-container {
  position: relative;
  width: 100%;
  max-height: 45px;
}

.facet-sparkline-preview-container:hover .zoom-sparkline-icon {
  visibility: visible;
}

.zoom-sparkline-icon:hover {
  opacity: 0.7;
}

.sparkline-elem-zoom {
  position: relative;
  padding: 32px 16px;
  border-radius: 4px;
}

#sparkline-zoom-modal .modal-dialog {
  max-width: 50%;
}

.sparkline-legend {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.sparkline-legend-label {
  margin-left: 6px;
}

.sparkline-legend-historical {
  background: var(--gray-700);
  width: 16px;
  height: 2px;
}

.sparkline-legend-predicted {
  background: var(--blue);
  width: 16px;
  height: 2px;
  margin-left: 14px;
}

.sparkline-legend-variability {
  background: var(--blue);
  opacity: 0.3;
  width: 12px;
  height: 12px;
  margin-left: 14px;
}

.sparkline-legend-scoring {
  background: var(--yellow);
  opacity: 0.5;
  width: 12px;
  height: 12px;
  margin-left: 14px;
}

.sparkline-legend-missing {
  border-color: var(--gray-500);
  border-style: dotted;
  border-width: 0 0 2px 0;
  width: 16px;
  margin-left: 14px;
}
</style>
