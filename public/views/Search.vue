<template>
	<div id="search-view" class="container-fluid">
		<div class="row justify-content-center mt-2 mb-2">
			<h4>Distil</h4>
		</div>
		<div class="row justify-content-center mt-2 mb-2">
			<search-bar class="col-md-6"></search-bar>
		</div>
		<div class="row justify-content-center mt-2 mb-2">
			<search-results class="col-md-6"></search-results>
		</div>
	</div>
</template>

<script>
import SearchBar from '../components/SearchBar';
import SearchResults from '../components/SearchResults';

export default {
	name: 'search',
	components: {
		SearchBar,
		SearchResults
	},
	mounted() {
		// set search terms
		this.$store.commit('setSearchTerms', this.$route.query.terms || '');
		// send initial search
		this.$store.dispatch('searchDatasets', this.$store.getters.getSearchTerms());
	}
};
</script>

<style>
</style>
