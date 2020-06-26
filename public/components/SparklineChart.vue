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

export default Vue.extend({
  name: "sparkline-chart",
  props: {
    margin: {
      type: Object as () => any,
      default: () => ({
        top: 24,
        right: 48,
        bottom: 16,
        left: 48
      })
    },
    timeseries: {
      type: Array as () => TimeSeriesValue[]
    },
    forecast: {
      type: Array as () => TimeSeriesValue[]
    },
    highlightRange: {
      type: Array as () => number[]
    },
    xAxisTitle: {
      type: String as () => string
    },
    yAxisTitle: {
      type: String as () => string
    },
    xAxisDateTime: {
      type: Boolean as () => boolean
    },
    joinForecast: {
      type: Boolean as () => boolean
    }
  },
  data() {
    return {
      xScale: null,
      yScale: null
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
      const min = d3.min(this.timeseries, d => d.time);
      return this.forecast
        ? Math.min(
            min,
            d3.min(this.forecast, d => d.time)
          )
        : min;
    },
    maxX(): number {
      const max = d3.max(this.timeseries, d => d.time);
      return this.forecast
        ? Math.max(
            max,
            d3.max(this.forecast, d => d.time)
          )
        : max;
    },
    minY(): number {
      const timeSeriesMin = d3.min(this.timeseries, d => d.value);
      const forecastMin = d3.min(this.forecast, d => d.value);
      const confidenceMin = d3.min(this.forecast, d => d.confidenceLow);
      return d3.min([timeSeriesMin, forecastMin, confidenceMin], d => d);
    },
    maxY(): number {
      const timeSeriesMax = d3.max(this.timeseries, d => d.value);
      const forecastMax = d3.max(this.forecast, d => d.value);
      const confidenceMax = d3.max(this.forecast, d => d.confidenceHigh);
      return d3.max([timeSeriesMax, forecastMax, confidenceMax], d => d);
    },
    displayForecast(): TimeSeriesValue[] {
      // Join the last element of the truth timeseries and the first element of the forecast
      // time series.  Used when not visualizing an in-sample forecast.
      if (this.joinForecast) {
        return [_.last(this.timeseries)].concat(this.forecast);
      }
      return this.forecast;
    }
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
      const line = d3
        .line()
        .x(d => this.xScale(d[0]))
        .y(d => this.yScale(d[1]))
        .curve(d3.curveLinear);

      const g = this.svg
        .append("g")
        .attr("transform", `translate(${this.margin.left}, 0)`)
        .attr("class", "line-chart");

      g.datum(this.timeseries.map(x => [x.time, x.value]));

      g.append("path")
        .attr("fill", "none")
        .attr("class", "line-timeseries")
        .attr("d", line);
    },
    injectForecast() {
      const line = d3
        .line()
        .x(d => this.xScale(d[0]))
        .y(d => this.yScale(d[1]))
        .curve(d3.curveLinear);

      const g = this.svg
        .append("g")
        .attr("transform", `translate(${this.margin.left}, 0)`)
        .attr("class", "line-chart");

      g.datum(this.displayForecast.map(x => [x.time, x.value]));

      g.append("path")
        .attr("fill", "none")
        .attr("class", "line-forecast")
        .attr("d", line);
    },
    injectConfidence(): boolean {
      const area = d3
        .area<[number, number, number]>()
        .x(d => this.xScale(d[0]))
        .y0(d => this.yScale(d[1]))
        .y1(d => this.yScale(d[2]));

      const g = this.svg
        .append("g")
        .attr("transform", `translate(${this.margin.left}, 0)`);

      g.datum(
        this.displayForecast.map(x => [
          x.time,
          x.confidenceHigh,
          x.confidenceLow
        ])
      );

      g.append("path")
        .attr("class", "line-confidence")
        .attr("d", area);

      return true;
    },
    injectTimeRangeHighligh() {
      if (!this.highlightRange || this.highlightRange.length !== 2) {
        return;
      }

      this.svg
        .append("rect")
        .attr("class", "area-score")
        .attr("transform", `translate(${this.margin.left}, 0)`)
        .attr("x", this.xScale(this.highlightRange[0]))
        .attr("y", 0)
        .attr(
          "width",
          this.xScale(this.highlightRange[1]) -
            this.xScale(this.highlightRange[0])
        )
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
      if (this.forecast) {
        this.injectTimeRangeHighligh();
      }
      this.injectTimeseries();
      if (this.forecast) {
        this.injectConfidence();
        this.injectForecast();
      }
    }
  }
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
  stroke: rgb(200, 200, 200);
}

.line-forecast {
  stroke: rgb(2, 117, 216);
}

.line-confidence {
  fill: rgb(2, 117, 216);
  opacity: 0.3;
}

.area-score {
  fill: rgb(200, 200, 200);
  opacity: 0.2;
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
