<template>
	<div
		v-observe-visibility="visibilityChanged"
		v-bind:class="{'is-hidden': !isVisible}">
		<svg v-if="isLoaded" ref="svg" class="line-chart-row" @click.stop="onClick"></svg>
		<div v-if="!isLoaded" v-html="spinnerHTML"></div>
		<div class="highlight-tooltip" ref="tooltip"></div>
	</div>
</template>

<script lang="ts">

import * as d3 from 'd3';
import _ from 'lodash';
import $ from 'jquery';
import Vue from 'vue';
import { circleSpinnerHTML } from '../util/spinner';
import { TimeseriesExtrema } from '../store/dataset/index';

const INJECT_DEBOUNCE = 200;

export default Vue.extend({
	name: 'sparkline-svg',

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
		timeseriesExtrema: {
			type: Object as () => TimeseriesExtrema
		}
	},
	data() {
		const component = this as any;
		return {
			zoomSparkline: false,
			isVisible: false,
			hasRendered: false,
			xScale: null,
			yScale: null,
			debouncedInjection: _.debounce(() => {
				Vue.nextTick(() => {
					component.injectTimeseries();
				});
			}, INJECT_DEBOUNCE),
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
		},
		showTooltip(): boolean {
			return this.highlightPixelX !== null && this.hasRendered && this.isVisible;
		},
		$tooltip(): any {
			const tooltip = this.$refs.tooltip as any;
			return $(tooltip);
		}
	},

	watch: {
		timeseries() {
			console.log('watch timeseries');
			if (this.isVisible && !this.hasRendered) {
				this.debouncedInjection();
			}
		},
		timeseriesExtrema: {
			handler(newExtrema, oldExtrema) {
				console.log('watch timeseriesExtrema');
				if (this.isVisible && this.isLoaded) {
					// only redraw if it is currently visible, the data has
					// loaded
					// NOTE: there is a race condition in which `isLoaded`
					// returns true, but the svg element using `v-if="isLoaded"`
					// has not yet rendered. Use this to ensure the DOM updates
					// before attempting to inject

					if (newExtrema.x.min === oldExtrema.x.min &&
						newExtrema.x.max === oldExtrema.x.max &&
						newExtrema.y.min === oldExtrema.y.min &&
						newExtrema.y.max === oldExtrema.y.max) {
						return;
					}
					this.debouncedInjection();
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
					this.$tooltip.css({
						left: this.highlightPixelX
					}).text(yVal.toFixed(2)).show();
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
				this.debouncedInjection();
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

			if (!this.$svg || !this.timeseries || this.timeseries.length === 0) {
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
		injectPrediction() {

			if (!this.$svg || !this.forecast || this.forecast.length === 0) {
				return;
			}

			const line = d3.line()
			.x(d => this.xScale(d[0]))
			.y(d => this.yScale(d[1]))
			.curve(d3.curveLinear);

			const g = this.svg.append('g')
			.attr('transform', `translate(${this.margin.left}, ${this.margin.top})`);

			g.datum(this.forecast);

			g.append('path')
			.attr('fill', 'none')
			.attr('class', 'line')
			.attr('stroke', '#00c6e1')
			.attr('d', line);

		},
		injectTimeseries() {
			if (_.isEmpty(this.timeseries) || !this.$refs.svg) {
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
			this.injectPrediction();

			this.hasRendered = true;
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

.is-hidden {
	visibility: hidden;
}

.highlight-tooltip {
	position: absolute;
	top: 0;
	pointer-events: none;
}

</style>
