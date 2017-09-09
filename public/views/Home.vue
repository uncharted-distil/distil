<template>
	<div class="home-view">
		<search-bar class="home-search-bar"></search-bar>
		<div class="home-items">
			<div>Running Pipelines</div>
			<div>Recent Data Sets</div>
			<div>Recent Pipelines</div>
		</div>
	</div>
</template>

<script>
import SearchBar from '../components/SearchBar';
export default {
	name: 'home',
	components: {
		SearchBar
	}
};
</script>

<style>
.home-view {
	display: flex;
	flex-direction: column;
	align-items: center;
}
.home-search-bar {
	margin: 8px;
	width: 50%;
}
.home-items {
	width: 50%;
}
</style>
