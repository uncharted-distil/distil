<template>
  <div v-observe-visibility="visibilityChanged">
    <svg
      v-if="isLoaded"
      ref="svg"
      class="line-chart-row"
      @click.stop="onClick"
    ></svg>
    <div v-if="!isLoaded" v-html="spinnerHTML"></div>
    <div class="highlight-tooltip" ref="tooltip"></div>
  </div>
</template>

<script lang="ts">
import * as d3 from "d3";
import _ from "lodash";
import $ from "jquery";
import Vue from "vue";
import { circleSpinnerHTML } from "../util/spinner";
import { TimeseriesExtrema, TimeSeriesValue } from "../store/dataset/index";

const MARGIN = { top: 2, right: 16, bottom: 2, left: 16 };

export default Vue.extend({
  name: "sparkline-svg",

  props: {
    margin: {
      type: Object as () => any,
      default: () => MARGIN,
    },
    highlightPixelX: Number,
    timeseries: Array as () => TimeSeriesValue[],
    forecast: Array as () => TimeSeriesValue[],
    highlightRange: Array as () => number[],
    timeseriesExtrema: Object as () => TimeseriesExtrema,
    forecastExtrema: Object as () => TimeseriesExtrema,
    isDateTime: Boolean,
    // join last element of timeseries to first element of forecast, or display both seperately.
    joinForecast: { type: Boolean, default: false },
  },

  data() {
    return {
      zoomSparkline: false,
      isVisible: false,
      hasRendered: false,
      xScale: null,
      yScale: null,
    };
  },

  computed: {
    isLoaded(): boolean {
      return !!this.timeseries;
    },

    spinnerHTML(): string {
      return circleSpinnerHTML();
    },

    svg(): d3.Selection<SVGElement, {}, HTMLElement, any> {
      return d3.select(this.$svg);
    },

    $svg(): any {
      return this.$refs.svg as any;
    },

    min(): number {
      return this.timeseries ? d3.min(this.timeseries, (d) => d.value) : 0;
    },

    max(): number {
      return this.timeseries ? d3.max(this.timeseries, (d) => d.value) : 0;
    },

    displayForecast(): TimeSeriesValue[] {
      // Join the last element of the truth timeseries and the first element of the forecast
      // time series.  Used when not visualizing an in-sample forecast.
      if (this.joinForecast) {
        return [_.last(this.timeseries)].concat(this.forecast);
      }
      return this.forecast;
    },

    showTooltip(): boolean {
      return (
        this.highlightPixelX !== null && this.hasRendered && this.isVisible
      );
    },

    $tooltip(): any {
      const tooltip = this.$refs.tooltip as any;
      return $(tooltip);
    },
  },

  mounted() {
    Vue.nextTick(() => {
      this.injectTimeseries();
    });
  },

  watch: {
    timeseries: {
      handler() {
        if (this.isVisible && !this.hasRendered) {
          Vue.nextTick(() => {
            this.injectTimeseries();
          });
        }
      },
      deep: true,
    },

    timeseriesExtrema: {
      handler(newExtrema, oldExtrema) {
        if (this.isVisible && this.isLoaded) {
          // only redraw if it is currently visible, the data has
          // loaded
          // NOTE: there is a race condition in which `isLoaded`
          // returns true, but the svg element using `v-if="isLoaded"`
          // has not yet rendered. Use this to ensure the DOM updates
          // before attempting to inject
          if (
            oldExtrema &&
            newExtrema.x.min === oldExtrema.x.min &&
            newExtrema.x.max === oldExtrema.x.max &&
            newExtrema.y.min === oldExtrema.y.min &&
            newExtrema.y.max === oldExtrema.y.max
          ) {
            return;
          }
          Vue.nextTick(() => {
            this.injectTimeseries();
          });
        } else {
          // ensure it re-renders once it comes back into view
          this.hasRendered = false;
        }
      },
      deep: true,
    },

    forecastExtrema: {
      handler(newExtrema, oldExtrema) {
        if (this.isVisible && this.isLoaded) {
          // only redraw if it is currently visible, the data has
          // loaded
          // NOTE: there is a race condition in which `isLoaded`
          // returns true, but the svg element using `v-if="isLoaded"`
          // has not yet rendered. Use this to ensure the DOM updates
          // before attempting to inject

          if (
            oldExtrema &&
            newExtrema.x.min === oldExtrema.x.min &&
            newExtrema.x.max === oldExtrema.x.max &&
            newExtrema.y.min === oldExtrema.y.min &&
            newExtrema.y.max === oldExtrema.y.max
          ) {
            return;
          }
          Vue.nextTick(() => {
            this.injectTimeseries();
          });
        } else {
          // ensure it re-renders once it comes back into view
          this.hasRendered = false;
        }
      },
      deep: true,
    },

    highlightPixelX() {
      if (this.showTooltip) {
        const xVal = this.xScale.invert(this.highlightPixelX);
        const bisect = d3.bisector((d) => {
          return d[0];
        }).left;
        const index = bisect(this.timeseries, xVal);
        if (index >= 0 && index < this.timeseries.length) {
          const yVal = this.timeseries[index].value;
          this.$tooltip
            .css({
              left: this.highlightPixelX,
            })
            .text(yVal.toFixed(2))
            .show();
          return;
        }
      } else {
        this.$tooltip.hide();
      }
    },
  },

  methods: {
    svgBounding(): any {
      return this.$svg.getBoundingClientRect();
    },

    width(): number {
      const dims = this.svgBounding();
      return dims.width - this.margin.left - this.margin.right;
    },

    height(): number {
      const dims = this.svgBounding();
      return dims.height - this.margin.top - this.margin.bottom;
    },

    visibilityChanged(isVisible: boolean) {
      this.isVisible = isVisible;
      if (this.isVisible && !this.hasRendered) {
        Vue.nextTick(() => {
          this.injectTimeseries();
        });
      }
    },

    onClick() {
      this.zoomSparkline = true;
    },

    hideModal() {
      this.zoomSparkline = false;
    },

    clearSVG() {
      this.svg.selectAll("*").remove();
    },

    computeLayout() {
      let minX = this.timeseriesExtrema.x.min;
      let maxX = this.timeseriesExtrema.x.max;
      let minY = this.timeseriesExtrema.y.min;
      let maxY = this.timeseriesExtrema.y.max;

      if (this.forecastExtrema) {
        minX = Math.min(
          this.timeseriesExtrema.x.min,
          this.forecastExtrema.x.min,
        );
        maxX = Math.max(
          this.timeseriesExtrema.x.max,
          this.forecastExtrema.x.max,
        );
        minY = Math.min(
          this.timeseriesExtrema.y.min,
          this.forecastExtrema.y.min,
        );
        maxY = Math.max(
          this.timeseriesExtrema.y.max,
          this.forecastExtrema.y.max,
        );
      }

      this.xScale = d3
        .scaleLinear()
        .domain([minX, maxX])
        .range([0, this.width()]);

      this.yScale = d3
        .scaleLinear()
        .domain([minY, maxY])
        .range([this.height(), 0]);
    },

    injectSparkline(): boolean {
      if (!this.$svg || !this.timeseries || this.timeseries.length === 0) {
        return false;
      }

      const datum = this.timeseries.map((x) => [x.time, x.value]);

      // Define a filter for non number values.
      const filterMissingData = (d) => _.isFinite(d[1]);

      // the Sparkline
      const line = d3
        .line()
        .defined(filterMissingData)
        .x((d) => this.xScale(d[0]))
        .y((d) => this.yScale(d[1]))
        .curve(d3.curveLinear);

      // the area underneath the Sparkline
      const y0 = this.yScale(0);
      const area = d3
        .area()
        .defined(filterMissingData)
        .x((d) => this.xScale(d[0]))
        .y0(y0)
        .y1((d) => this.yScale(d[1]));

      // Graph to use a container
      const g = this.svg
        .append("g")
        .attr(
          "transform",
          `translate(${this.margin.left}, ${this.margin.top})`,
        );

      // Empty values line
      g.append("path")
        .datum(datum.filter(line.defined()))
        .attr("class", "sparkline-void")
        .attr("d", line);

      // Area underneath the line
      g.append("path")
        .datum(datum)
        .attr("class", "sparkline-area")
        .attr("d", area);

      // Sparkline.
      g.append("path")
        .datum(datum)
        .attr("class", "sparkline-timeseries")
        .attr("d", line);

      return true;
    },

    // draws a shaded rectangle
    injectHighlightRegion(): boolean {
      if (
        !this.$svg ||
        !this.highlightRange ||
        this.highlightRange.length !== 2
      ) {
        return false;
      }

      const g = this.svg.append("g").attr("class", "area-scoring");
      const x0 = this.xScale(this.highlightRange[0]);
      const x1 = this.xScale(this.highlightRange[1]);
      const translate = `translate(${this.margin.left}, ${this.margin.top})`;

      // Line to demarcate the scoring test.
      g.append("line")
        .attr("class", "sparkline-line-score")
        .attr("transform", translate)
        .attr("x1", x0)
        .attr("x2", x0)
        .attr("y1", 0)
        .attr("y2", this.height);

      // area to show the scoring test
      g.append("rect")
        .attr("class", "sparkline-area-score")
        .attr("transform", translate)
        .attr("x", x0)
        .attr("y", 0)
        .attr("width", x1 - x0)
        .attr("height", this.height);

      return true;
    },

    injectPrediction(): boolean {
      if (!this.$svg || !this.forecast || this.forecast.length === 0) {
        return false;
      }

      const line = d3
        .line()
        .x((d) => this.xScale(d[0]))
        .y((d) => this.yScale(d[1]))
        .curve(d3.curveLinear);

      const g = this.svg
        .append("g")
        .attr(
          "transform",
          `translate(${this.margin.left}, ${this.margin.top})`,
        );

      g.datum(
        this.displayForecast
          .filter((x) => !_.isNil(x.value))
          .map((x) => [x.time, x.value]),
      );

      g.append("path").attr("class", "sparkline-forecast").attr("d", line);

      return true;
    },

    formatExtremaX(extrema): string {
      if (this.isDateTime) {
        return new Date(extrema)
          .toISOString()
          .slice(0, 10)
          .replace(/-/g, "/")
          .toString();
      } else {
        const format = d3.format(".1~f");
        return format(extrema.toString());
      }
    },

    injectAxis(): boolean {
      if (!this.$svg || !this.timeseries || this.timeseries.length === 0) {
        return false;
      }

      let minX = this.timeseriesExtrema.x.min;
      let maxX = this.timeseriesExtrema.x.max;
      let minY = this.timeseriesExtrema.y.min;
      let maxY = this.timeseriesExtrema.y.max;

      if (this.forecastExtrema) {
        minX = Math.min(
          this.timeseriesExtrema.x.min,
          this.forecastExtrema.x.min,
        );
        maxX = Math.max(
          this.timeseriesExtrema.x.max,
          this.forecastExtrema.x.max,
        );
        minY = Math.min(
          this.timeseriesExtrema.y.min,
          this.forecastExtrema.y.min,
        );
        maxY = Math.max(
          this.timeseriesExtrema.y.max,
          this.forecastExtrema.y.max,
        );
      }

      const dateMinX = this.formatExtremaX(minX);
      const dateMaxX = this.formatExtremaX(maxX);

      const format = d3.format(".1~f");
      const minYFormatted = format(minY);
      const maxYFormatted = format(maxY);

      this.xScale = d3
        .scaleLinear()
        .domain([minX, maxX])
        .range([0, this.width()]);

      this.yScale = d3
        .scaleLinear()
        .domain([minY, maxY])
        .range([this.height(), 0]);

      this.xScale = d3
        .scaleLinear()
        .domain([minX, maxX])
        .range([0, this.width()]);

      this.yScale = d3
        .scaleLinear()
        .domain([minY, maxY])
        .range([this.height(), 0]);

      // Create axes
      const xAxis = d3.axisBottom(this.xScale).ticks(1);
      const yAxis = d3.axisLeft(this.yScale).ticks(1);

      // Create y-max
      this.svg
        .append("text")
        .attr("class", "sparkline-axis-title")
        .attr("x", 0)
        .attr("y", 10)
        .style("text-anchor", "start")
        .text(maxYFormatted);
      // Create y-min & x-min
      this.svg
        .append("text")
        .attr("class", "sparkline-axis-title")
        .attr("x", 0)
        .attr("y", 40)
        .style("text-anchor", "start")
        .text(`${minYFormatted} ${dateMinX}`);
      // Create x-max
      this.svg
        .append("text")
        .attr("class", "sparkline-axis-title")
        .attr("x", this.width() + 20)
        .attr("y", 40)
        .style("text-anchor", "end")
        .text(dateMaxX);
      return true;
    },

    injectTimeseries() {
      if (_.isEmpty(this.timeseries) || !this.$refs.svg) {
        return;
      }

      if (this.width() <= 0) {
        console.warn("Invalid width for line chart");
        return;
      }

      if (this.height() <= 0) {
        console.warn("Invalid height for line chart");
        return;
      }

      this.clearSVG();
      this.computeLayout();
      this.injectSparkline();
      this.injectHighlightRegion();
      this.injectPrediction();
      this.injectAxis();

      this.hasRendered = true;
    },
  },
});
</script>

