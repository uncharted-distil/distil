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

export default {
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
			const items = this.$store.getters.getResultDataItems(this.showError);
			return items.filter(this.filterFunc)
				.map(this.decorateFunc);
		},
		// extract the table field header from the store
		fields() {
			return this.$store.getters.getResultDataFields(this.showError);
		}
	},

	methods: {
		onRowHovered(event) {
			// set new values
			const highlights = {};
			_.forIn(this.fields, (field, key) => {
				highlights[key] = event[key];
			});
			this.$store.dispatch('highlightFeatureValues', highlights);
		},

		onMouseOut() {
			this.$store.dispatch('clearFeatureHighlightValues');
		}
	}
};
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
</style>
