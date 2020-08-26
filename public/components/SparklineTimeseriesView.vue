<template>
  <div
    class="sparkline-timeseries-view"
    ref="timeseries"
    @mousemove="mouseMove"
    @mouseleave="mouseLeave"
    @wheel="scroll"
  >
    <div class="timeseries-row-header">
      <div class="timeseries-var-col pad-top"><b>VARIABLES</b></div>
      <div class="timeseries-min-col pad-top"><b>MIN</b></div>
      <div class="timeseries-max-col pad-top"><b>MAX</b></div>
      <div
        class="timeseries-chart-axis"
        v-bind:class="{ 'has-prediction': showPredicted }"
      >
        <template>
          <svg ref="svg" class="axis"></svg>
        </template>
      </div>
      <div v-if="showPredicted" class="timeseries-prediction-col pad-top">
        <b>PREDICTION</b>
      </div>
    </div>
    <div class="timeseries-rows" v-if="hasData">
      <div
        class="sparkline-row-container"
        v-for="item in items"
        :key="item[timeseriesGrouping.idCol].value"
      >
        <sparkline-row
          :x-col="timeseriesGrouping.xCol"
          :y-col="timeseriesGrouping.yCol"
          :timeseries-col="timeseriesGrouping.idCol"
          :timeseries-id="item[timeseriesGrouping.idCol].value"
          :timeseries-extrema="
            timeseriesRowLocalExtrema(item[timeseriesGrouping.idCol].value)
          "
          :highlight-pixel-x="highlightPixelX"
          :prediction="getPrediction(item)"
          :solution-id="solutionId"
        >
        </sparkline-row>
      </div>
    </div>
    <div class="vertical-line"></div>
  </div>
</template>

<script lang="ts">
import * as d3 from "d3";
import _ from "lodash";
import $ from "jquery";
import moment from "moment";
import Vue from "vue";
import SparklineRow from "./SparklineRow";
import SparklineVariable from "./SparklineVariable";
import { Dictionary } from "../util/dict";
import { Filter } from "../util/filters";
import {
  TableRow,
  TableColumn,
  TimeseriesExtrema,
  TimeSeriesValue,
  Variable,
  Histogram,
  Bucket,
  VariableSummary,
  Grouping,
  RowSelection,
  Highlight,
  TaskTypes,
  TimeseriesGrouping,
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as resultsGetters } from "../store/results/module";
import { getters as requestGetters } from "../store/requests/module";
import { updateHighlight } from "../util/highlights";
import { getTimeseriesGroupingsFromFields } from "../util/data";
import { isTimeType } from "../util/types";
import { getSolutionIndex } from "../util/solutions";
import { getSolutionResultSummary } from "../util/summaries";

const TICK_SIZE = 8;
const SELECTED_TICK_SIZE = 18;
const MIN_PIXEL_WIDTH = 32;

