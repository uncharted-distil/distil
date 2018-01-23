<template>
	<div class="container-fluid d-flex flex-column h-100 search-view">
		<div class="row flex-0-nav">
		</div>
		<div class="row flex-1 align-items-center justify-content-center bg-white">
			<div class="col-12 col-md-10">
				<h5 class="header-label">Select a Search Result</h5>
			</div>
		</div>
		<div class="row flex-2 align-items-center justify-content-center">
			<div class="col-12 col-md-6">
				<search-bar class="search-search-bar"></search-bar>
			</div>
		</div>
		<div class="row flex-10 justify-content-center pb-3">
			<div class="col-12 col-md-10 d-flex">
				<search-results class="search-search-results"></search-results>
			</div>
		</div>
	</div>
</template>

<script lang="ts">
import SearchBar from '../components/SearchBar.vue';
import SearchResults from '../components/SearchResults.vue';
import { gotoHome } from '../util/nav';
import Vue from 'vue';

export default Vue.extend({
	name: 'search-view',
	components: {
		SearchBar,
		SearchResults
	},
	methods: {
		gotoHome() {
			gotoHome(this.$store, this.$router);
		}
	}
});
</script>

<style>
.header-label {
	padding: 1rem 0 0.5rem 0;
	font-weight: bold;
}
.search-search-bar {
	width: 100%;
	box-shadow: 0 1px 2px 0 rgba(0,0,0,0.10);
}
</style>
