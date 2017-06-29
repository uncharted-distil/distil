<template>
	<div class="data-table">
		<div class="nav bg-faded rounded-top">
			<h6 class="nav-link">Values</h6>
		</div>
		<div class="table-container">
			<div v-if="items.length===0">
				No results
			</div>
			<b-table v-if="items.length>0"
				responsive
				bordered
				hover
				striped
				small
				:items="items"
				:fields="fields"
				:current-page="currentPage">
			</b-table>
		</div>
	</div>
</template>

<script>

export default {
	name: 'data-table',

	data() {
		return {
			perPage: 10,
			currentPage: 1
		};
	},

	mounted() {
		this.$store.dispatch('updateFilteredData', this.dataset);
	},

	watch: {
		// if dataset changes, clear filter state
		'$route.query.dataset'() {
		},
		// if filters change, update data
		'$route.query'() {
			this.$store.dispatch('updateFilteredData', this.dataset);
		}
	},

	computed: {
		// get dataset from route
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		// extracts the table data from the store
		items() {
			return this.$store.getters.getFilteredDataItems(this.dataset);
		},
		// extract the table field header from the store
		fields() {
			return this.$store.getters.getFilteredDataFields(this.dataset);
		}
	}
};
</script>

<style>
.table-container {
	overflow: auto;
}
</style>
