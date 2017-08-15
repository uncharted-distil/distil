<template>
	<b-navbar toggleable type="light" variant="faded" fixed="top" class="bottom-shadowed">

		<b-nav-toggle target="nav_collapse"></b-nav-toggle>

		<i class="fa fa-rebel navbar-brand app-icon"></i>
		<span class="navbar-brand">Distil</span>

		<b-collapse is-nav id="nav_collapse">
			<b-nav is-nav-bar>
				<b-nav-item @click="onSearch" :active="searchActive">Search</b-nav-item>
				<b-nav-item @click="onData" :active="dataActive">Data</b-nav-item>
				<b-nav-item @click="onPipelines" :active="pipelinesActive">Pipelines</b-nav-item>
			</b-nav>
			<b-nav is-nav-bar class="ml-auto">
				<b-nav-text class="session-label">Session:</b-nav-text>
				<b-nav-text v-if="sessionID===null" class="session-not-ready">
					<i class="fa fa-close"></i>Unavailable
				</b-nav-text>
				<b-nav-text v-if="sessionID!==null" class="session-ready">
					<i class="fa fa-check"></i>{{sessionID}}
				</b-nav-text>
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
		this.$store.dispatch('getPipelineSession');
	},

	computed: {
		sessionID() {
			return this.$store.getters.getPipelineSessionID();
		}
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
		},
	},
	watch: {
		'$route.path'() {
			this.updateActive();
		}
	}
};

</script>

<style>
.session-not-ready {
  color: #cf3835 !important;
}
.session-ready {
  color: #00c07f !important;
}
.app-icon {
	color: #cf3835 !important;
}
.session-label {
	padding-right: 4px
}
.bottom-shadowed {
	-webkit-box-shadow: 0px 2px 5px -1px rgba(0,0,0,0.65);
	-moz-box-shadow: 0px 2px 5px -1px rgba(0,0,0,0.65);
	box-shadow: 0px 2px 5px -1px rgba(0,0,0,0.65);
}
</style>
