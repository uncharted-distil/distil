<template>

	<div class="sparkline-row" v-observe-visibility="visibilityChanged">
		<div class="timeseries-var-col">{{timeseriesId}}</div>
		<div class="timeseries-min-col">{{min.toFixed(2)}}</div>
		<div class="timeseries-max-col">{{max.toFixed(2)}}</div>
		<sparkline-svg class="sparkline-row-chart"
			:highlight-pixel-x="highlightPixelX"
			:timeseries-extrema="timeseriesExtrema"
			:timeseries="timeseries">
		</sparkline-svg>
	</div>

</template>

<script lang="ts">

import * as d3 from 'd3';
import $ from 'jquery';
import Vue from 'vue';
import SparklineSvg from './SparklineSvg';
import { getters as routeGetters } from '../store/route/module';
import { TimeseriesExtrema } from '../store/dataset/index';
import { getters as datasetGetters, actions as datasetActions } from '../store/dataset/module';

export default Vue.extend({
	name: 'sparkline-row',

	components: {
		SparklineSvg
	},

	props: {
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
			isVisible: false,
			hasRequested: false
		};
	},
	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		timeseries(): number[][] {
			return datasetGetters.getTimeseries(this.$store)[this.dataset][this.timeseriesId];
		},
		min(): number {
			return this.timeseries ? d3.min(this.timeseries, d => d[1]) : 0;
		},
		max(): number {
			return this.timeseries ? d3.max(this.timeseries, d => d[1]) : 0;
		}
	},

	methods: {
		visibilityChanged(isVisible: boolean) {
			this.isVisible = isVisible;
			if (this.isVisible && !this.hasRequested) {
				this.requestTimeseries();
				return;
			}
		},
		requestTimeseries() {
			this.hasRequested = true;
			datasetActions.fetchTimeseries(this.$store, {
				dataset: this.dataset,
				xColName: this.xCol,
				yColName: this.yCol,
				timeseriesColName: this.timeseriesCol,
				timeseriesID: this.timeseriesId
			});
		}
	}

});
</script>

<style>

.sparkline-row {
	position: relative;
	width: 100%;
	height: 32px;
	line-height: 32px;
	vertical-align: middle;
	border-bottom: 1px solid #999;
	padding: 0 8px;
}
.sparkline-row-chart {
	float: left;
	position: relative;
	line-height: 32px;
	height: 32px;
	width: calc(100% - 276px);
}
</style>
