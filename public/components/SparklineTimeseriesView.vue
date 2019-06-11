<template>

	<div class="sparkline-timeseries-view" ref="timeseries"
		@mousemove="mouseMove"
		@mouseleave="mouseLeave"
		@wheel="scroll">
		<div class="timeseries-row-header">
			<div class="timeseries-var-col pad-top"><b>VARIABLES</b></div>
			<div class="timeseries-min-col pad-top"><b>MIN</b></div>
			<div class="timeseries-max-col pad-top"><b>MAX</b></div>
			<div class="timeseries-chart-axis" v-bind:class="{'has-prediction': showPredicted}">
				<template v-if="hasData">
					<svg ref="svg" class="axis"></svg>
				</template>
			</div>
			<div v-if="showPredicted" class="timeseries-prediction-col pad-top"><b>PREDICTION</b></div>
		</div>
		<div class="timeseries-rows" v-if="hasData">
			<div v-if="isTimeseriesAnalysis">
				<!-- <div v-for="timeseries in predictedTimeseriesVariableSummaries">
					<sparkline-variable
						:label="timeseries.label"
						:timeseries="timeseries.timeseries"
						:forecast="timeseries.forecast"
						:timeseries-extrema="timeseriesVariableExtrema[timeseries.key]"
						:highlight-pixel-x="highlightPixelX">
					</sparkline-variable>
				</div> -->

				<div v-for="summary in variableSummaries" :key="summary.key">
					<sparkline-variable
						:summary="summary"
						:highlight-pixel-x="highlightPixelX"
						:min-x="microMin"
						:max-x="microMax">
					</sparkline-variable>
				</div>

			</div>
			<div v-if="!isTimeseriesAnalysis">
				<div class="sparkline-row-container" v-for="item in items">
					<sparkline-row
						:x-col="timeseriesGrouping.properties.xCol"
						:y-col="timeseriesGrouping.properties.yCol"
						:timeseries-col="timeseriesGrouping.idCol"
						:timeseries-id="item[timeseriesGrouping.idCol]"
						:timeseries-extrema="timeseriesRowExtrema"
						:highlight-pixel-x="highlightPixelX"
						:prediction="getPrediction(item)">
					</sparkline-row>
				</div>
			</div>
		</div>
		<div class="vertical-line"></div>
	</div>

</template>

<script lang="ts">

import * as d3 from 'd3';
import _ from 'lodash';
import $ from 'jquery';
import moment from 'moment';
import Vue from 'vue';
import SparklineRow from './SparklineRow';
import SparklineVariable from './SparklineVariable';
import { Dictionary } from '../util/dict';
import { Filter } from '../util/filters';
import { TableRow, TableColumn, TimeseriesExtrema, Variable, Histogram, Bucket, VariableSummary, Grouping, RowSelection, Highlight } from '../store/dataset/index';
import { getters as routeGetters } from '../store/route/module';
import { getters as datasetGetters } from '../store/dataset/module';
import { getters as resultsGetters } from '../store/results/module';
import { getters as solutionGetters } from '../store/solutions/module';
import { updateHighlight } from '../util/highlights';
import { getTimeseriesGroupingsFromFields } from '../util/data';
import { isTimeType } from '../util/types';
import { getSolutionIndex } from '../util/solutions';

const TICK_SIZE = 8;
const SELECTED_TICK_SIZE = 18;
const MIN_PIXEL_WIDTH = 32;

