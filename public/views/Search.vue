<template>
	<div class="container-fluid h-100 search-view">
		<div class="row h-30">
			<div class="col-12 d-flex justify-content-center col-p-top">
				<flow-bar
					left-text="Return to Home"
					:on-left="gotoHome"
					center-text="Search for a dataset">
				</flow-bar>
			</div>
			<div class="col-12 d-flex justify-content-center">
				<search-bar class="search-search-bar"></search-bar>
			</div>
			<div class="col-12 d-flex justify-content-center">
				<h5 class="header-label">Select a Search Result</h5>
			</div>
		</div>
		<div class="row d-flex h-70">
			<div class="col-12 d-flex justify-content-center h-100">
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
	margin: 8px;
	width: 50%;
}
.search-search-results {
	width: 80%;
}
.h-70 {
	height: 70%;
}
.h-30 {
	height: 30%;
}
</style>
