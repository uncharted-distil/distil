<template>

	<div class="select-timeseries-view" ref="timeseries"
		@mousemove="mouseMove"
		@mouseleave="mouseLeave"
		@wheel="scroll">
		<div class="timeseries-row-header">
			<div class="timeseries-var-col pad-top"><b>VARIABLES</b></div>
			<div class="timeseries-min-col pad-top"><b>MIN</b></div>
			<div class="timeseries-max-col pad-top"><b>MAX</b></div>
			<div class="timeseries-chart-axis">
				<template v-if="hasData">
					<svg ref="svg" class="axis"></svg>
				</template>
			</div>
		</div>
		<div class="timeseries-rows">
			<div v-if="isTimeseriesAnalysis">
				<div v-for="timeseries in timeseriesVariableSummaries">
					<sparkline-variable
						:label="timeseries.label"
						:timeseries="timeseries.timeseries"
						:timeseries-extrema="timeseriesVariableExtrema(timeseries.key)"
						:highlight-pixel-x="highlightPixelX">
					</sparkline-variable>
				</div>
			</div>

			<div v-if="!isTimeseriesAnalysis">
				<div v-for="item in items">
					<sparkline-row
						:x-col="timeseriesGrouping.properties.xCol"
						:y-col="timeseriesGrouping.properties.yCol"
						:timeseries-col="timeseriesGrouping.idCol"
						:timeseries-id="item[timeseriesGrouping.idCol]"
						:timeseries-extrema="timeseriesRowExtrema"
						:highlight-pixel-x="highlightPixelX">
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
import { RowSelection, HighlightRoot } from '../store/highlights/index';
import { TableRow, TableColumn, TimeseriesExtrema, Variable, VariableSummary, Grouping } from '../store/dataset/index';
import { getters as routeGetters } from '../store/route/module';
import { getters as datasetGetters } from '../store/dataset/module';
import { updateHighlightRoot } from '../util/highlights';
import { getTimeseriesGroupingsFromFields } from '../util/data';
import { isTimeType } from '../util/types';

const TICK_SIZE = 8;
const SELECTED_TICK_SIZE = 18;
const MIN_PIXEL_WIDTH = 32;

