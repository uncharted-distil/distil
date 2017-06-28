<template>
	<div id="dataset-view" class="container-fluid">
		<div class="row justify-content-center mt-2 mb-2">
			<h4>Distil</h4>
		</div>
		<div class="row justify-content-center mt-2 mb-2">
			<search-bar class="col-md-6"></search-bar>
		</div>
		<div class="row mt-2 mb-2">
			<data-table class="col-md-8"></data-table>
			<variable-summaries class="col-md-4"></variable-summaries>
		</div>
	</div>
</template>

<script>
import SearchBar from '../components/SearchBar';
import DataTable from '../components/DataTable';
import VariableSummaries from '../components/VariableSummaries';

export default {
	name: 'dataset',
	components: {
		SearchBar,
		DataTable,
		VariableSummaries
	},
	mounted() {
		// set active dataset
		this.$store.commit('setActiveDataset', this.$route.query.dataset);
		// clear the filter state
		this.$store.commit('setFilterState', {});
		// update filtered data
		this.$store.dispatch('updateFilteredData', this.$store.getters.getActiveDataset());
		// get variable summaries
		this.$store.dispatch('getVariableSummaries', this.$store.getters.getActiveDataset());
	}
};
</script>

<style>
</style>
