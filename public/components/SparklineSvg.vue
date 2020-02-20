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
import { TimeseriesExtrema } from "../store/dataset/index";

export default Vue.extend({
  name: "sparkline-svg",

  props: {
    margin: {
      type: Object as () => any,
      default: () => ({
        top: 2,
        right: 16,
        bottom: 2,
        left: 16
      })
    },
    highlightPixelX: {
      type: Number as () => number
    },
    timeseries: Array as () => number[][],
    forecast: Array as () => number[][],
    highlightRange: Array as () => number[],
    timeseriesExtrema: {
      type: Object as () => TimeseriesExtrema
    },
    forecastExtrema: {
      type: Object as () => TimeseriesExtrema
    }
  },
  data() {
    return {
      zoomSparkline: false,
      isVisible: false,
      hasRendered: false,
      xScale: null,
      yScale: null
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
    width(): number {
      const dims = this.$svg.getBoundingClientRect();
      return dims.width - this.margin.left - this.margin.right;
    },
    height(): number {
      const dims = this.$svg.getBoundingClientRect();
      return dims.height - this.margin.top - this.margin.bottom;
    },
    min(): number {
      return this.timeseries ? d3.min(this.timeseries, d => d[1]) : 0;
    },
    max(): number {
      return this.timeseries ? d3.max(this.timeseries, d => d[1]) : 0;
    },
    showTooltip(): boolean {
      return (
        this.highlightPixelX !== null && this.hasRendered && this.isVisible
      );
    },
    $tooltip(): any {
      const tooltip = this.$refs.tooltip as any;
      return $(tooltip);
    }
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
      deep: true
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
      deep: true
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
      deep: true
    },
    highlightPixelX() {
      if (this.showTooltip) {
        const xVal = this.xScale.invert(this.highlightPixelX);
        const bisect = d3.bisector(d => {
          return d[0];
        }).left;
        const index = bisect(this.timeseries, xVal);
        if (index >= 0 && index < this.timeseries.length) {
          const yVal = this.timeseries[index][1];
          this.$tooltip
            .css({
              left: this.highlightPixelX
            })
            .text(yVal.toFixed(2))
            .show();
          return;
        }
      } else {
        this.$tooltip.hide();
      }
    }
  },

  methods: {
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
    injectSparkline(): boolean {
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
          this.forecastExtrema.x.min
        );
        maxX = Math.max(
          this.timeseriesExtrema.x.max,
          this.forecastExtrema.x.max
        );
        minY = Math.min(
          this.timeseriesExtrema.y.min,
          this.forecastExtrema.y.min
        );
        maxY = Math.max(
          this.timeseriesExtrema.y.max,
          this.forecastExtrema.y.max
        );
      }

      this.xScale = d3
        .scaleLinear()
        .domain([minX, maxX])
        .range([0, this.width]);

      this.yScale = d3
        .scaleLinear()
        .domain([minY, maxY])
        .range([this.height, 0]);

      const line = d3
        .line()
        .x(d => this.xScale(d[0]))
        .y(d => this.yScale(d[1]))
        .curve(d3.curveLinear);

      const g = this.svg
        .append("g")
        .attr(
          "transform",
          `translate(${this.margin.left}, ${this.margin.top})`
        );

      g.datum(this.timeseries);

      g.append("path")
        .attr("fill", "none")
        .attr("class", "line")
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

      this.svg
        .append("rect")
        .attr("transform", `translate(${this.margin.left}, ${this.margin.top})`)
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

      return true;
    },
    injectPrediction(): boolean {
      if (!this.$svg || !this.forecast || this.forecast.length === 0) {
        return false;
      }

      const line = d3
        .line()
        .x(d => this.xScale(d[0]))
        .y(d => this.yScale(d[1]))
        .curve(d3.curveLinear);

      const g = this.svg
        .append("g")
        .attr(
          "transform",
          `translate(${this.margin.left}, ${this.margin.top})`
        );

      g.datum(this.forecast);

      g.append("path")
        .attr("fill", "none")
        .attr("class", "line")
        .attr("stroke", "#00c6e188")
        .attr("d", line);

      return true;
    },
    injectTimeseries() {
      if (_.isEmpty(this.timeseries) || !this.$refs.svg) {
        return;
      }

      if (this.width <= 0) {
        console.warn("Invalid width for line chart");
        return;
      }

      if (this.height <= 0) {
        console.warn("Invalid height for line chart");
        return;
      }

      this.clearSVG();
      this.injectSparkline();
      this.injectHighlightRegion();
      this.injectPrediction();

      this.hasRendered = true;
    }
  }
});
</script>

<style>
svg.line-chart-row {
  position: relative;
  max-height: 40px;
  width: 100%;
}

svg.line-chart-row g {
  stroke: #666;
  stroke-width: 2px;
}

.highlight-tooltip {
  position: absolute;
  top: 0;
  pointer-events: none;
}
</style>
