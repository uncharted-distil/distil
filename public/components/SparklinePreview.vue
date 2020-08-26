<template>
  <div :class="displayClass" v-observe-visibility="visibilityChanged">
    <sparkline-svg
      :timeseries="timeseries"
      :timeseries-extrema="timeseriesExtrema"
      :forecast="forecast"
      :forecast-extrema="forecastExtrema"
      :highlight-range="highlightRange"
      :join-forecast="!!predictionsId"
      :isDateTime="isDateTime()"
    />
    <i class="fa fa-plus zoom-sparkline-icon" @click.stop="onClick"></i>
    <b-modal
      id="sparkline-zoom-modal"
      hide-footer
      :title="timeseriesId"
      :visible="zoomSparkline"
      @hide="hideModal"
    >
      <div v-if="forecast" class="sparkline-legend">
        <div class="sparkline-legend-historical"></div>
        <div class="sparkline-legend-label">Historical</div>
        <div class="sparkline-legend-missing"></div>
        <div class="sparkline-legend-label">Missing</div>
        <div class="sparkline-legend-predicted"></div>
        <div class="sparkline-legend-label">Predicted</div>
        <div class="sparkline-legend-variability"></div>
        <div class="sparkline-legend-label">Variability</div>
        <div class="sparkline-legend-scoring"></div>
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
import SparklineChart from "../components/SparklineChart";
import SparklineSvg from "../components/SparklineSvg";
import { Dictionary } from "../util/dict";
import { TimeseriesExtrema, TimeSeriesValue } from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";
import {
  getters as resultsGetters,
  actions as resultsActions,
} from "../store/results/module";
import {
  getters as predictionsGetters,
  actions as predictionsActions,
} from "../store/predictions/module";
import * as types from "../util/types";

export default Vue.extend({
  name: "sparkline-preview",

  components: {
    SparklineSvg,
    SparklineChart,
  },

  props: {
    facetView: Boolean as () => Boolean,
    truthDataset: String as () => string,
    forecastDataset: String as () => string,
    xCol: String as () => string,
    yCol: String as () => string,
    timeseriesCol: String as () => string,
    timeseriesId: String as () => string,
    solutionId: String as () => string,
    predictionsId: String as () => string,
    includeForecast: Boolean as () => boolean,
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
    timeseries(): TimeSeriesValue[] {
      if (this.solutionId) {
        const timeseries = resultsGetters.getPredictedTimeseries(this.$store);
        const solutions = timeseries[this.solutionId];
        if (!solutions) {
          return null;
        }
        return solutions.timeseriesData[this.timeseriesId];
      }

      if (this.predictionsId) {
        const timeseries = predictionsGetters.getPredictedTimeseries(
          this.$store,
        );
        const predictions = timeseries[this.predictionsId];
        if (!predictions) {
          return null;
        }
        return predictions.timeseriesData[this.timeseriesId];
      }

      const timeseries = datasetGetters.getTimeseries(this.$store);
      const datasets = timeseries[this.truthDataset];
      if (!datasets) {
        return null;
      }
      return datasets.timeseriesData[this.timeseriesId];
    },

    forecast(): TimeSeriesValue[] {
      if (this.solutionId && this.includeForecast) {
        const forecasts = resultsGetters.getPredictedForecasts(this.$store);
        const solutions = forecasts[this.solutionId];
        if (!solutions || !solutions.forecastData[this.timeseriesId]) {
          return null;
        }
        return solutions.forecastData[this.timeseriesId];
      }

      if (this.predictionsId && this.includeForecast) {
        const forecasts = predictionsGetters.getPredictedForecasts(this.$store);
        const predictions = forecasts[this.predictionsId];
        if (!predictions || !predictions.forecastData[this.timeseriesId]) {
          return null;
        }
        return predictions.forecastData[this.timeseriesId];
      }

      return null;
    },

    highlightRange(): number[] {
      if (this.solutionId && this.includeForecast) {
        const forecasts = resultsGetters.getPredictedForecasts(this.$store);
        const solutions = forecasts[this.solutionId];
        if (!solutions || !solutions.forecastRange[this.timeseriesId]) {
          return null;
        }
        return solutions.forecastRange[this.timeseriesId];
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
        return solutions.isDateTime[this.timeseriesId];
      } else if (this.predictionsId) {
        const timeseries = predictionsGetters.getPredictedTimeseries(
          this.$store,
        );
        const datasets = timeseries[this.truthDataset];
        if (!datasets) {
          return null;
        }
        return datasets.isDateTime[this.timeseriesId];
      } else {
        const timeseries = datasetGetters.getTimeseries(this.$store);
        const datasets = timeseries[this.truthDataset];
        if (!datasets) {
          return null;
        }
        return datasets.isDateTime[this.timeseriesId];
      }
    },

    visibilityChanged(isVisible: boolean) {
      this.isVisible = isVisible;
      if (this.isVisible && !this.hasRequested) {
        this.requestTimeseries();
        return;
      }
    },
    onClick() {
      this.zoomSparkline = true;
    },
    hideModal() {
      this.zoomSparkline = false;
    },
    requestTimeseries() {
      this.hasRequested = true;

      if (this.solutionId) {
        resultsActions.fetchForecastedTimeseries(this.$store, {
          dataset: this.truthDataset,
          xColName: this.xCol,
          yColName: this.yCol,
          timeseriesColName: this.timeseriesCol,
          timeseriesId: this.timeseriesId,
          solutionId: this.solutionId,
        });
      } else if (this.predictionsId) {
        predictionsActions.fetchForecastedTimeseries(this.$store, {
          truthDataset: this.truthDataset,
          forecastDataset: this.forecastDataset,
          xColName: this.xCol,
          yColName: this.yCol,
          timeseriesColName: this.timeseriesCol,
          timeseriesId: this.timeseriesId,
          predictionsId: this.predictionsId,
        });
      } else {
        datasetActions.fetchTimeseries(this.$store, {
          dataset: this.truthDataset,
          xColName: this.xCol,
          yColName: this.yCol,
          timeseriesColName: this.timeseriesCol,
          timeseriesId: this.timeseriesId,
        });
      }
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
