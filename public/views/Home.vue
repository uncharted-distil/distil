<template>
	<div class="home-view">
		<flow-bar
			center-text="Search for a dataset"
			right-text="Continue to dataset Search"
			:on-right="gotoSearch">
		</flow-bar>
		<search-bar class="home-search-bar"></search-bar>
		<h5 class="header-label">Recent Activity</h5>
		<div class="home-items">
			<recent-datasets
				:max-datasets="5"></recent-datasets>
			<recent-pipelines
				:max-pipelines="5"></recent-pipelines>
			<running-pipelines
				:max-pipelines="5"></running-pipelines>
		</div>
	</div>
</template>

<script lange="ts">
import FlowBar from '../components/FlowBar';
import RecentDatasets from '../components/RecentDatasets';
import RecentPipelines from '../components/RecentPipelines';
import RunningPipelines from '../components/RunningPipelines';
import SearchBar from '../components/SearchBar';
import { gotoSearch } from '../util/nav';
import { actions, getters } from '../store/pipelines/module';
import Vue from 'vue';

export default Vue.extend({
	name: 'home-view',
	components: {
		FlowBar,
		RecentDatasets,
		RecentPipelines,
		RunningPipelines,
		SearchBar
	},
	computed: {
		sessionId() {
			return getters.getPipelineSessionID(this.$store);
		}
	},
	mounted() {
		actions.fetchPipelines(this.$store, {
			sessionId: this.sessionId
		});
	},
	methods: {
		gotoSearch() {
			gotoSearch(this.$store, this.$router);
		}
	}

});
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
	width: 80%;
	overflow: auto;
	margin-bottom: 4px;
}
.home-items .card {
	margin-bottom: 4px;
}
</style>
