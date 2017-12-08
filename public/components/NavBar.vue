<template>
	<b-navbar toggleable type="light" variant="faded" class="bottom-shadowed">

		<b-nav-toggle target="nav_collapse"></b-nav-toggle>

		<img src="/images/legendary.svg" class="app-icon"></img>
		<span class="navbar-brand">Distil</span>

		<b-collapse is-nav id="nav_collapse">
			<b-navbar-nav>
				<b-nav-item @click="onHome" :active="activeView===HOME">Home</b-nav-item>
				<b-nav-item @click="onSearch" :active="activeView===SEARCH">Search</b-nav-item>
				<b-nav-item @click="onExplore" :active="activeView===EXPLORE" :disabled="!hasDataset()">Explore</b-nav-item>
				<b-nav-item @click="onSelect" :active="activeView===SELECT" :disabled="!hasDataset()">Select</b-nav-item>
				<b-nav-item @click="onPipelines" :active="activeView===PIPELINES" :disabled="!hasDataset()">Pipelines</b-nav-item>
				<b-nav-item @click="onResults" :active="activeView===RESULTS" :disabled="true">results</b-nav-item>
			</b-navbar-nav>
			<b-navbar-nav class="ml-auto">
				<b-nav-item href="/help">Help</b-nav-item>
				<b-btn v-b-modal.abort size="sm" variant="outline-danger" class="abort-button">Abort</b-btn>
				<b-modal id="abort" title="Abort" @ok="onAbort">
					<div>
						<i class="fa fa-exclamation-triangle fa-3x abort-icon"></i>
						This action will terminate the session.
					</div>
				</b-modal>
			</b-navbar-nav>
		</b-collapse>
	</b-navbar>
</template>

<script>
import '../assets/images/legendary.svg';
import { gotoHome, gotoSearch, gotoExplore, gotoSelect, gotoPipelines, gotoResults } from '../util/nav';
import { getters as routeGetters } from '../store/route/module';
import { getters as appGetters } from '../store/app/module';
import { actions } from '../store/app/module';
import Vue from 'vue';

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

export default Vue.extend({
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

	computed: {
		sessionId() {
			return appGetters.getPipelineSessionID(this.$store);
		}
	},

	mounted() {
		this.updateActive();
		actions.getPipelineSession(this.$store, {
			sessionId: this.sessionId
		});
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
		onAbort() {
			this.$router.replace('/');
			actions.abort(this.$store);
		},
		hasDataset() {
			return !!routeGetters.getRouteDataset(this.$store);
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
});

</script>

<style>
.session-not-ready {
	color: #cf3835 !important;
}
.session-ready {
	color: #00c07f !important;
}
.app-icon {
	height: 36px;
	margin-right: 5px;
}
.app-icon path {
	fill:#cc9900;
}
.abort-icon {
	vertical-align: middle;
	color:#cf3835;
}
.abort-button {
	margin-left: 20px;
}
.session-label {
	padding-right: 4px
}
.bottom-shadowed {
	width: 100%;
	box-shadow: 0px 2px 5px -1px rgba(0,0,0,0.65);
}
</style>
