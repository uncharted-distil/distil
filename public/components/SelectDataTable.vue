<template>
	<fixed-header-table ref="fixedHeaderTable">
		<b-table
			bordered
			hover
			small
			:items="items"
			:fields="fields"
			@sort-changed="onSortChanged"
			@row-clicked="onRowClick">

			<template v-for="computedField in computedFields" :slot="'HEAD_' + computedField" slot-scope="data">
				{{ data.label }} <icon-base :key="computedField" icon-name="fork" class="icon-fork" width=14 height=14> <icon-fork /></icon-base>
			</template>

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
import IconBase from './icons/IconBase';
import IconFork from './icons/IconFork';
import FixedHeaderTable from './FixedHeaderTable';
import SparklinePreview from './SparklinePreview';
import ImagePreview from './ImagePreview';
import { getters as datasetGetters } from '../store/dataset/module';
import { Dictionary } from '../util/dict';
import { Filter } from '../util/filters';
import { TableColumn, TableRow, Grouping, Variable, D3M_INDEX_FIELD, RowSelection } from '../store/dataset/index';
import { getters as routeGetters } from '../store/route/module';
import { IMAGE_TYPE, TIMESERIES_TYPE, hasComputedVarPrefix } from '../util/types';
import { addRowSelection, removeRowSelection, isRowSelected, updateTableRowSelection } from '../util/row';
import { getTimeseriesGroupingsFromFields } from '../util/data';

export default Vue.extend({
	name: 'selected-data-table',

	components: {
		ImagePreview,
		SparklinePreview,
		FixedHeaderTable,
		IconBase,
		IconFork,
	},

	props: {
		instanceName: String as () => string,
		includedActive: Boolean as () => boolean
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},

		items(): TableRow[] {
			const items = this.includedActive ? datasetGetters.getIncludedTableDataItems(this.$store) : datasetGetters.getExcludedTableDataItems(this.$store);
			return updateTableRowSelection(items, this.rowSelection, this.instanceName);
		},

		fields(): Dictionary<TableColumn> {
			return this.includedActive ? datasetGetters.getIncludedTableDataFields(this.$store) : datasetGetters.getExcludedTableDataFields(this.$store);
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
		},

		computedFields(): string[] {
			return Object.keys(this.fields).filter(key => {
				return hasComputedVarPrefix(key);
			});
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
	updated() {
		const fixedHeaderTable = this.$refs.fixedHeaderTable as any;
		fixedHeaderTable.resizeTableCells();
	},
	methods: {
		onSortChanged() {
			// need a `nextTick` otherwise the cells get immediately overwritten
			const currentScrollLeft = this.$el.querySelector('tbody').scrollLeft;
			Vue.nextTick(() => {
				const fixedHeaderTable = this.$refs.fixedHeaderTable as any;
				fixedHeaderTable.resizeTableCells();
				fixedHeaderTable.setScrollLeft(currentScrollLeft);
			});
		},
		onRowClick(row: TableRow) {
			if (!isRowSelected(this.rowSelection, row[D3M_INDEX_FIELD])) {
				addRowSelection(this.$router, this.instanceName, this.rowSelection, row[D3M_INDEX_FIELD]);
			} else {
				removeRowSelection(this.$router, this.instanceName, this.rowSelection, row[D3M_INDEX_FIELD]);
			}
		},
		invertFilters(filters: Filter[]): Filter[] {
			// TODO: invert filters
			return filters;
		},
	}
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

.table-selected-row {
	border-left: 4px solid #ff0067;
	background-color: rgba(255, 0, 103, 0.2);
}

.table-hover tbody .table-selected-row:hover {
	border-left: 4px solid #ff0067;
	background-color: rgba(255, 0, 103, 0.4);
}
</style>
