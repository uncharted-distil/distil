<template>
	<div class="search-view">
		<flow-bar
			left-text="Return to Home"
			:on-left="gotoHome"
			center-text="Search for a Dataset">
		</flow-bar>
		<search-bar class="search-search-bar"></search-bar>
		<h5 class="header-label">Select a Search Result</h5>
		<search-results class="search-search-results"></search-results>
	</div>
</template>

<script lang="ts">
import FlowBar from '../components/FlowBar.vue';
import SearchBar from '../components/SearchBar.vue';
import SearchResults from '../components/SearchResults.vue';
import { gotoHome } from '../util/nav';
import Vue from 'vue';

export default Vue.extend({
	name: 'search',
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
	margin: 0.75rem 0;
}
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
	width: 80%;
}
</style>
