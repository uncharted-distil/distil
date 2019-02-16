<template>
	<div class="table-holder h-100" @scroll="handleScroll">
	<!-- <div class="h-100" @scroll="handleScroll"> -->
		<b-table
			bordered
			hover
			small
			:items="items"
			:fields="emphasizedFields"
			@head-clicked="onColumnClicked">

			<template v-for="imageField in imageFields" :slot="imageField" slot-scope="data">
				<image-preview :key="imageField" :image-url="data.item[imageField]"></image-preview>
			</template>

			<template v-for="timeseriesField in timeseriesFields" :slot="timeseriesField" slot-scope="data">
				<sparkline-preview :key="timeseriesField" :timeseries-url="data.item[timeseriesField]"></sparkline-preview>
			</template>

		</b-table>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import SparklinePreview from './SparklinePreview';
import ImagePreview from './ImagePreview';
import { Dictionary } from '../util/dict';
import { TableColumn, TableRow, D3M_INDEX_FIELD } from '../store/dataset/index';
import { getters as routeGetters } from '../store/route/module';
import { IMAGE_TYPE, TIMESERIES_TYPE, isJoinable } from '../util/types';

export default Vue.extend({
	name: 'join-data-table',

	components: {
		ImagePreview,
		SparklinePreview
	},

	props: {
		items: Array as () => TableRow[],
		fields: Object as () => Dictionary<TableColumn>,
		selectedColumn: Object as () => TableColumn,
		otherSelectedColumn: Object as () => TableColumn,
		instanceName: String as () => string
	},

	computed: {

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

		timeseriesFields(): string[] {
			return _.map(this.fields, (field, key) => {
				return {
					key: key,
					type: field.type
				};
			})
			.filter(field => field.type === TIMESERIES_TYPE)
			.map(field => field.key);
		}
	},

	methods: {
		onColumnClicked(key, field) {
			if (this.selectedColumn && this.selectedColumn.key === key) {
				this.$emit('col-clicked', null);
			} else {
				this.$emit('col-clicked', field);
			}
		},
		resizeTableColumns() {
			const theadCells = this.$el.querySelectorAll('thead tr')[0]
				.querySelectorAll('th');
			const firstRow = this.$el.querySelectorAll('tbody tr')[0];
			const tbodyCells = firstRow.querySelectorAll('td');
			for (let i = 0; i < theadCells.length; i++) {
				const headCellWidth = theadCells[i].offsetWidth;
				const bodyCellWidth = tbodyCells[i].offsetWidth;
				const targetCell = headCellWidth > bodyCellWidth
					? tbodyCells[i]
					: theadCells[i];
				targetCell.style.width = Math.max(headCellWidth, bodyCellWidth) + 'px';
				targetCell.style['min-width'] = targetCell.style.width;
			}
		}
	},
	mounted: function () {
		this.resizeTableColumns();
		console.log('mounted: ');
	},
	beforeUpdate: () => {
		console.log('before Update');
	},
	updated: () => {
		console.log('updated');
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
.table-holder {
	overflow-x: auto;
	height: 100%;
	width: 100%;
}
.table-holder table {
	table-layout: fixed;
	height: 100%;
	margin: 0;

	display: flex;
	flex-direction: column;
	align-items: flex-start;
}
.table-holder thead {
	width: 100%
}
.table-holder thead tr {
	display: flex;
}
.table-holder thead th {
	flex-shrink: 0;
	flex-grow: 1;
}
.table-holder tbody {
	overflow-y: auto;
	flex: 1;
}
.table-holder tbody td {
	overflow-wrap: break-word;
}

</style>
