<template>
	<div class="select-data-table">
		<p>
			<b-nav tabs>
				<b-nav-item class="font-weight-bold" @click="includedActive=true" :active="includedActive">Samples to Model From</b-nav-item>
				<b-nav-item class="font-weight-bold" @click="includedActive=false" :active="!includedActive">Excluded Samples</b-nav-item>
			</b-nav>
		</p>

		<div>
			<filter-badge v-if="activeFilter"
				active-filter
				:filter="activeFilter"></filter-badge>
			<div v-for="filter in filters">
				<filter-badge
					:filter="filter"></filter-badge>
			</div>
		</div>

		<p class="small-margin">
			<b-button class="float-right" v-if="includedActive"
				variant="outline-secondary"
				:disabled="!highlights.root"
				@click="onExcludeClick">
				<i class="fa fa-minus-circle pr-1"></i>Exclude
			</b-button>
			<b-button class="float-right" v-if="!includedActive"
				variant="outline-secondary"
				:disabled="!highlights.root"
				@click="onReincludeClick">
				<i class="fa fa-plus-circle pr-1"></i>Reinclude
			</b-button>
			<small class="row-number-label">Displaying {{items.length}} of {{numRows}} rows</small>
		</p>

		<div class="select-data-table-container">
			<div class="select-data-no-results" v-if="!hasData">
				<div class="bounce1"></div>
				<div class="bounce2"></div>
				<div class="bounce3"></div>
			</div>
			<div class="select-data-no-results" v-if="hasData && items.length===0">
				No data available
			</div>
			<b-table v-if="items.length>0"
				ref="selectTable"
				bordered
				hover
				small
				responsive
				:items="items"
				:fields="fields"
				@row-clicked="onRowClick">
			</b-table>
		</div>

	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import FilterBadge from './FilterBadge';
import { getters as dataGetters } from '../store/data/module';
import { Dictionary } from '../util/dict';
import { Filter } from '../util/filters';
import { FieldInfo, Highlight, RowSelection } from '../store/data/index';
import { getters as routeGetters } from '../store/route/module';
import { TableRow } from '../store/data/index';
import { addFilterToRoute, EXCLUDE_FILTER, INCLUDE_FILTER } from '../util/filters';
import { getHighlights, clearHighlightRoot, createFilterFromHighlightRoot } from '../util/highlights';
import { updateRowSelection, clearRowSelection, updateTableRowSelection } from '../util/row';

export default Vue.extend({
	name: 'selected-data-table',

	components: {
		FilterBadge
	},

	props: {
		instanceName: { type: String, default: 'select-table-highlight' }
	},

	data() {
		return {
			includedActive: true
		};
	},

	computed: {
		// get dataset from route
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		highlights(): Highlight {
			return getHighlights(this.$store);
		},

		numRows(): number {
			return dataGetters.getSelectedDataNumRows(this.$store);
		},

		hasData(): boolean {
			return dataGetters.hasSelectedData(this.$store);
		},

		// extracts the table data from the store
		items(): TableRow[] {
			const items = this.includedActive ? dataGetters.getSelectedDataItems(this.$store) : dataGetters.getExcludedDataItems(this.$store);
			return updateTableRowSelection(items, this.selectedRow, this.instanceName);
		},

		// extract the table field header from the store
		fields(): Dictionary<FieldInfo> {
			return this.includedActive ? dataGetters.getSelectedDataFields(this.$store) : dataGetters.getExcludedDataFields(this.$store);
		},

		activeFilter(): Filter {
			if (!this.highlights ||
				!this.highlights.root ||
				!this.highlights.root.value) {
				return null;
			}
			if (this.includedActive) {
				return createFilterFromHighlightRoot(this.highlights.root, INCLUDE_FILTER);
			}
			return createFilterFromHighlightRoot(this.highlights.root, EXCLUDE_FILTER);
		},

		filters(): Filter[] {
			if (this.includedActive) {
				return this.invertFilters(dataGetters.getFilters(this.$store));
			}
			return dataGetters.getFilters(this.$store);
		},

		selectedRow(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},

		selectedRowIndex(): number {
			return this.selectedRow ? this.selectedRow.index : -1;
		}
	},

	methods: {
		onExcludeClick() {
			const filter = createFilterFromHighlightRoot(this.highlights.root, EXCLUDE_FILTER);
			addFilterToRoute(this, filter);
			clearHighlightRoot(this);
		},
		onReincludeClick() {
			const filter = createFilterFromHighlightRoot(this.highlights.root, INCLUDE_FILTER);
			addFilterToRoute(this, filter);
			clearHighlightRoot(this);
		},
		onRowClick(row: TableRow) {
			if (row._key !== this.selectedRowIndex) {
				// clicked on a different row than last time - new selection
				updateRowSelection(this, {
					context: this.instanceName,
					index: row._key,
					cols: _.map(this.fields, (field, key) => {
						return {
							key: key,
							value: row[key]
						};
					})
				});
			} else {
				// clicked on same row - reset the selection key and clear highlights
				clearRowSelection(this);
			}
		},
		invertFilters(filters: Filter[]): Filter[] {
			// TODO: invert filters
			return filters;
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
.missing-icon {
	padding-right: 4px;
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
.row-number-label {
	position: relative;
	top: 20px;
}
</style>
