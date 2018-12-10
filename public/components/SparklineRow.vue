<template>
	<div class="sparkline-container" v-observe-visibility="visibilityChanged" v-bind:class="{'is-hidden': !isVisible}">
		<svg v-if="isLoaded" ref="svg" class="line-chart" @click.stop="onClick" ></svg>
		<div v-if="!isLoaded" v-html="spinnerHTML"></div>
	</div>
</template>

<script lang="ts">

import * as d3 from 'd3';
import _ from 'lodash';
import Vue from 'vue';
import { Dictionary } from '../util/dict';
import { circleSpinnerHTML } from '../util/spinner';
import { getters as routeGetters } from '../store/route/module';
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
		timeSeriesUrl: {
			type: String as () => string
		},
		min: number,
		max: number
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
		files(): Dictionary<any> {
			return datasetGetters.getFiles(this.$store);
		},
		isLoaded(): boolean {
			return this.files[this.timeSeriesUrl];
		},
		timeseries(): number[][] {
			return this.files[this.timeSeriesUrl];
		},
		spinnerHTML(): string {
			return circleSpinnerHTML();
		},
		svg(): d3.Selection<SVGElement, {}, HTMLElement, any> {
			const $svg = this.$refs.svg as any;
			return  d3.select($svg);
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
				this.injectTimeseries();
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
			const timeseries = this.timeseries;

			this.xScale = d3.scalePoint()
				.range([0, this.width]);
			this.xScale.domain(timeseries.map(d => d[0]));

			const min = d3.min(timeseries, d => d[1]);
			const max = d3.max(timeseries, d => d[1]);

			this.yScale = d3.scaleLinear()
				.domain([min, max])
				.range([this.height, 0]);

			const line = d3.line()
				.x(d => this.xScale(d[0]))
				.y(d => this.yScale(d[1]))
				.curve(d3.curveLinear);

			const className = 'line-chart';
			const g = this.svg.append('g')
				.attr('transform', `translate(${this.margin.left}, ${this.margin.top})`)
				.attr('class', className);

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
				url: this.timeSeriesUrl
			}).then(() => {
				if (this.isVisible) {
					this.injectTimeseries();
				}
			});
		}
	}

});
</script>

<style>

svg.line-chart {
	position: relative;
	max-height: 32px;
	width: 100%;
	border: 1px solid rgba(0,0,0,0);
}

svg.line-chart g {
	stroke: #666;
	stroke-width: 2px;
}

svg.line-chart:hover g {
	stroke: #00c6e1;
}

.sparkline-container {
	position: relative;
}

.is-hidden {
	visibility: hidden;
}
</style>
