<template>
	<div>
		<div class="row justify-content-center mt-2 mb-2">
			<h4>Distil</h4>
		</div>
		<div class="row justify-content-center mt-2 mb-2">
			<search-bar class="col-md-6"></search-bar>
		</div>
	</div>
</template>

<script>
import SearchBar from '../components/SearchBar';

export default {
	name: 'navigation',
	components: {
		SearchBar
	}
};
</script>

<style>
</style>
