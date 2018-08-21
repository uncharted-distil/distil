<template>
	<div class="sparkline-container">
		<svg v-if="isLoaded" ref="svg" class="line-chart" @click.stop="onClick"></svg>
		<i class="fa fa-plus zoom-icon"></i>
		<div v-if="isErrored">Error</div>
		<div v-if="!isErrored && !isLoaded" v-html="spinnerHTML"></div>
		<b-modal id="sparkline-zoom-modal" :title="timeSeriesUrl"
			@hide="hideModal"
			:visible="!!zoomSparkline"
			hide-footer>
			<div class="sparkline-elem-zoom" ref="sparklineElemZoom"></div>
		</b-modal>
	</div>
</template>

<script lang="ts">

import * as d3 from 'd3';
import _ from 'lodash';
import axios from 'axios';
import Vue from 'vue';
import { circleSpinnerHTML } from '../util/spinner';

export default Vue.extend({
	name: 'sparkline-preview',
	props: {
		margin: {
			type: Object as () => any,
			default: () => ({
				top: 8,
				right: 4,
				bottom: 8,
				left: 4
			})
		},
		smoothing: {
			type: String as () => string,
			default: 'basis'
		},
		zeroBased: {
			type: Boolean as () => boolean,
			default: false
		},
		className: {
			type: String as () => string,
			default: null
		},
		lastPointRadius: {
			type: Number as () => number,
			default: 0
		},
		xScaleType: {
			type: String as () => string,
			default: 'band'
		},
		timeSeriesUrl: {
			type: String as () => string
		}
	},
	data() {
		return {
			zoomSparkline: false,
			entry: null
		};
	},
	computed: {
		isLoaded(): boolean {
			return this.entry && this.entry.timeseries;
		},
		isErrored(): boolean {
			return this.entry && this.entry.err;
		},
		timeseries(): any[] {
			return this.entry ? this.entry.timeseries : null;
		},
		spinnerHTML(): string {
			return circleSpinnerHTML();
		}
	},
	mounted() {
		this.requestTimeseries(this.timeSeriesUrl);
	},
	updated() {
		if (_.isEmpty(this.timeseries)) {
			return;
		}

		const timeseries = this.timeseries;

		const $svg = this.$refs.svg as any;
		const svg = d3.select($svg);
		svg.selectAll('*').remove();

		const hasLastPoint = (this.lastPointRadius > 0 && timeseries.length > 0);
		const dims = $svg.getBoundingClientRect();

		let width = dims.width - this.margin.left - this.margin.right;
		let height = dims.height - this.margin.top - this.margin.bottom;

		height = hasLastPoint ? height - this.lastPointRadius : height;
		width = hasLastPoint ? width - this.lastPointRadius : width;

		if (width <= 0) {
			console.warn('Invalid width for line chart');
			return;
		}

		if (height <= 0) {
			console.warn('Invalid height for line chart');
			return;
		}

		let xScale;
		if (this.xScaleType === 'point') {
			xScale = d3.scalePoint().range([0, width]);
		} else {
			xScale = d3.scaleBand().rangeRound([0, width], 0);
		}
		xScale.domain(timeseries.map(d => d.timestamp));

		const min = this.zeroBased ? 0 : d3.min(this.timeseries, d => d.count);
		const max = d3.max(timeseries, d => d.count);

		const yScale = d3.scaleLinear()
			.domain([min, max])
			.range([height, 0]);

		let curveType = d3.curveBasis;
		if (this.smoothing === 'linear') {
			curveType = d3.curveLinear;
		}

		const line = d3.line()
			.x(d => xScale(d.timestamp))
			.y(d => yScale(d.count))
			.curve(curveType);

		const className = this.className || 'line-chart';
		const g = svg.append('g')
			.attr('transform', `translate(${this.margin.left}, ${this.margin.top})`)
			.attr('class', className);

		g.datum(timeseries);

		g.append('path')
			.attr('fill', 'none')
			.attr('class', 'line')
			.attr('d', line);

		if (hasLastPoint) {
			const lastPoint = timeseries[timeseries.length - 1];
			g.append('circle')
				.attr('cx', xScale(lastPoint.timestamp))
				.attr('cy', yScale(lastPoint.count))
				.attr('r', this.lastPointRadius)
				.attr('class', 'last-point');
		}
	},
	methods: {
		onClick() {
			const $svg = this.$refs.svg as any;
			const $elem = this.$refs.sparklineElemZoom as any;
			$elem.innerHTML = '';
			$elem.appendChild($svg.cloneNode(true));
			this.zoomSparkline = true;
		},
		hideModal() {
			this.zoomSparkline = false;
		},
		requestTimeseries(url: string) {
			// DEBUG:
			const TIME_SERIES = [
				'a.csv',
				'b.csv',
				'c.csv',
				'd.csv',
				'e.csv'
			];
			return axios.get(`timeseries/${TIME_SERIES[Math.floor(Math.random() * TIME_SERIES.length)]}`)
				.then(response => {
					const lines = response.data.split('\n');
					const timeseries = lines.slice(1, lines.length - 1).map(entry => {
						const split = entry.split(',');
						return {
							timestamp: split[0],
							count: _.toNumber(split[1])
						};
					});
					this.entry = { url: url, timeseries: timeseries };
				})
				.catch(err => {
					console.error(err);
					this.entry = { url: url, err: err };
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

.zoom-icon {
	position: absolute;
	right: 4px;
	top: 4px;
	color: #666;
	visibility: hidden;
}

.sparkline-container {
	position: relative;
}

.sparkline-container:hover .zoom-icon {
	visibility: visible;
}

.zoom-icon {
	pointer-events: none;
}

.sparkline-elem-zoom {
	position: relative;
	padding: 32px 16px;
	border-radius: 4px;
}

#sparkline-zoom-modal .modal-dialog {
	max-width: 50%;
}

</style>
