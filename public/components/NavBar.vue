<template>
	<b-navbar toggleable type="inverse" variant="primary" fixed="top">

		<b-nav-toggle target="nav_collapse"></b-nav-toggle>

		<span class="navbar-brand">Distil</span>

		<b-collapse is-nav id="nav_collapse">
			<b-nav is-nav-bar>
				<b-nav-item @click="onSearch" :active="searchActive">Search</b-nav-item>
				<b-nav-item @click="onData" :active="dataActive">Data</b-nav-item>
				<b-nav-item @click="onPipelines" :active="pipelinesActive">Pipelines</b-nav-item>
			</b-nav>
		</b-collapse>
	</b-navbar>
</template>

<script>

import { createRouteEntry } from '../util/routes';

export default {
	name: 'nav-bar',

	data() {
		return {
			searchActive: false,
			dataActive: false,
			pipelinesActive: false
		};
	},

	mounted() {
		this.updateActive();
	},

	methods: {
		// switch to the search view
		onSearch() {
			const entry = createRouteEntry(
				'/search',
				this.$store.getters.getRouteDataset(),
				this.$store.getters.getRouteTerms(),
				this.$store.getters.getRouteFilters());
			this.$router.push(entry);
		},

		// switch to data view
		onData() {
			const entry = createRouteEntry(
				'/dataset',
				this.$store.getters.getRouteDataset(),
				this.$store.getters.getRouteTerms(),
				this.$store.getters.getRouteFilters());
			this.$router.push(entry);
		},

		// switch to the pipelines view
		onPipelines() {
			const entry = createRouteEntry(
				'/pipelines',
				this.$store.getters.getRouteDataset(),
				this.$store.getters.getRouteTerms(),
				this.$store.getters.getRouteFilters());
			this.$router.push(entry);
		},

		updateActive() {
			if (this.$route.path === '/pipelines') {
				this.pipelinesActive = true;
				this.searchActive = false;
				this.dataActive = false;
			} else if (this.$route.path === '/dataset') {
				this.pipelinesActive = false;
				this.searchActive = false;
				this.dataActive = true;
			} else if (this.$route.path === '/search') {
				this.pipelinesActive = false;
				this.searchActive = true;
				this.dataActive = false;
			}
		}
	},
	watch: {
		'$route.path'() {
			this.updateActive();
		}
	}
};

</script>

<style>

</style>
