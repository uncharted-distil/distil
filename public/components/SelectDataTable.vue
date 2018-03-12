<template>
	<div class="select-data-table">
		<p>
			<b-nav tabs>
				<b-nav-item class="font-weight-bold" @click="includedActive=true" :active="includedActive">Samples to Model From</b-nav-item>
				<b-nav-item class="font-weight-bold" @click="includedActive=false" :active="!includedActive">Excluded Samples</b-nav-item>
			</b-nav>
		<p>

		<div>
			<div v-for="filter in filters">
				<filter-badge
					:filter="filter"></filter-badge>
			</div>
		</div>

		<p class="small-margin">
			<small>Displaying {{items.length}} of {{numRows}} rows</small>
			<b-button
				variant="outline-secondary"
				:disabled="!highlights.root"
				@click="onExcludeClick">
				<i class="fa fa-minus-circle pr-1"></i>Exclude
			</b-button>
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
				@row-clicked="onRowClick"
				:items="items"
				:fields="fields">
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
import { FieldInfo, Highlight } from '../store/data/index';
import { getters as routeGetters } from '../store/route/module';
import { TableRow } from '../store/data/index';
import { addFilterToRoute, EXCLUDE_FILTER } from '../util/filters';
import { updateTableHighlights, getHighlights, updateHighlightRoot, clearHighlightRoot, scrollToFirstHighlight, createFilterFromHighlightRoot } from '../util/highlights';

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

		selectedRowKey(): number {
			return routeGetters.getDecodedHighlightRoot(this.$store) ? _.toNumber(routeGetters.getDecodedHighlightRoot(this.$store).key) : -1;
		},

		hasData(): boolean {
			return dataGetters.hasSelectedData(this.$store);
		},

		// extracts the table data from the store
		items(): TableRow[] {
			const items = this.includedActive ? dataGetters.getSelectedDataItems(this.$store) : dataGetters.getExcludedDataItems(this.$store);

			// clear any existing selections
			items.forEach(f => f._rowVariant = null);

			// if we have highlights defined and the select table is not the source then updated
			// the highlight visuals.
			if ((_.get(this.highlights, 'root.context') !== this.instanceName)) {
				updateTableHighlights(items, this.highlights, this.instanceName);

				// On data / highlights change, scroll to first selected row
				scrollToFirstHighlight(this, 'selectTable', true);
			}

			if (this.selectedRowKey >= 0) {
				const toSelect = items.find(r => r._key === this.selectedRowKey);
				if (toSelect) {
					if (_.get(this.highlights, 'root.context') === this.instanceName) {
						toSelect._rowVariant = 'primary';
					} else {
						toSelect._rowVariant = null;
					}
				}
			}

			return items;
		},

		// extract the table field header from the store
		fields(): Dictionary<FieldInfo> {
			return this.includedActive ? dataGetters.getSelectedDataFields(this.$store) : dataGetters.getExcludedDataFields(this.$store);
		},

		filters(): Filter[] {
			return dataGetters.getFilters(this.$store);
		}
	},

	methods: {
		onExcludeClick() {
			const filter = createFilterFromHighlightRoot(this.highlights.root, EXCLUDE_FILTER);
			addFilterToRoute(this, filter);
			clearHighlightRoot(this);
		},
		onRowClick(row: TableRow) {
			if (row._key !== this.selectedRowKey) {
				// clicked on a different row than last time - new selection
				updateHighlightRoot(this, {
					context: this.instanceName,
					key: row._key.toString(),
					value: _.map(this.fields, (field, key) => [ key, row[key] ])
				});
			} else {
				// clicked on same row - reset the selection key and clear highlights
				clearHighlightRoot(this);
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
.missing-icon {
	padding-right: 4px;
}
table.b-table>tfoot>tr>th.sorting:before,
table.b-table>thead>tr>th.sorting:before,
table.b-table>tfoot>tr>th.sorting:after,
table.b-table>thead>tr>th.sorting:after {
	top: 0;
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
</style>
