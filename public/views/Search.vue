<template>
	<div class="container-fluid d-flex flex-column h-100 search-view">
		<div class="row flex-0-nav">
		</div>
		<div class="row flex-1 align-items-center justify-content-center">
			<div class="col-12 d-flex">
				<flow-bar
					left-text="Return to Home"
					:on-left="gotoHome"
					center-text="Search for a dataset">
				</flow-bar>
			</div>
		</div>
		<div class="row flex-1 align-items-center justify-content-center">
			<div class="col-12 col-md-6 d-flex justify-content-cente">
				<search-bar class="search-search-bar"></search-bar>
			</div>
		</div>
		<div class="row flex-1 align-items-center justify-content-center">
			<div class="col-12 col-md-10 d-flex">
				<h5 class="header-label">Select a Search Result</h5>
			</div>
		</div>
		<div class="row flex-11 justify-content-center">
			<div class="col-12 col-md-10 d-flex mb-3">
				<search-results class="search-search-results"></search-results>
			</div>
		</div>
	</div>
</template>

<script lang="ts">
import FlowBar from '../components/FlowBar.vue';
import SearchBar from '../components/SearchBar.vue';
import SearchResults from '../components/SearchResults.vue';
import { gotoHome } from '../util/nav';
import Vue from 'vue';

export default Vue.extend({
	name: 'search-view',
	components: {
		FlowBar,
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
	color: #333;
}
.search-search-bar {
	width: 100%;
}

</style>
