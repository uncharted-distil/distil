<template>

	<div class="select-timeseries-view">
		<div class="timeseries-row-header">
			<div class="timeseries-var-col"><b>VARIABLES</b></div>
			<div class="timeseries-min-col"><b>MIN</b></div>
			<div class="timeseries-max-col"><b>MAX</b></div>
			<div class="timeseries-chart-col"></div>
		</div>
		<div v-for="item in items">
			<sparkline-row :timeseries-url="item[timeseriesField]">
			</sparkline-row>
		</div>

	</div>

</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import SparklineRow from './SparklineRow';
import { Dictionary } from '../util/dict';
import { Filter } from '../util/filters';
import { RowSelection } from '../store/highlights/index';
import { TableRow, TableColumn } from '../store/dataset/index';
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
		}
	},

	methods: {
		invertFilters(filters: Filter[]): Filter[] {
			// TODO: invert filters
			return filters;
		}
	},

	mounted() {
	}

});
</script>

<style>

.select-timeseries-view {
	flex: 1;
}
.timeseries-row-header {
	height: 32px;
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
</style>
