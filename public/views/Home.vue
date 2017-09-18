<template>
	<div class="home-view">
		<search-bar class="home-search-bar"></search-bar>
		<div class="home-items">
			<recent-datasets></recent-datasets>
			<recent-pipelines></recent-pipelines>
			<running-pipelines></running-pipelines>
		</div>
	</div>
</template>

<script>
import RecentDatasets from '../components/RecentDatasets';
import RecentPipelines from '../components/RecentPipelines';
import RunningPipelines from '../components/RunningPipelines';
import SearchBar from '../components/SearchBar';

export default {
	name: 'home',
	components: {
		RecentDatasets,
		RecentPipelines,
		RunningPipelines,
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
	overflow: auto;
	margin-bottom: 4px;
}
.home-items .card {
	margin-bottom: 4px;
}
</style>
