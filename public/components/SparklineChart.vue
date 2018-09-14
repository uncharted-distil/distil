<template>
	<div class="sparkline-container">
		<svg ref="svg" class="line-chart-big"></svg>
	</div>
</template>

<script lang="ts">

import * as d3 from 'd3';
import _ from 'lodash';
import Vue from 'vue';

export default Vue.extend({
	name: 'sparkline-chart',
	props: {
		margin: {
			type: Object as () => any,
			default: () => ({
				top: 16,
				right: 32,
				bottom: 16,
				left: 32
			})
		},
		timeseries: {
			type: Array as () => number[][]
		}
	},
	data() {
		return {
			xAxisTitle: 'X-Axis',
			yAxisTitle: 'Y-Axis',
			xScale: null,
			yScale: null
		};
	},
	computed: {
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
	mounted() {
		setTimeout(() => {
			// vue.js sucks
			this.injectTimeseries();
		});
	},
	methods: {
		clearSVG() {
			this.svg.selectAll('*').remove();
		},
		injectAxes() {
			const timeseries = this.timeseries;

			this.xScale = d3.scalePoint()
				.rangeRound([0, this.width]);

			this.xScale.domain(timeseries.map(d => d[0]));

			const min = d3.min(timeseries, d => d[1]);
			const max = d3.max(timeseries, d => d[1]);

			this.yScale = d3.scaleLinear()
				.domain([min, max])
				.range([this.height, 0]);

			// Create axes
			const xAxis = d3.axisBottom(this.xScale)
				.tickValues(this.timeseries.filter((d, i) => {
					return i % 10 === 0;
				}).map(d => d[0]));
			const yAxis = d3.axisLeft(this.yScale)
				.ticks(5);

			// Create x-axis
			const svgXAxis = this.svg.append('g')
				.attr('class', 'x axis')
				.attr('transform', `translate(${this.margin.left}, ${-this.margin.bottom + this.height})`)
				.call(xAxis);

			svgXAxis.append('text')
				 .attr('class', 'axis-title')
				.attr('x', this.width / 2)
				.attr('y', this.margin.bottom)
				.attr('dy', this.margin.bottom)
				.style('text-anchor', 'middle')
				.text(this.xAxisTitle);

			// Create y-axis
			const svgYAxis = this.svg.append('g')
				.attr('class', 'y axis')
				.attr('transform', `translate(${this.margin.left}, ${-this.margin.bottom})`)
				.call(yAxis);

			svgYAxis.append('text')
				.attr('class', 'axis-title')
				.attr('transform', 'rotate(-90)')
				.attr('x', -(this.height / 2))
				.attr('y', -this.margin.left / 2)
				.style('text-anchor', 'middle')
				.text(this.yAxisTitle);
		},
		injectSparkline() {
			const line = d3.line()
				.x(d => this.xScale(d[0]))
				.y(d => this.yScale(d[1]))
				.curve(d3.curveBasis);

			const g = this.svg.append('g')
				.attr('transform', `translate(${this.margin.left}, ${-this.margin.bottom})`)
				.attr('class', 'line-chart');

			g.datum(this.timeseries);

			g.append('path')
				.attr('fill', 'none')
				.attr('class', 'line')
				.attr('d', line);
		},
		injectTimeseries() {
			if (_.isEmpty(this.timeseries)) {
				console.log('no data');
				return;
			}

			if (this.width <= 0) {
				console.warn('Invalid width for line chart', this.width);
				return;
			}

			if (this.height <= 0) {
				console.warn('Invalid height for line chart', this.height);
				return;
			}

			this.clearSVG();
			this.injectAxes();
			this.injectSparkline();
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
	height: 512px;
	width: 100%;
	border: 1px solid rgba(0,0,0,0);
}

.line-chart  {
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
