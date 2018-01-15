<template>
	<b-navbar toggleable type="light" variant="faded" class="bottom-shadowed">

		<b-nav-toggle target="nav_collapse"></b-nav-toggle>

		<img src="/images/uncharted.svg" class="app-icon"></img>
		<span class="navbar-brand">Distil</span>

		<b-collapse is-nav id="nav_collapse">
			<b-navbar-nav>
				<b-nav-item @click="onHome" :active="isActive(HOME)">
					<i class="fa fa-angle-right nav-selection" v-bind:class="{ active: isActive(HOME) }"></i>
					<i class="fa fa-home nav-icon" v-bind:class="{ active: isActive(HOME) }"></i>
					Home
				</b-nav-item>
				<b-nav-item @click="onSearch" :active="isActive(SEARCH)">
					<i class="fa fa-angle-right nav-selection" v-bind:class="{ active: isActive(SEARCH) }"></i>
					<i class="fa fa-dot-circle-o nav-icon" v-bind:class="{ active: isActive(SEARCH) }"></i>
					Search
				</b-nav-item>
				<b-nav-item @click="onSelect" :active="isActive(SELECT)" :disabled="!hasSelectView()">
					<i class="fa fa-angle-right nav-selection" v-bind:class="{ active: isActive(SELECT) }"></i>
					<i class="fa fa-code-fork nav-icon" v-bind:class="{ active: isActive(SELECT) }"></i>
					Select
				</b-nav-item>
				<b-nav-item @click="onResults" :active="isActive(RESULTS)" :disabled="!hasResultView()">
					<i class="fa fa-angle-right nav-selection" v-bind:class="{ active: isActive(RESULTS) }"></i>
					<i class="fa fa-line-chart nav-icon" v-bind:class="{ active: isActive(RESULTS) }"></i>
					Results
				</b-nav-item>
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

<script lang="ts">
import '../assets/images/uncharted.svg';
import { gotoHome, gotoSearch, gotoSelect, gotoResults } from '../util/nav';
import { actions as appActions } from '../store/app/module';
import { getters as routeGetters } from '../store/route/module';
import { getters as pipelineGetters, actions as pipelineActions } from '../store/pipelines/module';
import { restoreView } from '../util/view';
import Vue from 'vue';

const HOME = Symbol();
const SEARCH = Symbol();
const SELECT = Symbol();
const RESULTS = Symbol();

const ROUTE_MAPPINGS = {
	'/home': HOME,
	'/search': SEARCH,
	'/select': SELECT,
	'/results': RESULTS
};

export default Vue.extend({
	name: 'nav-bar',

	data() {
		return {
			HOME: HOME,
			SEARCH: SEARCH,
			SELECT: SELECT,
			RESULTS: RESULTS,
			activeView: SEARCH
		};
	},

	computed: {
		sessionId(): string {
			return pipelineGetters.getPipelineSessionID(this.$store);
		},
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		}
	},

	mounted() {
		this.updateActive();
		pipelineActions.startPipelineSession(this.$store, {
			sessionId: this.sessionId
		});
	},

	methods: {
		isActive(view) {
			return view === this.activeView;
		},
		onHome() {
			gotoHome(this.$store, this.$router);
		},
		onSearch() {
			gotoSearch(this.$store, this.$router);
		},
		onSelect() {
			gotoSelect(this.$store, this.$router);
		},
		onResults() {
			gotoResults(this.$store, this.$router);
		},
		onAbort() {
			this.$router.replace('/');
			appActions.abort(this.$store);
		},
		updateActive() {
			this.activeView = ROUTE_MAPPINGS[this.$route.path];
		},
		hasSelectView(): boolean {
			return !!restoreView(this.$store, '/select', this.dataset);
		},
		hasResultView(): boolean {
			return !!restoreView(this.$store, '/results', this.dataset);
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
.nav-selection {
	visibility: hidden;
}
.nav-selection.active {
	visibility: visible;
}
.nav-icon {
	width: 32px;
	height: 32px;
	text-align: center;
	border-radius: 50%;
}
.nav-icon.active {
	color: white;
	background-color: black;
}
</style>
