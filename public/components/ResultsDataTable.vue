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
				striped
				small
				@row-hovered="onRowHovered"
				@mouseout.native="onMouseOut"
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
import { updateTableHighlights } from '../util/highlights';
import { getTrainingVariablesForPipelineId } from '../util/pipelines';
import Vue from 'vue';

const RESULT_TABLE_HIGHLIGHTS = 'result_table';

export default Vue.extend({
	name: 'results-data-table',

	props: {
		'title': String,
		'filterFunc': Function,
		'decorateFunc': Function,
		'excludeNonTraining': Boolean
	},

	computed: {
		pipelineId(): string {
			return routeGetters.getRoutePipelineId(this.$store);
		},
		// extracts the table data from the store
		items(): TargetRow[] {
			const items = getters.getResultDataItems(this.$store);
			const filtered = this.excludeNonTraining ? removeNonTrainingItems(items, this.training) : items;
			const rangeHighlights = getters.getHighlightedFeatureRanges(this.$store);
			const valueHighlights = getters.getHighlightedFeatureValues(this.$store);
			updateTableHighlights(filtered, rangeHighlights, valueHighlights, RESULT_TABLE_HIGHLIGHTS);
			return filtered
				.filter(item => this.filterFunc(item))
				.map(item => this.decorateFunc(item));
		},

		// extract the table field header from the store
		fields(): Dictionary<FieldInfo> {
			const fields = getters.getResultDataFields(this.$store);
			return this.excludeNonTraining ? removeNonTrainingFields(fields, this.training) : fields;
		},

		training(): Dictionary<boolean> {
			const training = getTrainingVariablesForPipelineId(this.$store.state.pipelineModule, this.pipelineId);
			const trainingMap = {};
			training.forEach(t => {
				trainingMap[t.toLowerCase()] = true;
			});
			return trainingMap;
		},
	},

	methods: {
		onRowHovered(event: Event) {
			// set new values
			const highlights = {
				context: RESULT_TABLE_HIGHLIGHTS,
				values: {}
			};
			_.forIn(this.fields, (field, key) => highlights.values[key] = event[key]);
			mutations.highlightFeatureValues(this.$store, highlights);
		},

		onMouseOut() {
			mutations.clearFeatureHighlightValues(this.$store);
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
