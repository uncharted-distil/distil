<template>
  <div class="sparkline-container">
    <svg ref="svg" class="line-chart-big"></svg>
  </div>
</template>

<script lang="ts">
import * as d3 from "d3";
import _ from "lodash";
import Vue from "vue";
import { TimeSeriesValue } from "../store/dataset/index";

const MARGIN = { top: 24, right: 48, bottom: 16, left: 48 };

export default Vue.extend({
  name: "sparkline-chart",

  props: {
    margin: {
      type: Object as () => any,
      default: () => MARGIN,
    },
    timeseries: Array as () => TimeSeriesValue[],
    forecast: Array as () => TimeSeriesValue[],
    highlightRange: Array as () => number[],
    xAxisTitle: String,
    yAxisTitle: String,
    xAxisDateTime: Boolean,
    joinForecast: Boolean,
  },

  data() {
    return {
      xScale: null,
      yScale: null,
    };
  },

  computed: {
    svg(): d3.Selection<SVGElement, {}, HTMLElement, any> {
      const $svg = this.$refs.svg as any;
      return d3.select($svg);
    },
    width(): number {
      const $svg = this.$refs.svg as any;
      const dims = $svg.getBoundingClientRect();
      return dims.width - this.margin.left - this.margin.right;
    },
    height(): number {
      const $svg = this.$refs.svg as any;
      const dims = $svg.getBoundingClientRect();
      return dims.height - this.margin.top - this.margin.bottom;
    },
    minX(): number {
      const min = d3.min(this.timeseries, (d) => d.time);
      return this.forecast
        ? Math.min(
            min,
            d3.min(this.forecast, (d) => d.time),
          )
        : min;
    },
    maxX(): number {
      const max = d3.max(this.timeseries, (d) => d.time);
      return this.forecast
        ? Math.max(
            max,
            d3.max(this.forecast, (d) => d.time),
          )
        : max;
    },
    minY(): number {
      const timeSeriesMin = d3.min(this.timeseries, (d) => d.value);
      const forecastMin = this.forecast
        ? d3.min(this.forecast, (d) => d.value)
        : NaN;
      const confidenceMin = this.forecast
        ? d3.min(this.forecast, (d) => d.confidenceLow)
        : NaN;
      return d3.min([timeSeriesMin, forecastMin, confidenceMin], (d) => d);
    },
    maxY(): number {
      const timeSeriesMax = d3.max(this.timeseries, (d) => d.value);
      const forecastMax = this.forecast
        ? d3.max(this.forecast, (d) => d.value)
        : NaN;
      const confidenceMax = this.forecast
        ? d3.max(this.forecast, (d) => d.confidenceHigh)
        : NaN;
      return d3.max([timeSeriesMax, forecastMax, confidenceMax], (d) => d);
    },
    displayForecast(): TimeSeriesValue[] {
      // Join the last element of the truth timeseries and the first element of the forecast
      // time series.  Used when not visualizing an in-sample forecast.
      if (this.joinForecast) {
        const last = _.clone(_.last(this.timeseries));
        last.confidenceLow = last.value;
        last.confidenceHigh = last.value;
        return [last, ...this.forecast];
      }
      return this.forecast;
    },
  },

  mounted() {
    setTimeout(() => {
      this.draw();
    });
  },

  methods: {
    clearSVG() {
      this.svg.selectAll("*").remove();
    },

    injectAxes() {
      const minX = this.xAxisDateTime ? new Date(this.minX) : this.minX;
      const maxX = this.xAxisDateTime ? new Date(this.maxX) : this.maxX;

      if (this.xAxisDateTime) {
        this.xScale = d3
          .scaleTime()
          .domain([minX, maxX])
          .range([0, this.width]);
      } else {
        this.xScale = d3
          .scaleLinear()
          .domain([minX, maxX])
          .range([0, this.width]);
      }

      this.yScale = d3
        .scaleLinear()
        .domain([this.minY, this.maxY])
        .range([this.height, 0]);

      // Create axes
      const xAxis = d3.axisBottom(this.xScale).ticks(10);
      const yAxis = d3.axisLeft(this.yScale).ticks(5);

      // Create x-axis
      const svgXAxis = this.svg
        .append("g")
        .attr("class", "x axis")
        .attr("transform", `translate(${this.margin.left}, ${this.height})`)
        .call(xAxis);

      svgXAxis
        .append("text")
        .attr("class", "axis-title")
        .attr("x", this.width / 2)
        .attr("y", this.margin.bottom)
        .attr("dy", this.margin.bottom)
        .style("text-anchor", "middle")
        .text(this.xAxisTitle);

      // Create y-axis
      const svgYAxis = this.svg
        .append("g")
        .attr("class", "y axis")
        .attr("transform", `translate(${this.margin.left}, 0)`)
        .call(yAxis);

      svgYAxis
        .append("text")
        .attr("class", "axis-title")
        .attr("transform", "rotate(-90)")
        .attr("x", -(this.height / 2))
        .attr("y", -this.margin.left + 8)
        .style("text-anchor", "middle")
        .text(this.yAxisTitle);
    },

    injectTimeseries() {
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
      const y0 = this.yScale(this.minY);
      const area = d3
        .area()
        .defined(filterMissingData)
        .x((d) => this.xScale(d[0]))
        .y0(y0)
        .y1((d) => this.yScale(d[1]));

      // Graph to use a container
      const g = this.svg
        .append("g")
        .attr("transform", `translate(${this.margin.left}, 0)`)
        .attr("class", "line-chart");

      // Empty values line
      g.append("path")
        .datum(datum.filter(line.defined()))
        .attr("class", "line-void")
        .attr("d", line);

      // Area underneath the line
      g.append("path").datum(datum).attr("class", "line-area").attr("d", area);

      // Sparkline.
      g.append("path")
        .datum(datum)
        .attr("class", "line-timeseries")
        .attr("d", line);
    },

    injectForecast() {
      const line = d3
        .line()
        .x((d) => this.xScale(d[0]))
        .y((d) => this.yScale(d[1]))
        .curve(d3.curveLinear);

      const g = this.svg
        .append("g")
        .attr("transform", `translate(${this.margin.left}, 0)`)
        .attr("class", "line-chart");

      g.datum(
        this.displayForecast
          .filter((x) => !_.isNil(x.value))
          .map((x) => [x.time, x.value]),
      );

      g.append("path")
        .attr("fill", "none")
        .attr("class", "line-forecast")
        .attr("d", line);
    },

    injectConfidence(): boolean {
      const area = d3
        .area<[number, number, number]>()
        .x((d) => this.xScale(d[0]))
        .y0((d) => this.yScale(d[1]))
        .y1((d) => this.yScale(d[2]));

      const g = this.svg
        .append("g")
        .attr("transform", `translate(${this.margin.left}, 0)`);

      g.datum(
        this.displayForecast
          .filter(
            (x) => !_.isNil(x.confidenceHigh) && !_.isNil(x.confidenceLow),
          )
          .map((x) => [x.time, x.confidenceHigh, x.confidenceLow]),
      );

      g.append("path").attr("class", "line-confidence").attr("d", area);

      return true;
    },

    injectTimeRangeHighligh() {
      if (!this.highlightRange || this.highlightRange.length !== 2) {
        return;
      }

      const g = this.svg.append("g").attr("class", "area-scoring");
      const x0 = this.xScale(this.highlightRange[0]);
      const x1 = this.xScale(this.highlightRange[1]);
      const translate = `translate(${this.margin.left}, 0)`;

      // Line to demarcate the scoring test.
      g.append("line")
        .attr("class", "line-score")
        .attr("transform", translate)
        .attr("x1", x0)
        .attr("x2", x0)
        .attr("y1", 0)
        .attr("y2", this.height);

      // area to show the scoring test
      g.append("rect")
        .attr("class", "area-score")
        .attr("transform", translate)
        .attr("x", x0)
        .attr("y", 0)
        .attr("width", x1 - x0)
        .attr("height", this.height);
    },

    draw() {
      if (_.isEmpty(this.timeseries)) {
        return;
      }

      if (this.width <= 0) {
        console.warn("Invalid width for line chart", this.width);
        return;
      }

      if (this.height <= 0) {
        console.warn("Invalid height for line chart", this.height);
        return;
      }

      this.clearSVG();
      this.injectAxes();
      this.injectTimeseries();
      if (this.forecast) {
        this.injectTimeRangeHighligh();
        this.injectConfidence();
        this.injectForecast();
      }
    },
  },
});
</script>

<style>
.sparkline-container {
  position: relative;
}

.line-chart-big {
  position: relative;
  width: 100%;
  border: 1px solid rgba(0, 0, 0, 0);
}

.line-chart {
  stroke-width: 2px;
}

.line-timeseries {
  fill: none;
  stroke: var(--gray-700);
}

.line-area {
  fill: var(--gray-400);
  stroke: none;
}

.line-void {
  fill: none;
  stroke: var(--gray-500);
  stroke-dasharray: 2 5;
}

.line-forecast {
  stroke: var(--blue);
}

.line-confidence {
  fill: var(--blue);
  opacity: 0.3;
}

.line-score {
  stroke: var(--yellow);
  stroke-width: 2;
  opacity: 0.5;
}

.area-score {
  fill: var(--yellow);
  opacity: 0.04;
  stroke: none;
}

.axis {
  stroke-width: 1px;
}

.axis-title {
  fill: #000;
  stroke-width: 1px;
  font-size: 12px;
}
</style>