export default Vue.extend({
	name: 'select-timeseries-view',

	components: {
		SparklineRow,
		SparklineVariable,
	},

	props: {
		instanceName: String as () => string,
		includedActive: Boolean as () => boolean
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

		isTimeseriesAnalysis(): boolean {
			return !!routeGetters.getRouteTimeseriesAnalysis(this.$store);
		},

		timeseriesAnalysisVar(): string {
			return routeGetters.getRouteTimeseriesAnalysis(this.$store);
		},

		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},

		variableSummaries(): VariableSummary[] {
			const training = routeGetters.getTrainingVariableSummaries(this.$store);
			const target = routeGetters.getTargetVariableSummaries(this.$store);
			return training.concat(target);
		},

		items(): TableRow[] {
			return this.includedActive ? datasetGetters.getIncludedTableDataItems(this.$store) : datasetGetters.getExcludedTableDataItems(this.$store);
		},

		fields(): Dictionary<TableColumn> {
			return this.includedActive ? datasetGetters.getIncludedTableDataFields(this.$store) : datasetGetters.getExcludedTableDataFields(this.$store);
		},

		timeseriesGrouping(): Grouping {
			return getTimeseriesGroupingsFromFields(this.variables, this.fields)[0];
		},

		hasData(): boolean {
			if (this.isTimeseriesAnalysis && this.variableSummaries.length > 0) {
				return true;
			}
			return !!this.timeseriesExtrema;
		},

		isDateScale(): boolean {
			let timeVar = null;
			if (this.isTimeseriesAnalysis) {
				timeVar = this.variables.find(v => v.colName === this.timeseriesAnalysisVar);
			} else {
			const grouping = this.timeseriesGrouping;
				timeVar = this.variables.find(v => v.colName === grouping.properties.xCol);
			}
			return (timeVar && isTimeType(timeVar.colType));
		},

		timeseriesVariableSummaries(): any[] {
			let timeseries = [];
			this.variableSummaries.forEach(v => {
				if (v.categoryBuckets) {
					const categories = [];
					_.forIn(v.categoryBuckets, (buckets, category) => {
						categories.push({
							label: `${v.label} - ${category}`,
							key: v.key,
							timeseries: buckets.map(b => [ _.parseInt(b.key), b.count ]),
							xMin: v.extrema.min,
							xMax: v.extrema.max,
							yMin: _.minBy(buckets, d => d.count).count,
							yMax: _.maxBy(buckets, d => d.count).count,
							sum: _.sumBy(buckets, d => d.count)
						});
					});
					// highest sum first
					categories.sort((a, b) => { return b.sum - a.sum; });
					timeseries = timeseries.concat(categories);
				} else {
					timeseries.push({
						label: v.label,
						key: v.key,
						timeseries: v.buckets.map(b => [ _.parseInt(b.key), b.count ]),
						xMin: v.extrema.min,
						xMax: v.extrema.max,
						yMin: _.minBy(v.buckets, d => d.count).count,
						yMax: _.maxBy(v.buckets, d => d.count).count,
						sum: _.sumBy(v.buckets, d => d.count)
					});
				}
			});
			return timeseries;
		},

		timeseriesVarsMinX(): number {
			if (this.timeseriesVariableSummaries.length === 0) {
				return null;
			}
			// take first, all vars share same x axis
			return this.timeseriesVariableSummaries[0].xMin;
		},

		timeseriesVarsMaxX(): number {
			if (this.timeseriesVariableSummaries.length === 0) {
				return null;
			}
			// take first, all vars share same x axis
			return this.timeseriesVariableSummaries[0].xMax;
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

		highlightRoot(): HighlightRoot {
			return routeGetters.getDecodedHighlightRoot(this.$store);
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
			return this.highlightRoot &&
				(this.isTimeseriesAnalysis || this.highlightRoot.key === this.timeseriesGrouping.idCol) &&
				this.highlightRoot.value.from !== undefined &&
				this.highlightRoot.value.to !== undefined;
		},

		microMin(): number {
			if (this.selectedMicroMin !== null) {
				return this.selectedMicroMin;
			}
			if (this.isTimeseriesViewHighlight) {
				return this.highlightRoot.value.from;
			}
			return this.timeseriesMinX;
		},

		microMax(): number {
			if (this.selectedMicroMax !== null) {
				return this.selectedMicroMax;
			}
			if (this.isTimeseriesViewHighlight) {
				return this.highlightRoot.value.to;
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
		},

		microMinValue(): any {
			return this.microMin;
			// return this.isDateScale ? new Date(this.microMin * 1000) : this.microMin;
		},

		microMaxValue(): any {
			return this.microMax;
			// return this.isDateScale ? new Date(this.microMax * 1000) : this.microMax;
		}
	},

	methods: {

		timeseriesVariableExtrema(variableKey: string): TimeseriesExtrema {
			let yMin = Infinity;
			let yMax = -Infinity;
			this.timeseriesVariableSummaries.forEach(v => {
				if (v.key === variableKey) {
					yMin = Math.min(yMin, v.yMin);
					yMax = Math.max(yMax, v.yMax);
				}
			});
			return {
				x: {
					min: this.microMin,
					max: this.microMax
				},
				y: {
					min: yMin,
					max: yMax
				}
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

			// if (this.isDateScale) {
			// 	this.microScale = d3.scaleTime()
			// 		.domain([this.microMinValue, this.microMaxValue])
			// 		.range([0, this.width]);
			//
			// } else {
				this.microScale = d3.scaleLinear()
					.domain([this.microMinValue, this.microMaxValue])
					.range([0, this.width]);
			// }

			this.svg.append('g')
				.attr('class', 'micro-axis')
				.attr('transform', `translate(${this.margin.left}, ${-this.margin.bottom + this.height - TICK_SIZE * 2})`)
				.call(d3.axisBottom(this.microScale).tickFormat(this.axisFormat()));

			this.svg.append('rect')
				.attr('class', 'axis-selection-rect')
				.attr('x', this.macroScale(this.microMinValue) + this.margin.left)
				.attr('y', this.margin.top + TICK_SIZE * 2)
				.attr('width', this.macroScale(this.microMaxValue) - this.macroScale(this.microMinValue))
				.attr('height', SELECTED_TICK_SIZE);

			this.svg.select('.axis-selection').raise();

			this.attachTranslationHandlers();
		},
		injectSVG() {

			if (!this.hasData) {
				return;
			}

			this.clearSVG();

			// if (this.isDateScale) {
			// 	this.macroScale = d3.scaleTime()
			// 		.domain([new Date(this.timeseriesMinX * 1000), new Date(this.timeseriesMaxX * 1000)])
			// 		.range([0, this.width]);
			//
			// } else {
				this.macroScale = d3.scaleLinear()
					.domain([this.timeseriesMinX, this.timeseriesMaxX])
					.range([0, this.width]);
			// }

			this.svg.append('g')
				.attr('class', 'macro-axis')
				.attr('transform', `translate(${this.margin.left}, ${this.margin.top + SELECTED_TICK_SIZE + TICK_SIZE * 2})`)
				.call(d3.axisTop(this.macroScale).tickFormat(this.axisFormat()));

			this.microRangeSelection = d3.axisTop(this.macroScale)
				.tickSize(SELECTED_TICK_SIZE)
				.tickValues([
					this.microMinValue,
					this.microMaxValue
				])
				.tickFormat(this.axisFormat());

			this.svg.append('g')
				.attr('class', 'axis-selection')
				.attr('transform', `translate(${this.margin.left}, ${this.margin.top + SELECTED_TICK_SIZE + TICK_SIZE * 2})`)
				.call(this.microRangeSelection);

			this.injectMicroAxis();

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
					const maxPx = this.macroScale(this.microMaxValue);
					const clampedPx = Math.min(px, maxPx - MIN_PIXEL_WIDTH);
					this.selectedMicroMin = this.macroScale.invert(clampedPx);
					this.repositionMicroMin(this.selectedMicroMin);
				} else {
					const minPx = this.macroScale(this.microMinValue);
					const clampedPx = Math.max(px, minPx + MIN_PIXEL_WIDTH);
					this.selectedMicroMax = this.macroScale.invert(clampedPx);
					this.repositionMicroMax(this.selectedMicroMax);
				}

				this.injectMicroAxis();
			};

			const dragended = (d, index, elem) => {
				updateHighlightRoot(this.$router, {
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
					this.selectedMicroMin = this.microMinValue;
				}
				if (this.selectedMicroMax === null) {
					this.selectedMicroMax = this.microMaxValue;
				}

				const maxDelta = this.timeseriesMaxX - this.selectedMicroMax;
				const minDelta = this.timeseriesMinX - this.selectedMicroMin;

				const delta = this.macroScale.invert(d3.event.dx);
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
				updateHighlightRoot(this.$router, {
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
.select-timeseries-view {
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

</style>
