<template>
	<div class="explore-data-table">
		<h6 class="nav-link">Samples</h6>
		<div class="explore-data-table-container">
			<div class="explore-data-no-results" v-if="items.length===0">
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
	name: 'explore-data-table',

	computed: {
		// get dataset from route
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		// extracts the table data from the store
		items() {
			return this.$store.getters.getFilteredDataItems();
		},
		// extract the table field header from the store
		fields() {
			return this.$store.getters.getFilteredDataFields();
		},
		filters() {
			return this.$store.getters.getFilters();
		}
	},

	mounted() {
		this.fetch();
	},

	watch: {
		'$route.query.filters'() {
			this.fetch();
		},
		'$route.query.dataset'() {
			this.fetch();
		}
	},

	methods: {
		fetch() {
			this.$store.dispatch('updateFilteredData', {
				dataset: this.dataset,
				filters: this.filters
			});
		},
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

.explore-data-table {
	display: flex;
	flex-direction: column;
}
.explore-data-table-container {
	display: flex;
	overflow: auto;
}
.explore-data-no-results {
	width: 100%;
	background-color: #eee;
	padding: 8px;
}
</style>
