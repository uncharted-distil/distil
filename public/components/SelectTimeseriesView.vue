<template>

	<div class="select-timeseries-view">
		<div v-for="item in items">
			<sparkline-row :timeSeries-url="item[timeseriesField]">
			</sparkline-row>
		</div>

	</div>

</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import { Dictionary } from '../util/dict';
import { Filter } from '../util/filters';
import { RowSelection } from '../store/highlights/index';
import { TableRow, TableColumn } from '../store/dataset/index';
import { getters as routeGetters } from '../store/route/module';
import { getters as datasetGetters } from '../store/dataset/module';

export default Vue.extend({
	name: 'select-timeseries-view',

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

.select-graph-view {
	flex: 1;
}

#graph-container {
	position: relative;
	height: 100%;
	width: 100%;
}

</style>
