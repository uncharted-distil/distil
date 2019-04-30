<template>

	<div class="sparkline-variable">
		<div class="timeseries-var-col">{{label}}</div>
		<div class="timeseries-min-col">{{min.toFixed(2)}}</div>
		<div class="timeseries-max-col">{{max.toFixed(2)}}</div>
		<sparkline-svg class="sparkline-variable-chart"
			:highlight-pixel-x="highlightPixelX"
			:timeseries-extrema="timeseriesExtrema"
			:timeseries="timeseries"
			:forecast="forecast">
		</sparkline-svg>
	</div>

</template>

<script lang="ts">

import * as d3 from 'd3';
import $ from 'jquery';
import Vue from 'vue';
import SparklineSvg from './SparklineSvg';
import { TimeseriesExtrema } from '../store/dataset/index';

export default Vue.extend({
	name: 'sparkline-variable',

	components: {
		SparklineSvg
	},

	props: {
		label: String as () => string,
		timeseries: Array as () => number[][],
		forecast: Array as () => number[][],
		highlightPixelX: {
			type: Number as () => number
		},
		timeseriesExtrema: {
			type: Object as () => TimeseriesExtrema
		}
	},
	computed: {
		min(): number {
			const min = d3.min(this.timeseries, d => d[1]);
			return min !== undefined ? min : 0;
		},
		max(): number {
			const max = d3.max(this.timeseries, d => d[1]);
			return max !== undefined ? max : 0;
		}
	}

});
</script>

<style>

.sparkline-variable {
	position: relative;
	width: 100%;
	height: 32px;
	line-height: 32px;
	vertical-align: middle;
	border-bottom: 1px solid #999;
	padding: 0 8px;
}
.sparkline-variable-chart {
	float: left;
	position: relative;
	line-height: 32px;
	height: 32px;
	width: calc(100% - 276px);
}
</style>
