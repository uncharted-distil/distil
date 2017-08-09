<template>
	<div>
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
	}
};
</script>

<style>
</style>
