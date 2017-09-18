<template>
	<div class="search-view">
		<search-bar class="search-search-bar"></search-bar>
		<search-results class="search-search-results"></search-results>
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
.search-view {
	display: flex;
	flex-direction: column;
	align-items: center;
}
.search-search-bar {
	margin: 8px;
	width: 50%;
}
.search-search-results {
	display: flex;
	width: 50%;
	flex-direction: column;
}
</style>
