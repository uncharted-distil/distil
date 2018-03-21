<template>
	<div class="results-data-table">
		<p class="nav-link font-weight-bold">{{title}}</p>
		<p><small>Displaying {{items.length}} of {{numRows}} rows</small></p>
		<div class="results-data-table-container">
			<div class="results-data-no-results" v-if="!hasData">
				<div class="bounce1"></div>
				<div class="bounce2"></div>
				<div class="bounce3"></div>
			</div>
			<div class="results-data-no-results" v-if="hasData && items.length===0">
				No results available
			</div>
			<b-table v-if="items.length>0"
				bordered
				hover
				small
				responsive
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
import { getters } from '../store/data/module';
import { TargetRow, FieldInfo } from '../store/data/index';
import { getters as routeGetters } from '../store/route/module';
import { getters as pipelineGetters } from '../store/pipelines/module';
import { Dictionary } from '../util/dict';
import { removeNonTrainingItems, removeNonTrainingFields } from '../util/data';
import { updateTableHighlights, updateHighlightRoot, clearHighlightRoot, scrollToFirstHighlight, getHighlights } from '../util/highlights';
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
		pipelineId(): string {
			return routeGetters.getRoutePipelineId(this.$store);
		},

		numRows(): number {
			return getters.getResultDataNumRows(this.$store);
		},

		selectedRowKey(): number {
			return routeGetters.getDecodedHighlightRoot(this.$store) ? _.toNumber(routeGetters.getDecodedHighlightRoot(this.$store).key) : -1;
		},

		training(): Dictionary<boolean> {
			return pipelineGetters.getActivePipelineTrainingMap(this.$store);
		},

		hasData(): boolean {
			return getters.hasResultData(this.$store);
		},

		// extracts the table data from the store
		items(): TargetRow[] {
			const items = getters.getResultDataItems(this.$store);
			const filtered = removeNonTrainingItems(items, this.training);
			const highlights = getHighlights(this.$store);

			// clear all selections visuals
			items.forEach(r => r._rowVariant = null);

			// if we have highlights defined and the select table is not the source then updated
			// the highlight visuals.
			let updatedItems = <TargetRow[]>[];
			if (_.get(highlights, 'root.context') !== this.instanceName) {
				updateTableHighlights(filtered, highlights, this.instanceName);
			}

			updatedItems = filtered
				.filter(item => this.filterFunc(item))
				.map(item => this.decorateFunc(item));

			if (this.selectedRowKey >= 0) {
				const toSelect = updatedItems.find(r => r._key === this.selectedRowKey);
				if (toSelect) {
					if (_.get(highlights, 'root.context') === this.instanceName) {
						toSelect._rowVariant = 'primary';
					} else {
						toSelect._rowVariant = null;
					}
				}
			}

			// On data / highlights change, scroll to first selected row
			scrollToFirstHighlight(this, this.refName, false);

			return updatedItems;
		},

		// extract the table field header from the store
		fields(): Dictionary<FieldInfo> {
			const fields = getters.getResultDataFields(this.$store);
			return removeNonTrainingFields(fields, this.training);
		}
	},

	methods: {
		onRowClick(row: TargetRow) {
			if (row._key !== this.selectedRowKey) {
				// clicked on a different row than last time - new selection
				updateHighlightRoot(this, {
					context: this.instanceName,
					key: row._key.toString(),
					value: _.map(this.fields, (field, key) => [ key, row[key] ])
				});
			} else {
				// clicked on same row - remove the row selection visual
				clearHighlightRoot(this);
			}
		}
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

</style>