export default Vue.extend({
  name: "sparkline-timeseries-view",

  components: {
    SparklineRow,
    SparklineVariable,
  },

  props: {
    instanceName: String as () => string,
    includedActive: Boolean as () => boolean,
    variableSummaries: Array as () => VariableSummary[],
    items: Array as () => TableRow[],
    fields: Object as () => Dictionary<TableColumn>,
    predictedCol: String as () => string,
    disableHighlighting: Boolean as () => boolean,
  },

  data() {
    return {
      margin: {
        top: 2,
        right: 16,
        bottom: 2,
        left: 16,
      },
      macroScale: null,
      microScale: null,
      microRangeSelection: null,
      selectedMicroMin: null,
      selectedMicroMax: null,
      highlightPixelX: null,
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    solutionId(): string {
      return routeGetters.getRouteSolutionId(this.$store);
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    hasPredictedCol(): boolean {
      return !!this.predictedCol;
    },

    isForecasting(): boolean {
      return routeGetters
        .getRouteTask(this.$store)
        .includes(TaskTypes.FORECASTING);
    },

    showPredicted(): boolean {
      return this.hasPredictedCol && !this.isForecasting;
    },

    timeseriesGrouping(): TimeseriesGrouping {
      // TODO: support more than one grouping
      const groupings = getTimeseriesGroupingsFromFields(
        this.variables,
        this.fields,
      );
      return groupings[0];
    },

    hasData(): boolean {
      return this.timeseriesGrouping && !!this.timeseriesExtrema;
    },

    isDateScale(): boolean {
      const grouping = this.timeseriesGrouping;
      const timeVar = this.variables.find((v) => v.colName === grouping.xCol);
      return timeVar && isTimeType(timeVar.colType);
    },

    resultTargetSummary(): VariableSummary {
      return resultsGetters.getTargetSummary(this.$store);
    },

    predictedSummaries(): VariableSummary[] {
      return requestGetters
        .getRelevantSolutions(this.$store)
        .map((solution) => getSolutionResultSummary(solution.solutionId))
        .filter((summary) => !!summary); // remove errors
    },

    timeseriesExtrema(): TimeseriesExtrema {
      return datasetGetters.getTimeseriesExtrema(this.$store)[this.dataset];
    },

    timeseriesMinX(): number {
      if (this.timeseriesExtrema) {
        return this.timeseriesExtrema.x.min;
      }
      return 0;
    },

    timeseriesMaxX(): number {
      if (this.timeseriesExtrema) {
        return this.timeseriesExtrema.x.max;
      }
      return 1;
    },

    filters(): Filter[] {
      return routeGetters.getDecodedFilters(this.$store);
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    timeseriesRowGlobalExtrema(): TimeseriesExtrema {
      return {
        x: {
          min: this.microMin,
          max: this.microMax,
        },
        y: {
          min: this.timeseriesExtrema ? this.timeseriesExtrema.y.min : 0,
          max: this.timeseriesExtrema ? this.timeseriesExtrema.y.max : 1,
        },
      };
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

    svg(): d3.Selection<SVGElement, {}, HTMLElement, any> {
      const $svg = this.$refs.svg as any;
      return d3.select($svg);
    },

    isTimeseriesViewHighlight(): boolean {
      // ignore any highlights unless they are range highlights
      return (
        this.highlight &&
        this.highlight.key === this.timeseriesGrouping.idCol &&
        this.highlight.value.from !== undefined &&
        this.highlight.value.to !== undefined
      );
    },

    microMin(): number {
      if (this.selectedMicroMin !== null) {
        return this.selectedMicroMin;
      }
      if (this.isTimeseriesViewHighlight) {
        return this.highlight.value.from;
      }
      return this.timeseriesMinX;
    },

    microMax(): number {
      if (this.selectedMicroMax !== null) {
        return this.selectedMicroMax;
      }
      if (this.isTimeseriesViewHighlight) {
        return this.highlight.value.to;
      }
      return this.timeseriesMaxX;
    },

    $timeseries(): any {
      const timeseries = this.$refs.timeseries as any;
      return $(timeseries);
    },

    $line(): any {
      return this.$timeseries.find(".vertical-line");
    },

    $axis(): any {
      return this.$timeseries.find(".timeseries-chart-axis");
    },
  },

  methods: {
    getTimeseries(timeseriesId: string): TimeSeriesValue[] {
      const timeseries = datasetGetters.getTimeseries(this.$store);
      const datasets = timeseries[this.dataset];
      if (!datasets) {
        return [];
      }
      return datasets[timeseriesId] ? datasets[timeseriesId] : [];
    },

    getPredictedTimeseries(timeseriesId: string): TimeSeriesValue[] {
      const timeseries = resultsGetters.getPredictedTimeseries(this.$store);
      const solutions = timeseries[this.solutionId];
      if (!solutions) {
        return [];
      }
      return solutions[timeseriesId] ? solutions[timeseriesId] : [];
    },

    getPredictedForecasts(timeseriesId: string): TimeSeriesValue[] {
      const forecasts = resultsGetters.getPredictedForecasts(this.$store);
      const solutions = forecasts[this.solutionId];
      if (!solutions) {
        return [];
      }
      return solutions[timeseriesId] ? solutions[timeseriesId] : [];
    },

    timeseriesRowLocalExtrema(timeseriesId: string): TimeseriesExtrema {
      let yValues = null;

      if (this.solutionId) {
        const timeseries = this.getPredictedTimeseries(timeseriesId);
        const forecasts = this.getPredictedForecasts(timeseriesId);
        const both = timeseries.concat(forecasts);
        yValues = both.map((v) => v.value);
      } else {
        const timeseries = this.getTimeseries(timeseriesId);
        yValues = timeseries.map((v) => v.value);
      }

      const yMin: number = _.min(yValues);
      const yMax: number = _.max(yValues);

      return {
        x: {
          min: this.microMin,
          max: this.microMax,
        },
        y: {
          min: yMin !== undefined ? yMin : 0,
          max: yMax !== undefined ? yMax : 1,
        },
      };
    },

    getPrediction(row: TableRow): any {
      if (!this.showPredicted) {
        return null;
      }
      return {
        value: row[this.predictedCol],
        isCorrect: row[this.predictedCol] === row[this.target],
      };
    },

    mouseLeave() {
      this.$line.hide();
      this.highlightPixelX = null;
    },

    mouseMove(event) {
      const parentOffset = this.$timeseries.offset();
      const chartBounds = this.$axis.offset();
      const chartWidth = this.$axis.width();
      const chartScroll = this.$timeseries.parent().scrollTop();

      const relX = event.pageX - parentOffset.left;
      const chartLeft = chartBounds.left - parentOffset.left;

      if (relX >= chartLeft && relX <= chartLeft + chartWidth) {
        this.$line.show();
        this.$line.css({
          left: relX,
          top: chartScroll,
        });
        this.highlightPixelX = relX - chartLeft - this.margin.left;
      } else {
        this.$line.hide();
        this.highlightPixelX = null;
      }
    },

    scroll(event) {
      const chartScroll = this.$timeseries.parent().scrollTop();
      this.$line.css("top", chartScroll);
    },

    injectMicroAxis() {
      this.svg.select(".micro-axis").remove();
      this.svg.select(".axis-selection-rect").remove();

      this.microScale = d3
        .scaleLinear()
        .domain([this.microMin, this.microMax])
        .range([0, this.width]);

      this.svg
        .append("g")
        .attr("class", "micro-axis")
        .attr(
          "transform",
          `translate(${this.margin.left}, ${
            -this.margin.bottom + this.height - TICK_SIZE * 2
          })`,
        )
        .call(d3.axisBottom(this.microScale).tickFormat(this.axisFormat()));

      this.svg
        .append("rect")
        .attr("class", "axis-selection-rect")
        .attr("x", this.macroScale(this.microMin) + this.margin.left)
        .attr("y", this.margin.top + TICK_SIZE * 2)
        .attr(
          "width",
          this.macroScale(this.microMax) - this.macroScale(this.microMin),
        )
        .attr("height", SELECTED_TICK_SIZE);

      this.svg.select(".axis-selection").raise();

      this.attachTranslationHandlers();
    },

    injectSVG() {
      if (!this.hasData || !this.$refs.svg) {
        return;
      }

      this.clearSVG();

      this.macroScale = d3
        .scaleLinear()
        .domain([this.timeseriesMinX, this.timeseriesMaxX])
        .range([0, this.width]);

      this.svg
        .append("g")
        .attr("class", "macro-axis")
        .attr(
          "transform",
          `translate(${this.margin.left}, ${
            this.margin.top + SELECTED_TICK_SIZE + TICK_SIZE * 2
          })`,
        )
        .call(d3.axisTop(this.macroScale).tickFormat(this.axisFormat()));

      if (!this.disableHighlighting) {
        // highlighting axis / controls

        this.microRangeSelection = d3
          .axisTop(this.macroScale)
          .tickSize(SELECTED_TICK_SIZE)
          .tickValues([this.microMin, this.microMax])
          .tickFormat(this.axisFormat());

        this.svg
          .append("g")
          .attr("class", "axis-selection")
          .attr(
            "transform",
            `translate(${this.margin.left}, ${
              this.margin.top + SELECTED_TICK_SIZE + TICK_SIZE * 2
            })`,
          )
          .call(this.microRangeSelection);

        this.injectMicroAxis();
      }

      this.attachScalingHandlers();
    },
    repositionMicroMin(xVal: any) {
      const px = this.macroScale(xVal);
      const $lower = this.svg.select(".axis-selection .tick");
      $lower.attr("transform", `translate(${px}, 0)`);
      $lower.select("text").text(this.axisFormat()(xVal));
    },
    repositionMicroMax(xVal: any) {
      const px = this.macroScale(xVal);
      const $upper = this.svg.select(".axis-selection .tick:last-child");
      $upper.attr("transform", `translate(${px}, 0)`);
      $upper.select("text").text(this.axisFormat()(xVal));
    },
    repositionMicroRange(xMin: any, xMax: any) {
      const minPx = this.macroScale(xMin);
      const maxPx = this.macroScale(xMax);
      const widthPx = maxPx - minPx;
      const $range = this.svg.select(".axis-selection-rect");
      $range.attr("x", minPx + this.margin.left);
      $range.attr("width", widthPx);
    },
    attachScalingHandlers() {
      const dragstarted = (d, index, elem) => {
        this.highlightPixelX = null;
        this.$line.hide();
      };

      const dragged = (d, index, elem) => {
        const minX = 0;
        const maxX = this.width;

        const x = d3.event.x;
        const px = _.clamp(x, minX, maxX);

        if (index === 0) {
          const maxPx = this.macroScale(this.microMax);
          const clampedPx = Math.min(px, maxPx - MIN_PIXEL_WIDTH);
          this.selectedMicroMin = this.macroScale.invert(clampedPx);
          this.repositionMicroMin(this.selectedMicroMin);
        } else {
          const minPx = this.macroScale(this.microMin);
          const clampedPx = Math.max(px, minPx + MIN_PIXEL_WIDTH);
          this.selectedMicroMax = this.macroScale.invert(clampedPx);
          this.repositionMicroMax(this.selectedMicroMax);
        }

        this.injectMicroAxis();
      };

      const dragended = (d, index, elem) => {
        updateHighlight(this.$router, {
          context: this.instanceName,
          dataset: this.dataset,
          key: this.timeseriesGrouping.idCol,
          value: {
            from: this.microMin,
            to: this.microMax,
          },
        });
      };

      this.svg
        .selectAll(".axis-selection .tick")
        .call(
          d3
            .drag()
            .on("start", dragstarted)
            .on("drag", dragged)
            .on("end", dragended),
        );
    },
    attachTranslationHandlers() {
      const dragstarted = (d, index, elem) => {
        this.highlightPixelX = null;
        this.$line.hide();
      };

      const dragged = (d, index, elem) => {
        if (this.selectedMicroMin === null) {
          this.selectedMicroMin = this.microMin;
        }
        if (this.selectedMicroMax === null) {
          this.selectedMicroMax = this.microMax;
        }

        const maxDelta = this.timeseriesMaxX - this.selectedMicroMax;
        const minDelta = this.timeseriesMinX - this.selectedMicroMin;

        const delta = this.macroScale.invert(d3.event.dx) - this.timeseriesMinX;
        const clampedDelta = _.clamp(delta, minDelta, maxDelta);

        this.selectedMicroMin += clampedDelta;
        this.selectedMicroMax += clampedDelta;

        // update rect
        this.repositionMicroRange(this.selectedMicroMin, this.selectedMicroMax);

        // update ticks
        this.repositionMicroMin(this.selectedMicroMin);
        this.repositionMicroMax(this.selectedMicroMax);

        this.injectMicroAxis();
      };

      const dragended = (d, index, elem) => {
        updateHighlight(this.$router, {
          context: this.instanceName,
          dataset: this.dataset,
          key: this.timeseriesGrouping.idCol,
          value: {
            from: this.microMin,
            to: this.microMax,
          },
        });
      };

      this.svg
        .selectAll(".axis-selection-rect")
        .call(
          d3
            .drag()
            .on("start", dragstarted)
            .on("drag", dragged)
            .on("end", dragended),
        );
    },
    clearSVG() {
      this.svg.selectAll("*").remove();
    },
    axisFormat() {
      return (v) => {
        if (this.isDateScale) {
          let m = null;
          if (_.isString(v)) {
            m = moment(v);
          } else {
            m = moment.unix(v);
          }
          // TODO: format based on total span
          return m.format("MMM D");
        }
        return v.toFixed(2);
      };
    },
    isCorrect(item: TableRow): boolean {
      return item[this.predictedCol] === item[this.target];
    },
  },

  watch: {
    variableSummaries: {
      handler() {
        Vue.nextTick(() => {
          this.injectSVG();
        });
      },
      deep: true,
    },
    timeseriesExtrema: {
      handler() {
        Vue.nextTick(() => {
          this.injectSVG();
        });
      },
      deep: true,
    },
  },

  mounted() {
    this.injectSVG();
  },
});
</script>

<style>
svg.axis {
  position: relative;
  max-height: 64px;
  width: 100%;
}
.sparkline-timeseries-view {
  position: relative;
  flex: 1;
  height: inherit;
}
.timeseries-row-header {
  position: relative;
  width: 100%;
  height: 64px;
  line-height: 32px;
  border-bottom: 1px solid #999;
  padding: 0 8px;
  background-color: #fff;
}
.timeseries-rows {
  position: relative;
  height: calc(100% - 64px);
  z-index: 0;
  font-size: 10px;
  font-weight: bold;
  overflow-x: hidden;
  overflow-y: auto;
}
.timeseries-var-col {
  float: left;
  position: relative;
  line-height: 32px;
  height: 32px;
  width: 156px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.timeseries-min-col {
  float: left;
  position: relative;
  line-height: 32px;
  height: 32px;
  width: 48px;
}
.timeseries-max-col {
  float: left;
  position: relative;
  line-height: 32px;
  height: 32px;
  width: 48px;
}
.timeseries-chart-axis {
  float: left;
  position: relative;
  line-height: 32px;
  height: 64px;
  width: calc(100% - 276px);
}
.timeseries-chart-axis.has-prediction {
  width: calc(100% - 372px);
}
.timeseries-prediction-col {
  float: left;
  position: relative;
  line-height: 32px;
  height: 32px;
  width: 96px;
  text-align: center;
}
.pad-top {
  margin-top: 32px;
}
.vertical-line {
  position: absolute;
  display: none;
  top: 0;
  left: 0;
  width: 1px;
  height: 100%;
  border-left: 1px solid #00c6e1;
  box-shadow: 0px 0px 5px #00c6e1;
  pointer-events: none;
}
.axis-selection {
}

.axis-selection .tick {
  cursor: pointer;
  stroke-width: 3;
}

.axis-selection path.domain {
  visibility: hidden;
}

.axis-selection-rect {
  fill: #00c6e1;
  opacity: 0.25;
  cursor: pointer;
}

.sparkline-row-container {
  position: relative;
}
.sparkline-prediction {
  position: absolute;
  top: 4px;
  right: 4px;
}
</style>
