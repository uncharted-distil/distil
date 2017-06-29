<template>
	<div class="row justify-content-center mt-2 mb-2">
		<search-results class="col-md-6"></search-results>
	</div>
</template>

<script>
import SearchResults from '../components/SearchResults';

export default {
	name: 'search',
	components: {
		SearchResults
	}
};
</script>

<style>
</style>
