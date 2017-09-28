<template>
	<div class="select-data-table">
		<div class="bg-faded rounded-top">
			<h6 class="nav-link">Values</h6>
		</div>
		<div class="select-data-table-container">
			<div v-if="items.length===0">
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
	name: 'selected-data-table',

	computed: {
		// get dataset from route
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		// extracts the table data from the store
		items() {
			return this.$store.getters.getSelectedDataItems();
		},
		// extract the table field header from the store
		fields() {
			return this.$store.getters.getSelectedDataFields();
		},
		filters() {
			return this.$store.getters.getSelectedFilters();
		}
	},

	mounted() {
		this.fetch();
	},

	watch: {
		'$route.query.training'() {
			this.fetch();
		},
		'$route.query.target'() {
			this.fetch();
		}
	},

	methods: {
		fetch() {
			this.$store.dispatch('updateSelectedData', {
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

.select-data-table {
	display: flex;
	flex-direction: column;
}
.select-data-table-container {

	display: flex;
	overflow: auto;
}
</style>
