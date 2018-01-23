<template>
	<div class="container-fluid d-flex flex-column h-100 home-view">
		<div class="row flex-0-nav">
		</div>	
		<div class="row flex-1 align-items-center justify-content-center bg-white">
			<div class="col-12 col-md-10">
				<h5 class="header-label">Recent Activity</h5>
			</div>
		</div>
		<div class="row flex-2 align-items-center justify-content-center">
			<div class="col-12 col-md-6 justify-content-center">
				<search-bar class="home-search-bar"></search-bar>
			</div>
		</div>
		<div class="row flex-10 justify-content-center pb-3">
			<div class="col-12 col-md-10 d-flex">
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
	padding: 1rem 0 0.5rem 0;
	font-weight: bold;
}
.home-search-bar {
	width: 100%;
	box-shadow: 0 1px 2px 0 rgba(0,0,0,0.10);
}
.home-items {
	overflow: auto;
}
.home-items .card {
	margin-bottom: 1rem;
}

</style>
