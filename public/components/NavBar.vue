<template>
	<b-navbar toggleable type="light" variant="faded" fixed="top" class="bottom-shadowed">

		<b-nav-toggle target="nav_collapse"></b-nav-toggle>

		<i class="fa fa-rebel navbar-brand app-icon"></i>
		<span class="navbar-brand">Distil</span>

		<b-collapse is-nav id="nav_collapse">
			<b-nav is-nav-bar>
				<b-nav-item @click="onSearch" :active="activeView===SEARCH">Search</b-nav-item>
				<b-nav-item @click="onData" :active="activeView===DATASETS">Data</b-nav-item>
				<b-nav-item @click="onPipelines" :active="activeView===PIPELINES">Pipelines</b-nav-item>
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

const SEARCH = Symbol();
const DATASETS = Symbol();
const PIPELINES = Symbol();
const ROUTE_MAPPINGS = {
	'/search': SEARCH,
	'/dataset': DATASETS,
	'/pipelines': PIPELINES
};

export default {
	name: 'nav-bar',

	data() {
		return {
			SEARCH: SEARCH,
			DATASETS: DATASETS,
			PIPELINES: PIPELINES,
			activeView: SEARCH,
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
			const entry = createRouteEntry('/search', {
				terms: this.$store.getters.getRouteTerms()
			});
			this.$router.push(entry);
		},

		// switch to data view
		onData() {
			const entry = createRouteEntry('/dataset',{
				dataset: this.$store.getters.getRouteDataset()
			});
			this.$router.push(entry);
		},

		// switch to the pipelines view
		onPipelines() {
			const entry = createRouteEntry('/pipelines', {
				dataset: this.$store.getters.getRouteDataset(),
				filters: this.$store.getters.getRouteFilters()
			});
			this.$router.push(entry);
		},

		updateActive() {
			this.activeView = ROUTE_MAPPINGS[this.$route.path];
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
