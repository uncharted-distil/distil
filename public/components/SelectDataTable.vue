<template>
	<div class="select-data-table">
		<p class="nav-link font-weight-bold">Training Set Samples</p>
		<p><small>Displaying {{items.length}} of {{numRows}} rows</small></p>
		<div class="select-data-table-container">
			<div class="select-data-no-results" v-if="items.length===0">
				<div class="text-danger">
					<i class="fa fa-times missing-icon"></i><strong>No Training features Selected</strong>
				</div>
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
import { getters as dataGetters, actions, mutations } from '../store/data/module';
import { Dictionary } from '../util/dict';
import { Filter } from '../util/filters';
import { FieldInfo } from '../store/data/index';
import { getters as routeGetters } from '../store/route/module';
import { Highlights, TableRow } from '../store/data/index';
import { updateTableHighlights, scrollToFirstHighlight } from '../util/highlights';
import TypeChangeMenu from '../components/TypeChangeMenu';

export default Vue.extend({
	name: 'selected-data-table',

	components: {
		TypeChangeMenu
	},

	props: {
		instanceName: { type: String, default: 'select-table-highlight' }
	},

	data() {
		return {
			selectedRowKey: -1
		};
	},

	computed: {
		// get dataset from route
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		numRows(): number {
			return dataGetters.getSelectedDataNumRows(this.$store);
		},

		// extracts the table data from the store
		items(): TableRow[] {
			const data = dataGetters.getSelectedDataItems(this.$store);
			const valueHighlights = dataGetters.getHighlightedFeatureValues(this.$store);

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
				if (_.get(valueHighlights, 'root.context') === this.instanceName) {
					toSelect._rowVariant = 'primary';
				} else {
					toSelect._rowVariant = null;
					this.selectedRowKey = -1;
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

	mounted() {
		this.fetch();
	},

	watch: {
		filters() {
			this.fetch();
		}
	},

	methods: {
		fetch() {
			actions.updateSelectedData(this.$store, {
				dataset: this.dataset,
				filters: this.filters
			});
		},

		onRowClick(row: TableRow) {
			if (row._key !== this.selectedRowKey) {
				this.selectedRowKey = row._key;

				// publish the highlight change
				const highlights = <Highlights> {
					root: {
						context: this.instanceName,
						key: row._key.toString(),
						value: ''
					},
					values: <Dictionary<string[]>>{}
				};
				_.forEach(this.fields, (field, key) => highlights.values[key] = [row[key]]);
				mutations.highlightFeatureValues(this.$store, highlights);
			} else {
				// clicked on same row - reset the selection key and clear highlights
				mutations.clearFeatureHighlights(this.$store);
				this.selectedRowKey = -1;
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
}
.missing-icon {
	padding-right: 4px;
}
.var-type-button {
	width: 100%;
}
.var-type-button button {
	border: none;
	padding: 0;
	width: 100%;
	text-align: left;
	outline: none;
	font-size: 0.9rem;
}
.var-type-button button:hover,
.var-type-button button:active,
.var-type-button button:focus,
.var-type-button.show > .dropdown-toggle  {
	border: none;
	border-radius: 0;
	padding: 0;
	color: inherit;
	background-color: inherit;
	border-color: inherit;
}
table.b-table>tfoot>tr>th.sorting:before,
table.b-table>thead>tr>th.sorting:before,
table.b-table>tfoot>tr>th.sorting:after,
table.b-table>thead>tr>th.sorting:after {
	top: 0;
}
</style>
