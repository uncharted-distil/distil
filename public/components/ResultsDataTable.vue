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
				:ref="refName"
				:items="items"
				:fields="fields">
			</b-table>
		</div>

	</div>
</template>

<script lang="ts">

import { getters } from '../store/data/module';
import { TargetRow, FieldInfo } from '../store/data/index';
import { getters as routeGetters } from '../store/route/module';
import { getters as pipelineGetters } from '../store/pipelines/module';
import { Dictionary } from '../util/dict';
import { removeNonTrainingItems, removeNonTrainingFields } from '../util/data';
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

		training(): Dictionary<boolean> {
			return pipelineGetters.getActivePipelineTrainingMap(this.$store);
		},

		hasData(): boolean {
			return getters.hasResultData(this.$store);
		},

		items(): TargetRow[] {
			const items = getters.getResultDataItems(this.$store);
			const filtered = removeNonTrainingItems(items, this.training);
			return filtered
				.filter(item => this.filterFunc(item))
				.map(item => this.decorateFunc(item));
		},

		fields(): Dictionary<FieldInfo> {
			const fields = getters.getResultDataFields(this.$store);
			return removeNonTrainingFields(fields, this.training);
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