export default Vue.extend({
	name: 'sparkline-timeseries-view',

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
		disableHighlighting: Boolean as () => boolean
	},

	data() {
		return {
			margin: {
				top: 2,
				right: 16,
				bottom: 2,
				left: 16
			},
			macroScale: null,
			microScale: null,
			microRangeSelection: null,
			selectedMicroMin: null,
			selectedMicroMax: null,
			highlightPixelX: null
		};
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		isTimeseriesAnalysis(): boolean {
			return !!routeGetters.getRouteTimeseriesAnalysis(this.$store);
		},

		timeseriesAnalysisVar(): string {
			return routeGetters.getRouteTimeseriesAnalysis(this.$store);
		},

		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},

		hasPredictedCol(): boolean {
			return !!this.predictedCol;
		},

		showPredicted(): boolean {
			return this.hasPredictedCol && !this.isTimeseriesAnalysis;
		},

		timeseriesGrouping(): Grouping {
			// TODO: support more than one grouping
			const groupings = getTimeseriesGroupingsFromFields(this.variables, this.fields);
			return groupings[0];
		},

		hasData(): boolean {
			if (this.isTimeseriesAnalysis && this.variableSummaries.length > 0) {
				return true;
			}
			return this.timeseriesGrouping && !!this.timeseriesExtrema;
		},

		timeseriesAnalysisVariable(): Variable {
			return datasetGetters.getTimeseriesAnalysisVariable(this.$store);
		},

		isDateScale(): boolean {
			let timeVar = null;
			if (this.isTimeseriesAnalysis) {
				timeVar = datasetGetters.getTimeseriesAnalysisVariable(this.$store);
			} else {
				const grouping = this.timeseriesGrouping;
				timeVar = this.variables.find(v => v.colName === grouping.properties.xCol);
			}
			return (timeVar && isTimeType(timeVar.colType));
		},

		// resultTargetSummary(): VariableSummary {
		// 	return resultsGetters.getTargetSummary(this.$store);
		// },
		//
		// predictedSummaries(): VariableSummary[] {
		// 	const summaries = resultsGetters.getPredictedSummaries(this.$store);
		// 	const solutions = solutionGetters.getRelevantSolutions(this.$store);
		// 	return solutions.map(solution => {
		// 		return _.find(summaries, summary => {
		// 			return summary.solutionId === solution.solutionId;
		// 		});
		// 	}).filter(summary => !!summary); // remove errors
		// },

		timeseriesVarsMinX(): number {
			if (!this.timeseriesAnalysisVariable) {
				return null;
			}
			return this.timeseriesAnalysisVariable.min;
		},

		timeseriesVarsMaxX(): number {
			if (!this.timeseriesAnalysisVariable) {
				return null;
			}
			return this.timeseriesAnalysisVariable.max;
		},

		timeseriesExtrema(): TimeseriesExtrema {
			return datasetGetters.getTimeseriesExtrema(this.$store)[this.dataset];
		},

		timeseriesMinX(): number {
			if (this.isTimeseriesAnalysis) {
				return this.timeseriesVarsMinX;
			}
			if (this.timeseriesExtrema) {
				return this.timeseriesExtrema.x.min;
			}
			return 0;
		},

		timeseriesMaxX(): number {
			if (this.isTimeseriesAnalysis) {
				return this.timeseriesVarsMaxX;
			}
			if (this.timeseriesExtrema) {
				return this.timeseriesExtrema.x.max;
			}
			return 1;
		},

		filters(): Filter[] {
			if (this.includedActive) {
				return this.invertFilters(routeGetters.getDecodedFilters(this.$store));
			}
			return routeGetters.getDecodedFilters(this.$store);
		},

		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},

		highlight(): Highlight {
			return routeGetters.getDecodedHighlight(this.$store);
		},

		timeseriesRowExtrema(): TimeseriesExtrema {
			return {
				x: {
					min: this.microMin,
					max: this.microMax
				},
				y: {
					min: this.timeseriesExtrema ? this.timeseriesExtrema.y.min : 0,
					max: this.timeseriesExtrema ? this.timeseriesExtrema.y.max : 1
				}
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
			return this.highlight &&
				(this.isTimeseriesAnalysis || this.highlight.key === this.timeseriesGrouping.idCol) &&
				this.highlight.value.from !== undefined &&
				this.highlight.value.to !== undefined;
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
			return this.$timeseries.find('.vertical-line');
		},

		$axis(): any {
			return this.$timeseries.find('.timeseries-chart-axis');
		}
	},

	methods: {

		getPrediction(row: TableRow): any {
			if (!this.showPredicted) {
				return null;
			}
			return {
				value: row[this.predictedCol],
				isCorrect: row[this.predictedCol] === row[this.target]
			};
		},

		invertFilters(filters: Filter[]): Filter[] {
			// TODO: invert filters
			return filters;
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
					top: chartScroll
				});
				this.highlightPixelX = relX - chartLeft - this.margin.left;
			} else {
				this.$line.hide();
				this.highlightPixelX = null;
			}
		},

		scroll(event) {
			const chartScroll = this.$timeseries.parent().scrollTop();
			this.$line.css('top', chartScroll);
		},

		injectMicroAxis() {

			this.svg.select('.micro-axis').remove();
			this.svg.select('.axis-selection-rect').remove();

			this.microScale = d3.scaleLinear()
				.domain([this.microMin, this.microMax])
				.range([0, this.width]);

			this.svg.append('g')
				.attr('class', 'micro-axis')
				.attr('transform', `translate(${this.margin.left}, ${-this.margin.bottom + this.height - TICK_SIZE * 2})`)
				.call(d3.axisBottom(this.microScale).tickFormat(this.axisFormat()));

			this.svg.append('rect')
				.attr('class', 'axis-selection-rect')
				.attr('x', this.macroScale(this.microMin) + this.margin.left)
				.attr('y', this.margin.top + TICK_SIZE * 2)
				.attr('width', this.macroScale(this.microMax) - this.macroScale(this.microMin))
				.attr('height', SELECTED_TICK_SIZE);

			this.svg.select('.axis-selection').raise();

			this.attachTranslationHandlers();
		},

		injectSVG() {

			if (!this.hasData || !this.$refs.svg) {
				return;
			}

			this.clearSVG();

			this.macroScale = d3.scaleLinear()
				.domain([this.timeseriesMinX, this.timeseriesMaxX])
				.range([0, this.width]);

			this.svg.append('g')
				.attr('class', 'macro-axis')
				.attr('transform', `translate(${this.margin.left}, ${this.margin.top + SELECTED_TICK_SIZE + TICK_SIZE * 2})`)
				.call(d3.axisTop(this.macroScale).tickFormat(this.axisFormat()));

			if (!this.disableHighlighting) {
				// highlighting axis / controls

				this.microRangeSelection = d3.axisTop(this.macroScale)
					.tickSize(SELECTED_TICK_SIZE)
					.tickValues([
						this.microMin,
						this.microMax
					])
					.tickFormat(this.axisFormat());

				this.svg.append('g')
					.attr('class', 'axis-selection')
					.attr('transform', `translate(${this.margin.left}, ${this.margin.top + SELECTED_TICK_SIZE + TICK_SIZE * 2})`)
					.call(this.microRangeSelection);

				this.injectMicroAxis();
			}

			this.attachScalingHandlers();
		},
		repositionMicroMin(xVal: any) {
			const px = this.macroScale(xVal);
			const $lower = this.svg.select('.axis-selection .tick');
			$lower.attr('transform', `translate(${px}, 0)`);
			$lower.select('text').text(this.axisFormat()(xVal));
		},
		repositionMicroMax(xVal: any) {
			const px = this.macroScale(xVal);
			const $upper = this.svg.select('.axis-selection .tick:last-child');
			$upper.attr('transform', `translate(${px}, 0)`);
			$upper.select('text').text(this.axisFormat()(xVal));
		},
		repositionMicroRange(xMin: any, xMax: any) {
			const minPx = this.macroScale(xMin);
			const maxPx = this.macroScale(xMax);
			const widthPx = maxPx - minPx;
			const $range = this.svg.select('.axis-selection-rect');
			$range .attr('x', minPx + this.margin.left);
			$range .attr('width', widthPx);
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
					key: this.isTimeseriesAnalysis ? this.timeseriesAnalysisVar : this.timeseriesGrouping.idCol,
					value: {
						from: this.microMin,
						to: this.microMax
					}
				});
			};

			this.svg.selectAll('.axis-selection .tick')
				.call(d3.drag()
					.on('start', dragstarted)
					.on('drag', dragged)
					.on('end', dragended)
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
					key: this.isTimeseriesAnalysis ? this.timeseriesAnalysisVar : this.timeseriesGrouping.idCol,
					value: {
						from: this.microMin,
						to: this.microMax
					}
				});
			};

			this.svg.selectAll('.axis-selection-rect')
				.call(d3.drag()
					.on('start', dragstarted)
					.on('drag', dragged)
					.on('end', dragended));
		},
		clearSVG() {
			this.svg.selectAll('*').remove();
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
					return m.format('MMM D');
				}
				return v.toFixed(2);
			};
		},
		isCorrect(item: TableRow): boolean {
			return item[this.predictedCol] === item[this.target];
		}

	},

	watch: {
		timeseriesVariableSummaries: {
			handler() {
				Vue.nextTick(() => {
					this.injectSVG();
				});
			},
			deep: true
		},
		timeseriesExtrema: {
			handler() {
				Vue.nextTick(() => {
					this.injectSVG();
				});
			},
			deep: true
		}
	},

	mounted() {
		this.injectSVG();
	}

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
