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
					<recent-solutions
						:max-solutions="5"></recent-solutions>
					<running-solutions
						:max-solutions="5"></running-solutions>
				</div>
			</div>
		</div>
		<div class="home-version-text text-muted">
			{{version}}
		</div>
	</div>
</template>

<script lang="ts">
import RecentDatasets from '../components/RecentDatasets';
import RecentSolutions from '../components/RecentSolutions';
import RunningSolutions from '../components/RunningSolutions';
import SearchBar from '../components/SearchBar';
import { actions } from '../store/solutions/module';
import { getters as appGetters } from '../store/app/module';
import Vue from 'vue';

export default Vue.extend({
	name: 'home-view',

	components: {
		RecentDatasets,
		RecentSolutions,
		RunningSolutions,
		SearchBar
	},

	computed: {
		version(): string {
			return `version: ${appGetters.getVersionNumber(this.$store)} at ${appGetters.getVersionTimestamp(this.$store)}`;
		}
	},

	beforeMount() {
		this.fetch();
	},

	methods: {
		fetch() {
			actions.fetchSolutions(this.$store, {});
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
	width: 100%;
	overflow: auto;
}
.home-items .card {
	margin-bottom: 1rem;
}
.home-version-text {
	margin: 0 auto;
	font-size: 0.8rem;
}

</style>
