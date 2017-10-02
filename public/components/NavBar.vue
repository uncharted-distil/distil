<template>
	<b-navbar toggleable type="light" variant="faded" class="bottom-shadowed">

		<b-nav-toggle target="nav_collapse"></b-nav-toggle>

		<i class="fa fa-rebel navbar-brand app-icon"></i>
		<span class="navbar-brand">Distil</span>

		<b-collapse is-nav id="nav_collapse">
			<b-nav is-nav-bar>
				<b-nav-item @click="onHome" :active="activeView===HOME">Home</b-nav-item>
				<b-nav-item @click="onSearch" :active="activeView===SEARCH">Search</b-nav-item>
				<b-nav-item @click="onExplore" :active="activeView===EXPLORE" :disabled="!hasDataset()">Explore</b-nav-item>
				<b-nav-item @click="onSelect" :active="activeView===SELECT" :disabled="!hasDataset()">Select</b-nav-item>
				<b-nav-item @click="onPipelines" :active="activeView===PIPELINES" :disabled="!hasDataset()">Pipelines</b-nav-item>
				<b-nav-item @click="onResults" :active="activeView===RESULTS">Results</b-nav-item>
			</b-nav>
			<!--
			<b-nav is-nav-bar class="ml-auto">
				<b-nav-text class="session-label">Session:</b-nav-text>
				<b-nav-text v-if="sessionID===null" class="session-not-ready">
					<i class="fa fa-close"></i>Unavailable
				</b-nav-text>
				<b-nav-text v-if="sessionID!==null" class="session-ready">
					<i class="fa fa-check"></i>{{sessionID}}
				</b-nav-text>
			</b-nav>
			-->
			<b-nav is-nav-bar class="ml-auto">
				<b-nav is-nav-bar>
					<b-nav-item href="/help">Help</b-nav-item>
				</b-nav>
			</b-nav>
		</b-collapse>
	</b-navbar>
</template>

<script>
import { gotoHome, gotoSearch, gotoExplore, gotoSelect, gotoPipelines, gotoResults } from '../util/nav';

const HOME = Symbol();
const SEARCH = Symbol();
const EXPLORE = Symbol();
const SELECT = Symbol();
const PIPELINES = Symbol();
const RESULTS = Symbol();

const ROUTE_MAPPINGS = {
	'/home': HOME,
	'/search': SEARCH,
	'/explore': EXPLORE,
	'/select': SELECT,
	'/pipelines': PIPELINES,
	'/results': RESULTS
};

export default {
	name: 'nav-bar',

	data() {
		return {
			HOME: HOME,
			SEARCH: SEARCH,
			EXPLORE: EXPLORE,
			SELECT: SELECT,
			PIPELINES: PIPELINES,
			RESULTS: RESULTS,
			activeView: SEARCH
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
		onHome() {
			gotoHome(this.$store, this.$router);
		},
		onSearch() {
			gotoSearch(this.$store, this.$router);
		},
		onExplore() {
			gotoExplore(this.$store, this.$router);
		},
		onSelect() {
			gotoSelect(this.$store, this.$router);
		},
		onPipelines() {
			gotoPipelines(this.$store, this.$router);
		},
		onResults() {
			gotoResults(this.$store, this.$router);
		},
		hasDataset() {
			return !!this.$store.getters.getRouteDataset();
		},
		updateActive() {
			this.activeView = ROUTE_MAPPINGS[this.$route.path];
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
	width: 100%;
	box-shadow: 0px 2px 5px -1px rgba(0,0,0,0.65);
}
</style>
