<template>
	<div class="sparkline-row" v-observe-visibility="visibilityChanged" v-bind:class="{'is-hidden': !isVisible}">
		<div class="timeseries-var-col">{{timeseriesId}}</div>
		<div class="timeseries-min-col">{{min.toFixed(2)}}</div>
		<div class="timeseries-max-col">{{max.toFixed(2)}}</div>
		<div class="timeseries-chart-col">
			<svg v-if="isLoaded" ref="svg" class="line-chart-row" @click.stop="onClick"></svg>
			<div v-if="!isLoaded" v-html="spinnerHTML"></div>
			<div class="highlight-tooltip" ref="tooltip"></div>
		</div>
	</div>
</template>

<script lang="ts">

import * as d3 from 'd3';
import _ from 'lodash';
import $ from 'jquery';
import Vue from 'vue';
import { Dictionary } from '../util/dict';
import { circleSpinnerHTML } from '../util/spinner';
import { getters as routeGetters } from '../store/route/module';
import { TimeseriesExtrema } from '../store/dataset/index';
import { getters as datasetGetters, actions as datasetActions } from '../store/dataset/module';

export default Vue.extend({
	name: 'sparkline-row',

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
		xCol: String as () => string,
		yCol: String as () => string,
		timeseriesCol: String as () => string,
		timeseriesId: String as () => string,
		timeseriesExtrema: {
			type: Object as () => TimeseriesExtrema
		}
	},
	data() {
		return {
			zoomSparkline: false,
			entry: null,
			isVisible: false,
			hasRendered: false,
			hasRequested: false,
			xAxisTitle: '',
			yAxisTitle: '',
			xScale: null,
			yScale: null
		};
	},
	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		isLoaded(): boolean {
			return !!datasetGetters.getTimeseries(this.$store)[this.dataset][this.timeseriesId];
		},
		timeseries(): number[][] {
			return datasetGetters.getTimeseries(this.$store)[this.dataset][this.timeseriesId];
		},
		spinnerHTML(): string {
			return circleSpinnerHTML();
		},
		svg(): d3.Selection<SVGElement, {}, HTMLElement, any> {
			return  d3.select(this.$svg);
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
		}
	},

	watch: {
		timeseriesExtrema: {
			handler() {
				if (this.isVisible && this.isLoaded) {
					// only redraw if it is currently visible, the data has
					// loaded
					// NOTE: there is a race condition in which `isLoaded`
					// returns true, but the svg element using `v-if="isLoaded"`
					// has not yet rendered use this to ensure the DOM updates
					// before attempting to inject
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
			const tooltip = this.$refs.tooltip as any;
			if (this.highlightPixelX !== null && this.hasRendered && this.isVisible) {
				const xVal = this.xScale.invert(this.highlightPixelX);
				const bisect = d3.bisector(d => {
					return d[0];
				}).left;
				const index = bisect(this.timeseries, xVal);
				if (index >= 0 && index < this.timeseries.length) {
					const yVal = this.timeseries[index][1];
					$(tooltip).css({
						left: this.highlightPixelX,
						visibility: 'visible'
					}).text(yVal.toFixed(2));
					return;
				}
			}
			$(tooltip).css('visibility', 'hidden');
		}
	},

	methods: {
		visibilityChanged(isVisible: boolean) {
			this.isVisible = isVisible;
			if (this.isVisible && !this.hasRequested) {
				this.requestTimeseries();
				return;
			}
			if (this.isVisible && this.hasRequested && !this.hasRendered) {
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
			this.svg.selectAll('*').remove();
		},
		injectSparkline() {

			if (!this.$svg) {
				return;
			}

			const minX = this.timeseriesExtrema.x.min;
			const maxX = this.timeseriesExtrema.x.max;
			const minY = this.timeseriesExtrema.y.min;
			const maxY = this.timeseriesExtrema.y.max;

			this.xScale = d3.scaleLinear()
				.domain([minX, maxX])
				.range([0, this.width]);

			this.yScale = d3.scaleLinear()
				.domain([minY, maxY])
				.range([this.height, 0]);

			const line = d3.line()
				.x(d => this.xScale(d[0]))
				.y(d => this.yScale(d[1]))
				.curve(d3.curveLinear);

			const g = this.svg.append('g')
				.attr('transform', `translate(${this.margin.left}, ${this.margin.top})`);

			g.datum(this.timeseries);

			g.append('path')
				.attr('fill', 'none')
				.attr('class', 'line')
				.attr('d', line);
		},
		injectTimeseries() {
			if (_.isEmpty(this.timeseries)) {
				return;
			}

			if (this.width <= 0) {
				console.warn('Invalid width for line chart');
				return;
			}

			if (this.height <= 0) {
				console.warn('Invalid height for line chart');
				return;
			}

			this.clearSVG();
			this.injectSparkline();

			this.hasRendered = true;
		},
		requestTimeseries() {
			this.hasRequested = true;
			datasetActions.fetchTimeseries(this.$store, {
				dataset: this.dataset,
				xColName: this.xCol,
				yColName: this.yCol,
				timeseriesColName: this.timeseriesCol,
				timeseriesID: this.timeseriesId
			}).then(() => {
				if (this.isVisible) {
					Vue.nextTick(() => {
						this.injectTimeseries();
					});
				}
			});
		}
	}

});
</script>

<style>

svg.line-chart-row {
	position: relative;
	max-height: 32px;
	width: 100%;
}

svg.line-chart-row g {
	stroke: #666;
	stroke-width: 2px;
}

/*
svg.line-chart-row:hover g {
	stroke: #00c6e1;
}
*/

.sparkline-row {
	position: relative;
	width: 100%;
	height: 32px;
	line-height: 32px;
	vertical-align: middle;
	border-bottom: 1px solid #999;
	padding: 0 8px;
}

.is-hidden {
	visibility: hidden;
}

.highlight-tooltip {
	position: absolute;
	top: 0;
	pointer-events: none;
}

</style>
