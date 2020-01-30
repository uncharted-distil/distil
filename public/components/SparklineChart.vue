<template>
  <div class="sparkline-container">
    <svg ref="svg" class="line-chart-big"></svg>
  </div>
</template>

<script lang="ts">
import * as d3 from "d3";
import _ from "lodash";
import Vue from "vue";

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
      type: Array as () => number[][]
    },
    forecast: {
      type: Array as () => number[][]
    },
    highlightRange: {
      type: Array as () => number[]
    },
    xAxisTitle: {
      type: String as () => string
    },
    yAxisTitle: {
      type: String as () => string
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
      const min = d3.min(this.timeseries, d => d[0]);
      return this.forecast
        ? Math.min(min, d3.min(this.forecast, d => d[0]))
        : min;
    },
    maxX(): number {
      const max = d3.max(this.timeseries, d => d[0]);
      return this.forecast
        ? Math.max(max, d3.max(this.forecast, d => d[0]))
        : max;
    },
    minY(): number {
      const min = d3.min(this.timeseries, d => d[1]);
      return this.forecast
        ? Math.min(min, d3.min(this.forecast, d => d[1]))
        : min;
    },
    maxY(): number {
      const max = d3.max(this.timeseries, d => d[1]);
      return this.forecast
        ? Math.max(max, d3.max(this.forecast, d => d[1]))
        : max;
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
      this.xScale = d3
        .scaleLinear()
        .domain([this.minX, this.maxX])
        .range([0, this.width]);

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

      g.datum(this.timeseries);

      g.append("path")
        .attr("fill", "none")
        .attr("class", "line")
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

      g.datum(this.forecast);

      g.append("path")
        .attr("fill", "none")
        .attr("class", "line")
        .attr("stroke", "#00c6e1")
        .attr("d", line);
    },
    injectTimeRangeHighligh() {
      if (!this.highlightRange || this.highlightRange.length !== 2) {
        return;
      }

      this.svg
        .append("rect")
        .attr("transform", `translate(${this.margin.left}, 0)`)
        .attr("fill", "#00ffff44")
        .attr("stroke", "none")
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
      this.injectTimeseries();
      if (this.forecast) {
        this.injectForecast();
        this.injectTimeRangeHighligh();
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
  stroke: #666;
  stroke-width: 2px;
}

.axis {
  stroke-width: 1px;
}

.axis-title {
  fill: #000;
  stroke-width: 1px;
}
</style>
