<template>
	<fixed-header-table ref="fixedHeaderTable">
		<b-table
			bordered
			hover
			small
			:items="items"
			:fields="emphasizedFields"
			@sort-changed="onSortChanged"
			@head-clicked="onColumnClicked">

			<template v-for="imageField in imageFields" :slot="imageField" slot-scope="data">
				<image-preview :key="imageField" :image-url="data.item[imageField]"></image-preview>
			</template>

			<template v-for="timeseriesGrouping in timeseriesGroupings" :slot="timeseriesGrouping.idCol" slot-scope="data">

				<sparkline-preview :key="timeseriesGrouping.idCol"
					:dataset="dataset"
					:x-col="timeseriesGrouping.properties.xCol"
					:y-col="timeseriesGrouping.properties.yCol"
					:timeseries-col="timeseriesGrouping.idCol"
					:timeseries-id="data.item[timeseriesGrouping.idCol]">
				</sparkline-preview>

			</template>

		</b-table>
	</fixed-header-table>
</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import FixedHeaderTable from './FixedHeaderTable';
import SparklinePreview from './SparklinePreview';
import ImagePreview from './ImagePreview';
import { Dictionary } from '../util/dict';
import { TableColumn, TableRow, D3M_INDEX_FIELD, Grouping, Variable } from '../store/dataset/index';
import { getters as routeGetters } from '../store/route/module';
import { getters as datasetGetters } from '../store/dataset/module';
import { IMAGE_TYPE, TIMESERIES_TYPE, isJoinable } from '../util/types';
import { getTimeseriesGroupingsFromFields } from '../util/data';

export default Vue.extend({
	name: 'join-data-table',

	components: {
		ImagePreview,
		SparklinePreview,
		FixedHeaderTable,
	},

	props: {
		dataset: String as () => string,
		items: Array as () => TableRow[],
		fields: Object as () => Dictionary<TableColumn>,
		selectedColumn: Object as () => TableColumn,
		otherSelectedColumn: Object as () => TableColumn,
		instanceName: String as () => string
	},

	computed: {

		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},

		emphasizedFields(): Dictionary<TableColumn> {
			if (!this.selectedColumn && !this.otherSelectedColumn) {
				return this.fields;
			}
			const emphasized = {};
			_.forIn(this.fields, field => {
				const emph = {
					label: field.label,
					key: field.key,
					type: field.type,
					sortable: field.sortable,
					variant: null
				};

				const isFieldSelected = this.selectedColumn && field.key === this.selectedColumn.key;
				const isFieldJoinable = this.otherSelectedColumn && isJoinable(field.type, this.otherSelectedColumn.type);

				if (isFieldSelected) {
					emph.variant = 'primary';
				} else if (isFieldJoinable) {
					// show matching column types
					emph.variant = 'success';
				} else if (!isFieldJoinable) {
					// show unmatched column types
					emph.variant = 'warning';
				}

				if (this.otherSelectedColumn && isFieldSelected && !isFieldJoinable) {
					// flag bad selection
					emph.variant = 'danger';
				}
				emphasized[field.key] = emph;
			});
			return emphasized;
		},

		imageFields(): string[] {
			return _.map(this.fields, (field, key) => {
				return {
					key: key,
					type: field.type
				};
			})
			.filter(field => field.type === IMAGE_TYPE)
			.map(field => field.key);
		},

		timeseriesGroupings(): Grouping[] {
			return getTimeseriesGroupingsFromFields(this.variables, this.fields);
		}
	},

	methods: {
		onSortChanged() {
			// need a `nextTick` otherwise the cells get immediately overwritten
			Vue.nextTick(() => {
				const fixedHeaderTable = this.$refs.fixedHeaderTable as any;
				fixedHeaderTable.resizeTableCells();
			});
		},
		onColumnClicked(key, field) {
			if (this.selectedColumn && this.selectedColumn.key === key) {
				this.$emit('col-clicked', null);
			} else {
				this.$emit('col-clicked', field);
			}
		},
	},
});
</script>

<style>

table.b-table>tfoot>tr>th.sorting:before,
table.b-table>thead>tr>th.sorting:before,
table.b-table>tfoot>tr>th.sorting:after,
table.b-table>thead>tr>th.sorting:after {
	top: 0;
}

table tr {
	cursor: pointer;
}

</style>
