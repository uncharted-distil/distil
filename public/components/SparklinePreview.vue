<template>
  <div
    class="sparkline-preview-container"
    v-observe-visibility="visibilityChanged"
  >
    <sparkline-svg
      :timeseries-extrema="timeseriesExtrema"
      :timeseries="timeseries"
      :forecast="forecast"
      :forecast-extrema="forecastExtrema"
      :highlightRange="highlightRange"
    >
    </sparkline-svg>
    <i class="fa fa-plus zoom-sparkline-icon" @click.stop="onClick"></i>
    <b-modal
      id="sparkline-zoom-modal"
      :title="timeseriesId"
      @hide="hideModal"
      :visible="zoomSparkline"
      hide-footer
    >
      <sparkline-chart
        :timeseries="timeseries"
        :forecast="forecast"
        :highlightRange="highlightRange"
        :xAxisTitle="xCol"
        :yAxisTitle="yCol"
        :xAxisDateTime="isDateTime"
        v-if="zoomSparkline"
      ></sparkline-chart>
    </b-modal>
  </div>
</template>

<script lang="ts">
import * as d3 from "d3";
import Vue from "vue";
import SparklineChart from "../components/SparklineChart";
import SparklineSvg from "../components/SparklineSvg";
import { Dictionary } from "../util/dict";
import { TimeseriesExtrema } from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions
} from "../store/dataset/module";
import {
  getters as resultsGetters,
  actions as resultsActions
} from "../store/results/module";
import * as types from "../util/types";

export default Vue.extend({
  name: "sparkline-preview",

  components: {
    SparklineSvg,
    SparklineChart
  },

  props: {
    dataset: String as () => string,
    xCol: String as () => string,
    yCol: String as () => string,
    timeseriesCol: String as () => string,
    timeseriesId: String as () => string,
    solutionId: String as () => string,
    includeForecast: Boolean as () => boolean
  },
  data() {
    return {
      zoomSparkline: false,
      isVisible: false,
      hasRequested: false
    };
  },
  computed: {
    timeseries(): number[][] {
      if (this.solutionId) {
        const timeseries = resultsGetters.getPredictedTimeseries(this.$store);
        const solutions = timeseries[this.solutionId];
        if (!solutions) {
          return null;
        }
        return solutions.timeseriesData[this.timeseriesId];
      } else {
        const timeseries = datasetGetters.getTimeseries(this.$store);
        const datasets = timeseries[this.dataset];
        if (!datasets) {
          return null;
        }
        return datasets.timeseriesData[this.timeseriesId];
      }
    },
    forecast(): number[][] {
      if (this.solutionId && this.includeForecast) {
        const forecasts = resultsGetters.getPredictedForecasts(this.$store);
        const solutions = forecasts[this.solutionId];
        if (!solutions || !solutions.forecastData[this.timeseriesId]) {
          return null;
        }
        return solutions.forecastData[this.timeseriesId];
      } else {
        return null;
      }
    },
    highlightRange(): number[] {
      if (this.solutionId && this.includeForecast) {
        const forecasts = resultsGetters.getPredictedForecasts(this.$store);
        const solutions = forecasts[this.solutionId];
        if (!solutions || !solutions.forecastRange[this.timeseriesId]) {
          return null;
        }
        return solutions.forecastRange[this.timeseriesId];
      } else {
        return null;
      }
    },
    timeseriesExtrema(): TimeseriesExtrema {
      if (!this.timeseries) {
        return null;
      }
      return {
        x: {
          min: d3.min(this.timeseries, d => d[0]),
          max: d3.max(this.timeseries, d => d[0])
        },
        y: {
          min: d3.min(this.timeseries, d => d[1]),
          max: d3.max(this.timeseries, d => d[1])
        }
      };
    },
    forecastExtrema(): TimeseriesExtrema {
      if (!this.forecast) {
        return null;
      }
      return {
        x: {
          min: d3.min(this.forecast, d => d[0]),
          max: d3.max(this.forecast, d => d[0])
        },
        y: {
          min: d3.min(this.forecast, d => d[1]),
          max: d3.max(this.forecast, d => d[1])
        }
      };
    },
    isDateTime(): boolean {
      if (this.solutionId) {
        const timeseries = resultsGetters.getPredictedTimeseries(this.$store);
        const solutions = timeseries[this.solutionId];
        if (!solutions) {
          return null;
        }
        return solutions.isDateTime[this.timeseriesId];
      } else {
        const timeseries = datasetGetters.getTimeseries(this.$store);
        const datasets = timeseries[this.dataset];
        if (!datasets) {
          return null;
        }
        return datasets.isDateTime[this.timeseriesId];
      }
    }
  },
  methods: {
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
          dataset: this.dataset,
          xColName: this.xCol,
          yColName: this.yCol,
          timeseriesColName: this.timeseriesCol,
          timeseriesID: this.timeseriesId,
          solutionId: this.solutionId
        });
      } else {
        datasetActions.fetchTimeseries(this.$store, {
          dataset: this.dataset,
          xColName: this.xCol,
          yColName: this.yCol,
          timeseriesColName: this.timeseriesCol,
          timeseriesID: this.timeseriesId
        });
      }
    }
  }
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
</style>
