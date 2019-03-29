<template>

	<div class="sparkline-container" v-observe-visibility="visibilityChanged">
		<sparkline-svg
			:timeseries-extrema="timeseriesExtrema"
			:timeseries="timeseries">
		</sparkline-svg>
		<i class="fa fa-plus zoom-sparkline-icon" @click.stop="onClick"></i>
		<b-modal id="sparkline-zoom-modal" :title="timeseriesId"
			@hide="hideModal"
			:visible="zoomSparkline"
			hide-footer>
			<sparkline-chart :timeseries="timeseries" v-if="zoomSparkline"></sparkline-chart>
		</b-modal>
	</div>

</template>

<script lang="ts">

import * as d3 from 'd3';
import Vue from 'vue';
import SparklineChart from '../components/SparklineChart';
import SparklineSvg from '../components/SparklineSvg';
import { Dictionary } from '../util/dict';
import { TimeseriesExtrema } from '../store/dataset/index';
import { getters as datasetGetters, actions as datasetActions } from '../store/dataset/module';

export default Vue.extend({
	name: 'sparkline-preview',

	components: {
		SparklineSvg,
		SparklineChart
	},

	props: {
		dataset: String as () => string,
		xCol: String as () => string,
		yCol: String as () => string,
		timeseriesCol: String as () => string,
		timeseriesId: String as () => string
	},
	data() {
		return {
			zoomSparkline: false,
			isVisible: false,
			hasRequested: false,
		};
	},
	computed: {
		timeseriesForDataset(): Dictionary<number[][]> {
			const timeseries = datasetGetters.getTimeseries(this.$store);
			return timeseries[this.dataset];
		},
		isLoaded(): boolean {
			return !!this.timeseries;
		},
		timeseries(): number[][] {
			if (!this.timeseriesForDataset) {
				return null;
			}
			return this.timeseriesForDataset[this.timeseriesId];
		},
		timeseriesExtrema(): TimeseriesExtrema {
			if (!this.timeseries) {
				return null;
			}
			return {
				x: {
					min: d3.min(this.timeseries, d => d[0]),
					max: d3.max(this.timeseries, d => d[0])
				},
				y: {
					min: d3.min(this.timeseries, d => d[1]),
					max: d3.max(this.timeseries, d => d[1])
				}
			};
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
		onClick() {
			this.zoomSparkline = true;
		},
		hideModal() {
			this.zoomSparkline = false;
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

.zoom-sparkline-icon {
	position: absolute;
	right: 4px;
	top: 4px;
	color: #666;
	visibility: hidden;
}

.sparkline-container {
	position: relative;
	width: 100%;
}

.sparkline-container:hover .zoom-sparkline-icon {
	visibility: visible;
}

.zoom-sparkline-icon:hover {
	opacity: 0.7;
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
