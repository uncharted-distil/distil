<template>
	<div class="results-data-table">
		<div class="bg-faded rounded-top">
			<h6 class="nav-link">Values</h6>
		</div>
		<div class="results-data-table-container">
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
	name: 'results-data-table',

	data() {
		return {
			perPage: 10,
			currentPage: 1
		};
	},

	mounted() {
		this.$store.dispatch('updateFilteredData', this.dataset).then(() => {
			this.$store.dispatch('updateResults', {
				dataset: this.dataset,
				resultId: atob(this.$store.getters.getRouteResultId())
			});
		});
	},

	watch: {
		// if filters change, update data
		'$route.query'() {
			this.$store.dispatch('updateFilteredData', this.dataset).then(() => {
				this.$store.dispatch('updateResults', {
					dataset: this.dataset,
					resultId: atob(this.$store.getters.getRouteResultId())
				});
			});
		}
	},

	computed: {
		// get dataset from route
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		// extracts the table data from the store
		items() {
			return this.$store.getters.getResultDataItems();
		},
		// extract the table field header from the store
		fields() {
			return this.$store.getters.getResultDataFields();
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
</style>
