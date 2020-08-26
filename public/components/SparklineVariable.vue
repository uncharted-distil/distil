<template>
  <div class="sparkline-variable">
    <div v-for="summary in timeseriesSet.summaries" :key="summary.label">
      <div class="timeseries-var-col" v-html="summary.label"></div>
      <div class="timeseries-min-col">
        {{ min(summary.timeseries).toFixed(2) }}
      </div>
      <div class="timeseries-max-col">
        {{ max(summary.timeseries).toFixed(2) }}
      </div>
      <sparkline-svg
        class="sparkline-variable-chart"
        :highlight-pixel-x="highlightPixelX"
        :timeseries-extrema="timeseriesExtrema"
        :timeseries="summary.timeseries"
      >
      </sparkline-svg>
    </div>
  </div>
</template>

<script lang="ts">
import * as d3 from "d3";
import _ from "lodash";
import $ from "jquery";
import Vue from "vue";
import SparklineSvg from "./SparklineSvg";
import { Dictionary } from "../util/dict";
import {
  TimeseriesExtrema,
  VariableSummary,
  Histogram,
  Bucket,
} from "../store/dataset/index";

interface TimeseriesSet {
  summaries: TimeseriesSummary[];
  extrema: TimeseriesExtrema;
}

interface TimeseriesSummary {
  label: string;
  key: string;
  category?: string;
  timeseries: number[][];
}
export default Vue.extend({
  name: "sparkline-variable",

  components: {
    SparklineSvg,
  },

  props: {
    summary: Object as () => VariableSummary,
    minX: Number as () => number,
    maxX: Number as () => number,
    highlightPixelX: {
      type: Number as () => number,
    },
  },
  computed: {
    timeseriesSet(): TimeseriesSet {
      if (!this.summary.filtered && !this.summary.baseline) {
        return {
          summaries: [],
          extrema: null,
        };
      }
      const key = this.summary.key;
      const label = this.summary.label;
      const histogram = this.summary.filtered
        ? this.summary.filtered
        : this.summary.baseline;
      return this.variableSummaryToTimeseries(key, label, histogram);
    },
    timeseriesExtrema(): TimeseriesExtrema {
      return {
        x: {
          min: this.minX,
          max: this.maxX,
        },
        y: {
          min: this.timeseriesSet.extrema.y.min,
          max: this.timeseriesSet.extrema.y.max,
        },
      };
    },
  },
  methods: {
    min(timeseries: number[][]): number {
      const min = d3.min(timeseries, (d) => d[1]);
      return min !== undefined ? min : 0;
    },
    max(timeseries: number[][]): number {
      const max = d3.max(timeseries, (d) => d[1]);
      return max !== undefined ? max : 0;
    },

    getExtremaFromTimeseries(points: number[][]): TimeseriesExtrema {
      const extrema = {
        x: { min: Infinity, max: -Infinity },
        y: { min: Infinity, max: -Infinity },
        sum: 0,
      };
      for (let i = 0; i < points.length; i++) {
        extrema.x.min = Math.min(extrema.x.min, Math.min(points[i][0]));
        extrema.x.max = Math.max(extrema.x.max, Math.max(points[i][0]));
        extrema.y.min = Math.min(extrema.y.min, Math.min(points[i][1]));
        extrema.y.max = Math.max(extrema.y.max, Math.max(points[i][1]));
        extrema.sum += points[i][1];
      }
      return extrema;
    },

    mergeExtrema(
      a: TimeseriesExtrema,
      b: TimeseriesExtrema,
    ): TimeseriesExtrema {
      return {
        x: {
          min: Math.min(a.x.min, b.x.min),
          max: Math.max(a.x.max, b.x.max),
        },
        y: {
          min: Math.min(a.y.min, b.y.min),
          max: Math.max(a.y.max, b.y.max),
        },
        sum: a.sum + b.sum,
      };
    },

    numericBucketsToTimeseries(
      key: string,
      label: string,
      buckets: Bucket[],
    ): TimeseriesSet {
      const timeseries = buckets.map((b) => [_.parseInt(b.key), b.count]);
      const extrema = this.getExtremaFromTimeseries(timeseries);

      const summaries = [
        {
          label: label,
          key: key,
          timeseries: timeseries,
        },
      ];

      return {
        summaries: summaries,
        extrema: extrema,
      };
    },

    categoryBucketsToTimeseries(
      key: string,
      label: string,
      buckets: Dictionary<Bucket[]>,
    ): TimeseriesSet {
      let extrema: TimeseriesExtrema = {
        x: { min: Infinity, max: -Infinity },
        y: { min: Infinity, max: -Infinity },
        sum: 0,
      };
      const summaries = _.map(buckets, (buckets, category) => {
        const sublabel = `${label} - ${category}`;
        const timeseries = buckets.map((b) => [_.parseInt(b.key), b.count]);
        const subExtrema = this.getExtremaFromTimeseries(timeseries);

        extrema = this.mergeExtrema(extrema, subExtrema);

        return {
          label: sublabel,
          key: key,
          category: category,
          timeseries: timeseries,
          sum: subExtrema.sum,
        };
      });

      // highest sum first
      summaries.sort((a, b) => {
        return b.sum - a.sum;
      });

      return {
        summaries: summaries,
        extrema: extrema,
      };
    },

    variableSummaryToTimeseries(
      key: string,
      label: string,
      histogram: Histogram,
    ): TimeseriesSet {
      if (histogram.categoryBuckets) {
        return this.categoryBucketsToTimeseries(
          key,
          label,
          histogram.categoryBuckets,
        );
      }
      return this.numericBucketsToTimeseries(key, label, histogram.buckets);
    },
  },
});
</script>

<style>
.sparkline-variable {
  position: relative;
  width: 100%;
  height: 32px;
  line-height: 32px;
  vertical-align: middle;
  border-bottom: 1px solid #999;
  padding: 0 8px;
}
.sparkline-variable-chart {
  float: left;
  position: relative;
  line-height: 32px;
  height: 32px;
  width: calc(100% - 276px);
}
</style>