<style>
svg.line-chart-row {
  position: relative;
  max-height: 40px;
  width: 100%;
}

svg.line-chart-row g {
  stroke-width: 1px;
}

svg .sparkline-axis-title {
  font-size: 11px;
  font-family: Helvetica Neue, Helvetica, Arial, sans-serif;
  fill: #000;
  text-shadow: 0 0 3px #fff, 1px -1px 1px #fff, -1px -1px 1px #fff,
    0 -1px 1px #fff;
}

.highlight-tooltip {
  position: absolute;
  top: 0;
  pointer-events: none;
}

.sparkline-timeseries {
  fill: none;
  stroke: var(--gray-700);
  stroke-width: 1.5;
}

.sparkline-area {
  fill: var(--gray-400);
  stroke: none;
}

.sparkline-void {
  fill: none;
  stroke: var(--gray-500);
  stroke-dasharray: 2 5;
}

.sparkline-forecast {
  stroke: var(--blue); /* rgb(2, 117, 216); */
  fill: none;
}

.sparkline-confidence {
  stroke: var(--blue); /* rgb(2, 117, 216); */
  opacity: 0.3;
  stroke: none;
}

.sparkline-line-score {
  stroke: var(--yellow);
  stroke-width: 2;
  opacity: 0.5;
}

.sparkline-area-score {
  fill: var(--yellow);
  opacity: 0.04;
  stroke: none;
}
</style>
