<template>
	<div class="home-view">
		<flow-bar
			center-text="Search for a Dataset"
			right-text="Continue to Dataset Search"
			:on-right="gotoSearch">
		</flow-bar>
		<search-bar class="home-search-bar"></search-bar>
		<h4 class="header-label">Recent Activity</h4>
		<div class="home-items">
			<recent-datasets></recent-datasets>
			<recent-pipelines></recent-pipelines>
			<running-pipelines></running-pipelines>
		</div>
	</div>
</template>

<script>
import FlowBar from '../components/FlowBar';
import RecentDatasets from '../components/RecentDatasets';
import RecentPipelines from '../components/RecentPipelines';
import RunningPipelines from '../components/RunningPipelines';
import SearchBar from '../components/SearchBar';
import { gotoSearch } from '../util/nav';

export default {
	name: 'home',
	components: {
		FlowBar,
		RecentDatasets,
		RecentPipelines,
		RunningPipelines,
		SearchBar
	},
	methods: {
		gotoSearch() {
			gotoSearch(this.$store, this.$router);
		}
	}
};
</script>

<style>
.header-label {
	color: #333;
	margin: 0.75rem 0;
}
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
	overflow: auto;
	margin-bottom: 4px;
}
.home-items .card {
	margin-bottom: 4px;
}
</style>
