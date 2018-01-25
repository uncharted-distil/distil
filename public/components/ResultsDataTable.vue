<template>
	<div class="results-data-table">
		<p class="nav-link font-weight-bold">{{title}}</p>
		<p><small>Displaying {{items.length}} of {{numRows}} rows</small></p>
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
import { getters as routeGetters } from '../store/route/module';
import { Dictionary } from '../util/dict';
import { removeNonTrainingItems, removeNonTrainingFields } from '../util/data';
import { updateTableHighlights, scrollToFirstHighlight } from '../util/highlights';
import { getTrainingVariablesForPipelineId } from '../util/pipelines';
import Vue from 'vue';

export default Vue.extend({
	name: 'results-data-table',

	props: {
		title: String,
		filterFunc: Function,
		decorateFunc: Function,
		excludeNonTraining: Boolean,
		refName: String,
		instanceName: { type: String, default: 'results-table-table' }
	},

	data() {
		return {
			selectedRowKey: -1
		};
	},

	computed: {
		pipelineId(): string {
			return routeGetters.getRoutePipelineId(this.$store);
		},

		numRows(): number {
			return getters.getResultDataNumRows(this.$store);
		},

		// extracts the training set from the store
		training(): Dictionary<boolean> {
			const training = getTrainingVariablesForPipelineId(this.$store.state.pipelineModule, this.pipelineId);
			const trainingMap = {};
			training.forEach(t => {
				trainingMap[t.toLowerCase()] = true;
			});
			return trainingMap;
		},

		// extracts the table data from the store
		items(): TargetRow[] {
			const items = getters.getResultDataItems(this.$store);
			const filtered = this.excludeNonTraining ? removeNonTrainingItems(items, this.training) : items;
			const valueHighlights = getters.getHighlightedFeatureValues(this.$store);

			// clear all selections visuals
			items.forEach(r => r._rowVariant = null);

			// if we have highlights defined and the select table is not the source then updated
			// the highlight visuals.
			let updatedItems = <TargetRow[]>[];
			if (_.get(valueHighlights, 'root', 'context') !== this.instanceName) {
				updateTableHighlights(filtered, valueHighlights, this.instanceName);
			}

			updatedItems = filtered
				.filter(item => this.filterFunc(item))
				.map(item => this.decorateFunc(item));

			if (this.selectedRowKey >= 0) {
				const toSelect = updatedItems.find(r => r._key === this.selectedRowKey);
				if (_.get(valueHighlights, 'root.context') === this.instanceName) {
					toSelect._rowVariant = 'primary';
				} else {
					toSelect._rowVariant = null;
					this.selectedRowKey = -1;
				}
			}

			// On data / highlights change, scroll to first selected row
			scrollToFirstHighlight(this, this.refName, false);

			return updatedItems;
		},

		// extract the table field header from the store
		fields(): Dictionary<FieldInfo> {
			const fields = getters.getResultDataFields(this.$store);
			return this.excludeNonTraining ? removeNonTrainingFields(fields, this.training) : fields;
		}
	},

	methods: {
		onRowClick(row: TargetRow) {
			if (row._key !== this.selectedRowKey) {
				// clicked on a different row than last time - new selection
				this.selectedRowKey = row._key;

				// publish the highlight change
				const highlights = {
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
				// clicked on same row - remove the row selection visual
				this.selectedRowKey = -1;
				mutations.clearFeatureHighlights(this.$store);
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
