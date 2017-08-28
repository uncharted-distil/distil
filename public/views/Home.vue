<template>
	<div>
		<div class="row justify-content-center mt-2 mb-2">
			<search-bar class="col-md-6"></search-bar>
		</div>
		<div class="row justify-content-center mt-2 mb-2">
			<div class="col-md-6">Running Pipelines</div>
		</div>
		<div class="row justify-content-center mt-2 mb-2">
			<div class="col-md-6">Recent Data Sets</div>
		</div>
		<div class="row justify-content-center mt-2 mb-2">
			<div class="col-md-6">Recent Pipelines</div>
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
</style>
