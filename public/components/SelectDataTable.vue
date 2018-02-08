<template>
	<div class="select-data-table">
		<p class="nav-link font-weight-bold">Samples to Model From</p>
		<p class="small-margin"><small>Displaying {{items.length}} of {{numRows}} rows</small></p>
		<div class="select-data-table-container">
			<div class="select-data-no-results" v-if="!hasData">
				<div class="bounce1"></div>
				<div class="bounce2"></div>
				<div class="bounce3"></div>
			</div>
			<div class="select-data-no-results" v-if="hasData && items.length===0">
				No results
			</div>
			<b-table
				ref="selectTable"
				v-if="items.length>0"
				bordered
				hover
				small
				@row-clicked="onRowClick"
				:items="items"
				:fields="fields">

				<template :slot="`HEAD_${field.label}`" v-for="field in fields">
					{{field.label}}
					<type-change-menu
						:key="field.label"
						:field="field.label"></type-change-menu>
				</template>

			</b-table>
		</div>

	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import { getters as dataGetters } from '../store/data/module';
import { Dictionary } from '../util/dict';
import { Filter } from '../util/filters';
import { FieldInfo } from '../store/data/index';
import { getters as routeGetters } from '../store/route/module';
import { TableRow } from '../store/data/index';
import { getHighlights } from '../util/highlights';
import { updateTableHighlights, updateHighlightRoot, clearHighlightRoot, scrollToFirstHighlight } from '../util/highlights';
import TypeChangeMenu from '../components/TypeChangeMenu';

export default Vue.extend({
	name: 'selected-data-table',

	components: {
		TypeChangeMenu
	},

	props: {
		instanceName: { type: String, default: 'select-table-highlight' }
	},

	computed: {
		// get dataset from route
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
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
			const data = dataGetters.getSelectedDataItems(this.$store);
			const valueHighlights = getHighlights(this.$store);

			dataGetters.getSelectedDataItems(this.$store).forEach(f => f._rowVariant = null);

			// if we have highlights defined and the select table is not the source then updated
			// the highlight visuals.
			if ((_.get(valueHighlights, 'root.context') !== this.instanceName)) {
				updateTableHighlights(data, valueHighlights, this.instanceName);

				// On data / highlights change, scroll to first selected row
				scrollToFirstHighlight(this, 'selectTable', true);
			}

			if (this.selectedRowKey >= 0) {
				const toSelect = dataGetters.getSelectedDataItems(this.$store).find(r => r._key === this.selectedRowKey);
				if (toSelect) {
					if (_.get(valueHighlights, 'root.context') === this.instanceName) {
						toSelect._rowVariant = 'primary';
					} else {
						toSelect._rowVariant = null;
					}
				}

			}

			return data;
		},

		// extract the table field header from the store
		fields(): Dictionary<FieldInfo> {
			return dataGetters.getSelectedDataFields(this.$store);
		},

		filters(): Filter[] {
			return dataGetters.getSelectedFilters(this.$store);
		}
	},

	methods: {
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

</style>
