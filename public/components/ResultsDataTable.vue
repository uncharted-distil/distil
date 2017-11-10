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

<script>
import _ from 'lodash';
import { getters, actions } from '../store/data/module';
import Vue from 'vue';

export default Vue.extend({
	name: 'results-data-table',

	props: [
		'title',
		'filterFunc',
		'decorateFunc',
		'showError'
	],

	computed: {
		// extracts the table data from the store
		items() {
			const items = getters.getResultDataItems(this.$store);
			return items
				.filter(this.filterFunc)
				.map(this.decorateFunc);
		},
		// extract the table field header from the store
		fields() {
			return getters.getResultDataFields(this.$store);
		}
	},

	methods: {
		onRowHovered(event) {
			// set new values
			const highlights = {};
			_.forIn(this.fields, (field, key) => {
				highlights[key] = event[key];
			});
			actions.highlightFeatureValues(this.$store, highlights);
		},

		onMouseOut() {
			actions.clearFeatureHighlightValues(this.$store);
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
