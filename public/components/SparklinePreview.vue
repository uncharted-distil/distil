<template>
	<div class="sparkline-container">
		<svg v-if="isLoaded" ref="svg" class="line-chart" @click.stop="onClick"></svg>
		<i class="fa fa-plus zoom-icon"></i>
		<div v-if="!isLoaded" v-html="spinnerHTML"></div>
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
import Vue from 'vue';
import { Dictionary } from '../util/dict';
import { circleSpinnerHTML } from '../util/spinner';
import { getters as routeGetters } from '../store/route/module';
import { getters as datasetGetters, actions as datasetActions } from '../store/dataset/module';

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
		}
	},
	mounted() {
		this.requestTimeseries();
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
		injectTimeseries() {
			if (_.isEmpty(this.timeseries)) {
				return;
			}

			const $svg = this.$refs.svg as any;
			const svg = d3.select($svg);
			svg.selectAll('*').remove();

			const timeseries = this.timeseries;
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
			xScale.domain(timeseries.map(d => d[0]));

			const min = d3.min(this.timeseries, d => d[1]);
			const max = d3.max(timeseries, d => d[1]);

			const yScale = d3.scaleLinear()
				.domain([min, max])
				.range([height, 0]);

			let curveType = d3.curveBasis;
			if (this.smoothing === 'linear') {
				curveType = d3.curveLinear;
			}

			const line = d3.line()
				.x(d => xScale(d[0]))
				.y(d => yScale(d[1]))
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
					.attr('cx', xScale(lastPoint[0]))
					.attr('cy', yScale(lastPoint[1]))
					.attr('r', this.lastPointRadius)
					.attr('class', 'last-point');
			}
		},
		requestTimeseries() {
			datasetActions.fetchTimeseries(this.$store, {
				dataset: this.dataset,
				url: this.timeSeriesUrl
			}).then(() => {
				this.injectTimeseries();
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
