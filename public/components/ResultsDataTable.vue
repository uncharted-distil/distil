<template>
	<div class="results-data-table">
		<h6 class="nav-link">{{title}}</h6>
		<div class="results-data-table-container">
			<div class="results-data-no-results" v-if="items.length===0">
				No results
			</div>
			<b-table v-if="items.length>0"
				bordered
				hover
				small
				@row-clicked="onRowClick"
				:ref="refName"
				:items="items"
				:fields="fields">
			</b-table>
		</div>

	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import { getters, mutations } from '../store/data/module';
import { TargetRow, FieldInfo } from '../store/data/index';
import { Dictionary } from '../util/dict';
import { removeNonTrainingItems, removeNonTrainingFields } from '../util/data';
import { updateTableHighlights, scrollToFirstHighlight } from '../util/highlights';
import Vue from 'vue';

const RESULT_TABLE_HIGHLIGHTS = 'result_table';

export default Vue.extend({
	name: 'results-data-table',

	props: {
		'title': String,
		'filterFunc': Function,
		'decorateFunc': Function,
		'excludeNonTraining': Boolean,
		'refName': String
	},

	data() {
		return {
			selectedRowKey: -1
		};
	},

	computed: {
		// extracts the table data from the store
		items(): TargetRow[] {
			const items = getters.getResultDataItems(this.$store);

			const training = getters.getTrainingVariablesMap(this.$store);
			const filtered = this.excludeNonTraining ? removeNonTrainingItems(items, training) : items;

			const rangeHighlights = getters.getHighlightedFeatureRanges(this.$store);
			const valueHighlights = getters.getHighlightedFeatureValues(this.$store);

			// clear all selections visuals
			items.forEach(r => r._rowVariant = null);

			// if we have highlights defined and the select table is not the source then updated
			// the highlight visuals.
			if ((valueHighlights.context && valueHighlights.context !== RESULT_TABLE_HIGHLIGHTS) ||
				(rangeHighlights.context && rangeHighlights.context !== RESULT_TABLE_HIGHLIGHTS)) {
					updateTableHighlights(filtered, rangeHighlights, valueHighlights, RESULT_TABLE_HIGHLIGHTS);
			}

			const updatedItems = filtered
				.filter(item => this.filterFunc(item))
				.map(item => this.decorateFunc(item));

			// apply the currently selected row highlight - use the key because it is invarant across filter/sort
			// apply the currently selected row highlight - if there were value or range highlights applied,
			// then disable row selection
			if (this.selectedRowKey >= 0 &&
				valueHighlights.context === RESULT_TABLE_HIGHLIGHTS ||
				rangeHighlights.context === RESULT_TABLE_HIGHLIGHTS) {
				const toSelect = updatedItems.find(r => r._key === this.selectedRowKey);
				toSelect._rowVariant = 'primary';
			} else {
				this.selectedRowKey = -1;
			}


			// On data / highlights change, scroll to first selected row
			scrollToFirstHighlight(this, this.refName);

			return updatedItems;
		},

		// extract the table field header from the store
		fields(): Dictionary<FieldInfo> {
			const fields = getters.getResultDataFields(this.$store);
			const training = getters.getTrainingVariablesMap(this.$store);
			return this.excludeNonTraining ? removeNonTrainingFields(fields, training) : fields;
		}
	},

	methods: {
		onRowClick(row: TargetRow) {

			// clear out any highlights currently in the table and at the app level
			mutations.clearFeatureHighlights(this.$store);

			if (row._key !== this.selectedRowKey) {
				// clicked on a different row than last time - new selection
				this.selectedRowKey = row._key;

				// publish the highlight change
				const highlights = {
					context: RESULT_TABLE_HIGHLIGHTS,
					values: {}
				};
				_.forEach(this.fields, (field, key) => highlights.values[key] = row[key]);
				mutations.highlightFeatureValues(this.$store, highlights);
			} else {
				// clicked on same row - remove the row selection visual
				this.selectedRowKey = -1;
			}
		}
	}
});
</script>

<style>

results-data-table {
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
}
.table-sm th, .table-sm td {
	font-size: 0.9rem;
}
</style>
