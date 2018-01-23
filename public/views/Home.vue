<template>
	<div class="container-fluid d-flex flex-column h-100 home-view">
		<div class="row flex-0-nav">
		</div>
		<div class="row flex-1 align-items-center">
			<div class="col-12 d-flex justify-content-center">
				<flow-bar
					center-text="Search for a dataset"
					right-text="Continue to dataset Search"
					:on-right="gotoSearch">
				</flow-bar>
			</div>
		</div>	
		<div class="row flex-1 align-items-center justify-content-center">
			<div class="col-12 col-md-6 d-flex justify-content-center">
				<search-bar class="home-search-bar"></search-bar>
			</div>
		</div>
		<div class="row flex-1 align-items-center justify-content-center">
			<div class="col-12 col-md-10 d-flex">
				<h5 class="header-label">Recent Activity</h5>
			</div>
		</div>
		<div class="row flex-11 justify-content-center">
			<div class="col-12 col-md-10 d-flex mb-3">
				<div class="home-items">
					<recent-datasets
						:max-datasets="5"></recent-datasets>
					<recent-pipelines
						:max-pipelines="5"></recent-pipelines>
					<running-pipelines
						:max-pipelines="5"></running-pipelines>
				</div>
			</div>
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
}
.home-search-bar {
	width: 100%;
}
.home-items {
	overflow: auto;
}
.home-items .card {
	margin-bottom: 1rem;
}

</style>
