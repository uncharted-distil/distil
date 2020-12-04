<template>
  <div class="sparkline-row">
    <div class="timeseries-var-col">{{ timeseriesId }}</div>
    <div class="timeseries-min-col">{{ min.toFixed(2) }}</div>
    <div class="timeseries-max-col">{{ max.toFixed(2) }}</div>
    <sparkline-svg
      class="sparkline-row-chart"
      v-bind:class="{ 'has-prediction': !!prediction }"
      :highlight-pixel-x="highlightPixelX"
      :timeseries-extrema="timeseriesExtrema"
      :timeseries="timeseries"
      :forecast="forecast"
      :highlightRange="highlightRange"
      :isDateTime="isDateTime"
    >
    </sparkline-svg>
    <div
      v-if="prediction"
      class="timeseries-prediction-col"
      v-bind:class="{
        'correct-prediction': prediction.isCorrect,
        'incorrect-prediction': !prediction.isCorrect,
      }"
    >
      {{ prediction.value }}
    </div>
  </div>
</template>

<script lang="ts">
import * as d3 from "d3";
import $ from "jquery";
import Vue from "vue";
import SparklineSvg from "./SparklineSvg.vue";
import { getters as routeGetters } from "../store/route/module";
import { TimeseriesExtrema, TimeSeriesValue } from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";
import {
  getters as resultsGetters,
  actions as resultsActions,
} from "../store/results/module";

export default Vue.extend({
  name: "sparkline-row",

  components: {
    SparklineSvg,
  },

  props: {
    highlightPixelX: {
      type: Number as () => number,
    },
    xCol: String as () => string,
    yCol: String as () => string,
    timeseriesCol: String as () => string,
    timeseriesId: String as () => string,
    timeseriesExtrema: {
      type: Object as () => TimeseriesExtrema,
    },
    solutionId: String as () => string,
    prediction: Object as () => any,
    includeForecast: Boolean as () => boolean,
  },
  data() {
    return {
      isVisible: false,
      hasRequested: false,
    };
  },
  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    timeseries(): TimeSeriesValue[] {
      if (this.solutionId) {
        return resultsGetters.getPredictedTimeseries(this.$store)[
          this.solutionId
        ][this.timeseriesId];
      } else {
        return datasetGetters.getTimeseries(this.$store)[this.dataset][
          this.timeseriesId
        ];
      }
    },
    forecast(): TimeSeriesValue[] {
      if (this.solutionId && this.includeForecast) {
        const forecasts = resultsGetters.getPredictedForecasts(this.$store);
        const solutions = forecasts[this.solutionId];
        if (!solutions || !solutions[this.timeseriesId]) {
          return null;
        }
        return solutions[this.timeseriesId].forecast;
      } else {
        return null;
      }
    },
    highlightRange(): number[] {
      if (this.solutionId && this.includeForecast) {
        const forecasts = resultsGetters.getPredictedForecasts(this.$store);
        const solutions = forecasts[this.solutionId];
        if (!solutions || !solutions[this.timeseriesId]) {
          return null;
        }
        return solutions[this.timeseriesId].forecastTestRange;
      } else {
        return null;
      }
    },
    min(): number {
      return this.timeseries ? d3.min(this.timeseries, (d) => d.value) : 0;
    },
    max(): number {
      return this.timeseries ? d3.max(this.timeseries, (d) => d.value) : 0;
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
    },
  },
});
</script>

<style>
.sparkline-row {
  position: relative;
  width: 100%;
  height: 32px;
  line-height: 32px;
  vertical-align: middle;
  border-bottom: 1px solid #999;
  padding: 0 8px;
}
.sparkline-row-chart {
  float: left;
  position: relative;
  line-height: 32px;
  height: 32px;
  width: calc(100% - 276px);
}
.sparkline-row-chart.has-prediction {
  width: calc(100% - 372px);
}
.correct-prediction {
  color: #00c6e1;
}
.incorrect-prediction {
  color: #e05353;
}
</style>
