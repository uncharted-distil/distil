<template>
	<div class="search-view">
		<flow-bar
			left-text="Return to Home"
			:on-left="gotoHome"
			center-text="Search for a dataset">
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
	padding: 1rem 0 0.5rem 0;
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
