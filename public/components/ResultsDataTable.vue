<template>
	<div class="results-data-table">
		<p class="nav-link font-weight-bold">{{title}}</p>
		<p><small>Displaying {{items.length}} of {{numRows}} rows</small></p>
		<div class="results-data-table-container">
			<div class="results-data-no-results" v-if="!hasData">
				<div v-html="spinnerHTML"></div>
			</div>
			<div class="results-data-no-results" v-if="hasData && items.length===0">
				No results available
			</div>
			<b-table v-if="items.length>0"
				bordered
				hover
				small
				responsive
				:ref="refName"
				:items="items"
				:fields="fields"
				@row-clicked="onRowClick">
			</b-table>
		</div>

	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import { spinnerHTML } from '../util/spinner';
import { getters } from '../store/data/module';
import { TargetRow, TableRow, FieldInfo, RowSelection } from '../store/data/index';
import { getters as routeGetters } from '../store/route/module';
import { getters as solutionGetters } from '../store/solutions/module';
import { Dictionary } from '../util/dict';
import { removeNonTrainingItems, removeNonTrainingFields } from '../util/data';
import { updateRowSelection, clearRowSelection, updateTableRowSelection } from '../util/row';
import Vue from 'vue';

export default Vue.extend({
	name: 'results-data-table',

	props: {
		title: String,
		filterFunc: Function,
		decorateFunc: Function,
		refName: String,
		instanceName: { type: String, default: 'results-table-table' }
	},

	computed: {
		solutionId(): string {
			return routeGetters.getRouteSolutionId(this.$store);
		},

		numRows(): number {
			return getters.getResultDataNumRows(this.$store);
		},

		training(): Dictionary<boolean> {
			return solutionGetters.getActiveSolutionTrainingMap(this.$store);
		},

		hasData(): boolean {
			return getters.hasResultData(this.$store);
		},

		items(): TargetRow[] {
			const items = getters.getResultDataItems(this.$store);
			const filtered = removeNonTrainingItems(items, this.training);

			const selected = updateTableRowSelection(filtered, this.selectedRow, this.instanceName);

			return selected
				.filter(item => this.filterFunc(item))
				.map(item => this.decorateFunc(item));
		},

		fields(): Dictionary<FieldInfo> {
			const fields = getters.getResultDataFields(this.$store);
			return removeNonTrainingFields(fields, this.training);
		},

		selectedRow(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},

		selectedRowIndex(): number {
			return this.selectedRow ? this.selectedRow.index : -1;
		},

		spinnerHTML(): string {
			return spinnerHTML();
		}
	},

	methods: {

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
	}

});
</script>

<style>

.results-data-table {
	display: flex;
	flex-direction: column;
}
.results-data-table-container {
	display: flex;
	overflow: auto;
}
.results-data-no-results {
	width: 100%;
	background-color: #eee;
	padding: 8px;
	text-align: center;
}
table tr {
	cursor: pointer;
}

</style>
