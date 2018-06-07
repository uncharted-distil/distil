<template>
	<svg ref="svg" class="line-chart"></svg>
</template>

<script lang="ts">
import * as d3 from 'd3';
import _ from 'lodash';
import Vue from 'vue';
import { getters as timeseriesGetters, actions as timeseriesActions } from '../store/timeseries/module';

const RENDER_DEBOUNCE = 200;

export default Vue.extend({
	name: 'SparkLine',
	props: {
		imageUrl: String,
		margin: {
			type: Object,
			default: () => ({
				top: 0,
				right: 0,
				bottom: 0,
				left: 0
			})
		},
		smoothing: {
			type: String,
			default: 'basis'
		},
		zeroBased: {
			type: Boolean,
			default: false
		},
		className: {
			type: String,
			default: null
		},
		lastPointRadius: {
			type: Number,
			default: 0
		},
		xScaleType: {
			type: String,
			default: 'band'
		},
		timeSeriesUrl: {
			type: String
		}
	},
	watch: {
		timeseries(currTimeSeries) {
			if (currTimeSeries) {
				this.debouncedRender();
			}
		}
	},
	data() {
		const component = this as any;
		return {
			debouncedRender: _.debounce(component.render, RENDER_DEBOUNCE)
		};
	},
	computed: {
		timeseries(): any[] {
			const arg = timeseriesGetters.getTimeSeries(this.$store)[this.timeSeriesUrl];
			return arg ? arg.timeseries : null;
		},
	},
	mounted() {
		timeseriesActions.fetchTimeSeries(this.$store, { url: this.timeSeriesUrl });
	},
	methods: {
		render() {
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
				.attr('stroke', '#000')
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
		refreshLayout() {
			this.debouncedRender();
		}
	}
});
</script>

<style>

svg {
	position: relative;
	max-height: 32px;
	width: 100%;
}

</style>
