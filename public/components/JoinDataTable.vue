<template>

	<b-table
		bordered
		hover
		small
		responsive
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

</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import SparklinePreview from './SparklinePreview';
import ImagePreview from './ImagePreview';
import { Dictionary } from '../util/dict';
import { Filter } from '../util/filters';
import { TableColumn, TableRow, D3M_INDEX_FIELD } from '../store/dataset/index';
import { RowSelection } from '../store/highlights/index';
import { getters as routeGetters } from '../store/route/module';
import { IMAGE_TYPE, TIMESERIES_TYPE } from '../util/types';
import { addRowSelection, removeRowSelection, isRowSelected } from '../util/row';

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
				if (this.selectedColumn && field.key === this.selectedColumn.key) {
					emph.variant = 'primary';
				} else if (this.otherSelectedColumn && field.type === this.otherSelectedColumn.type) {
					// show matching column types
					emph.variant = 'success';
				} else if (this.otherSelectedColumn && field.type !== this.otherSelectedColumn.type) {
					// show unmatched column types
					emph.variant = 'warning';
				}

				if (this.otherSelectedColumn && this.selectedColumn &&
					field.key === this.selectedColumn.key &&
					this.selectedColumn.type !== this.otherSelectedColumn.type) {
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
		},

		filters(): Filter[] {
			return this.invertFilters(routeGetters.getDecodedFilters(this.$store));
		},

		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		}
	},

	methods: {
		invertFilters(filters: Filter[]): Filter[] {
			// TODO: invert filters
			return filters;
		},
		onColumnClicked(key, field) {
			if (this.selectedColumn && this.selectedColumn.key === key) {
				this.$emit('col-clicked', null);
			} else {
				this.$emit('col-clicked', field);
			}
		}
	}
});
</script>

<style>

.select-data-table {
	display: flex;
	flex-direction: column;
}
.select-data-table-container {
	display: flex;
	overflow: auto;
	background-color: white;
}
.select-data-no-results {
	width: 100%;
	background-color: #eee;
	padding: 8px;
	text-align: center;
}
table.b-table>tfoot>tr>th.sorting:before,
table.b-table>thead>tr>th.sorting:before,
table.b-table>tfoot>tr>th.sorting:after,
table.b-table>thead>tr>th.sorting:after {
	top: 0;
}

table tr {
	cursor: pointer;
}
.select-data-table .small-margin {
	margin-bottom: 0.5rem
}
.select-view .nav-tabs .nav-item a {
	padding-left: 0.5rem;
	padding-right: 0.5rem;
}
.select-view .nav-tabs .nav-link {
	color: #757575;
}
.select-view .nav-tabs .nav-link.active {
	color: rgba(0, 0, 0, 0.87);
}
.include-highlight,
.exclude-highlight {
	color: #00c6e1;
}
.include-selection,
.exclude-selection {
	color: #ff0067;
}
.row-number-label {
	position: relative;
	top: 20px;
}
.matching-color {
	color: #00c6e1;
}
.fake-search-input {
	position: relative;
	height: 38px;
	padding: 2px 2px;
	margin-bottom: 4px;
	background-color: #eee;
	border: 1px solid #ccc;
	border-radius: 0.2rem;
}
.selected-color {
	color: #ff0067;
}
.table-selected-row {
	border-left: 4px solid #ff0067;
	background-color: rgba(255, 0, 103, 0.2);
}
</style>
