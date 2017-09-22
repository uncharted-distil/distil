<template>
	<div class="data-table">
		<div class="bg-faded rounded-top">
			<h6 class="nav-link">Values</h6>
		</div>
		<div class="data-table-container">
			<div v-if="items.length===0">
				No results
			</div>
			<b-table v-if="items.length>0"
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
			return this.$store.getters.getFilteredDataItems();
		},
		// extract the table field header from the store
		fields() {
			return this.$store.getters.getFilteredDataFields();
		}
	}
};
</script>

<style>

.data-table {
	display: flex;
	flex-direction: column;
}
.data-table-container {

	display: flex;
	overflow: auto;
}
</style>
