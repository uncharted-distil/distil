<template>

	<div class="select-timeseries-view" @mousemove="mouseMove">
		<div class="timeseries-row-header">
			<div class="timeseries-var-col pad-top"><b>VARIABLES</b></div>
			<div class="timeseries-min-col pad-top"><b>MIN</b></div>
			<div class="timeseries-max-col pad-top"><b>MAX</b></div>
			<div class="timeseries-chart-axis">
				<template v-if="!!timeseriesExtrema">
					x: {{timeseriesExtrema.x.min}}, {{timeseriesExtrema.x.max}}
					y: {{timeseriesExtrema.y.min}}, {{timeseriesExtrema.y.max}}
				</template>
			</div>
		</div>
		<div v-for="item in items">
			<sparkline-row :timeseries-url="item[timeseriesField]" :timeseries-extrema="timeseriesExtrema">
			</sparkline-row>
		</div>
		<div class="vertical-line"></div>
	</div>

</template>

<script lang="ts">

import _ from 'lodash';
import $ from 'jquery';
import Vue from 'vue';
import SparklineRow from './SparklineRow';
import { Dictionary } from '../util/dict';
import { Filter } from '../util/filters';
import { RowSelection } from '../store/highlights/index';
import { TableRow, TableColumn, TimeseriesExtrema } from '../store/dataset/index';
import { getters as routeGetters } from '../store/route/module';
import { getters as datasetGetters } from '../store/dataset/module';

export default Vue.extend({
	name: 'select-timeseries-view',

	components: {
		SparklineRow
	},

	props: {
		instanceName: String as () => string,
		includedActive: Boolean as () => boolean
	},

	data() {
		return {
		};
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		items(): TableRow[] {
			return this.includedActive ? datasetGetters.getIncludedTableDataItems(this.$store) : datasetGetters.getExcludedTableDataItems(this.$store);
		},

		fields(): Dictionary<TableColumn> {
			return this.includedActive ? datasetGetters.getIncludedTableDataFields(this.$store) : datasetGetters.getExcludedTableDataFields(this.$store);
		},

		timeseriesField(): string {
			const fields = _.map(this.fields, (field, key) => {
					return {
						key: key,
						type: field.type
					};
				})
				.filter(field => field.type === 'timeseries')
				.map(field => field.key);
			return fields[0];
		},

		filters(): Filter[] {
			if (this.includedActive) {
				return this.invertFilters(routeGetters.getDecodedFilters(this.$store));
			}
			return routeGetters.getDecodedFilters(this.$store);
		},

		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},

		timeseriesExtrema(): TimeseriesExtrema {
			const extrema = datasetGetters.getTimeseriesExtrema(this.$store);
			return extrema[this.dataset];
		}
	},

	methods: {
		invertFilters(filters: Filter[]): Filter[] {
			// TODO: invert filters
			return filters;
		},
		mouseMove(event) {
			const parentOffset = $('.select-timeseries-view').offset();
			const relX = event.pageX - parentOffset.left;

			const chartBounds = $('.timeseries-chart-axis').offset();
			const chartLeft = chartBounds.left - parentOffset.left;
			const chartWidth = $('.timeseries-chart-axis').width();
			if (relX >= chartLeft && relX <= chartLeft + chartWidth) {
				$('.vertical-line').show();
				$('.vertical-line').css('left', relX);
			} else {
				$('.vertical-line').hide();
			}
		}
	},

	mounted() {
	}

});
</script>

<style>

.select-timeseries-view {
	position: relative;
	flex: 1;
}
.timeseries-row-header {
	height: 64px;
	line-height: 32px;
	border-bottom: 1px solid #999;
	padding: 0 8px;
}
.timeseries-var-col {
	float: left;
	position: relative;
	line-height: 32px;
	height: 32px;
	width: 156px;
}
.timeseries-min-col {
	float: left;
	position: relative;
	line-height: 32px;
	height: 32px;
	width: 48px;
}
.timeseries-max-col {
	float: left;
	position: relative;
	line-height: 32px;
	height: 32px;
	width: 48px;
}
.timeseries-chart-col {
	float: left;
	position: relative;
	line-height: 32px;
	height: 32px;
	width: calc(100% - 276px);
}
.timeseries-chart-axis {
	float: left;
	position: relative;
	line-height: 32px;
	height: 64px;
	border: 1px solid red;
	width: calc(100% - 276px);
}
.pad-top {
	padding-top: 32px;
}
.vertical-line {
	position: absolute;
	display: none;
	top: 0;
	left: 0;
	width: 1px;
	height: 100%;
	border-left: 1px solid #00c6e1;
	box-shadow: 0px 0px 5px #00c6e1;
}
</style>
