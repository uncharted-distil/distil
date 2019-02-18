<template>
	<div class="table-holder h-100">
		<object class="resize-observer"></object>
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
		resizeTableCells() {
			const theadCells = this.$el.querySelectorAll('thead tr')[0]
				.querySelectorAll('th');
			const firstRow = this.$el.querySelectorAll('tbody tr')[0];
			const tbodyCells = firstRow.querySelectorAll('td');
			const headTargetCells = [];

			// reset element style so that table renders with initial layout set by css
			for (let i = 0; i < theadCells.length; i++) {
				tbodyCells[i].removeAttribute('style');
				theadCells[i].removeAttribute('style');
			}
			// get new adjusted header cell width based on the corresponding data cell width
			for (let i = 0; i < theadCells.length; i++) {
				const headCellWidth = theadCells[i].offsetWidth;
				const bodyCellWidth = tbodyCells[i].offsetWidth;
				if (headCellWidth < bodyCellWidth) {
					headTargetCells.push({ elem: theadCells[i], width: bodyCellWidth });
				}
			}
			const setCellWidth = cell => {
				cell.elem.style['max-width'] = cell.width + 'px';
				cell.elem.style['min-width'] = cell.width + 'px';
			};
			headTargetCells.forEach(setCellWidth);

			// get body cell width from computed table header cell width
			const bodyCells = [];
			for (let i = 0; i < theadCells.length; i++) {
				const headCellWidth = theadCells[i].offsetWidth;
				bodyCells.push({ elem: tbodyCells[i], width: headCellWidth });
			}
			// set new body cell width
			bodyCells.forEach(setCellWidth);
		}
	},
	mounted: function () {
		const resizeObserver = this.$el.querySelector('object');

		window.addEventListener('resize', () => {
			this.resizeTableCells();
		});
		this.resizeTableCells();
		console.log('mounted: ');
	},
	// beforeDestroy: function () {
	// 	// this.$el.querySelector('object').removeEventListener('resize', () => {});
	// },
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
	position: relative;
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
