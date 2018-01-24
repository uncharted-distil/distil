<template>
	<b-navbar toggleable="md" type="dark"  class="fixed-top bottom-shadowed">

		<b-nav-toggle target="nav_collapse"></b-nav-toggle>

		<img src="/images/uncharted.svg" class="app-icon"></img>
		<span class="navbar-brand">Distil</span>

		<b-collapse is-nav id="nav_collapse">
			<b-navbar-nav>
				<b-nav-item @click="onHome" :active="isActive(HOME)" v-bind:class="{ active: isActive(HOME) }">
					<i class="fa fa-home nav-icon"></i>
					<b-nav-text>Home</b-nav-text>
				</b-nav-item>
				<b-nav-item @click="onSearch" :active="isActive(SEARCH)" v-bind:class="{ active: isActive(SEARCH) }">
					<i class="fa fa-angle-right nav-arrow"></i>
					<i class="fa fa-dot-circle-o nav-icon"></i>
					<b-nav-text>Search</b-nav-text>
				</b-nav-item>
				<b-nav-item @click="onSelect" :active="isActive(SELECT)" :disabled="!hasSelectView()" v-bind:class="{ active: isActive(SELECT) }">
					<i class="fa fa-angle-right nav-arrow"></i>
					<i class="fa fa-code-fork nav-icon"></i>
					<b-nav-text>Select</b-nav-text>
				</b-nav-item>
				<b-nav-item @click="onResults" :active="isActive(RESULTS)" :disabled="!hasResultView()" v-bind:class="{ active: isActive(RESULTS) }">
					<i class="fa fa-angle-right nav-arrow"></i>
					<i class="fa fa-line-chart nav-icon"></i>
					<b-nav-text>Results</b-nav-text>
				</b-nav-item>
			</b-navbar-nav>
			<b-navbar-nav class="ml-auto">
				<b-nav-item href="/help">
					<b-nav-text>
					Help
					</b-nav-text>
				</b-nav-item>
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
			RESULTS: RESULTS
		};
	},

	computed: {
		path(): string {
			return routeGetters.getRoutePath(this.$store);
		},
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		activeView(): string {
			return ROUTE_MAPPINGS[this.path] || SEARCH;
		}
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
		hasSelectView(): boolean {
			return !!restoreView(this.$store, '/select', this.dataset);
		},
		hasResultView(): boolean {
			return !!restoreView(this.$store, '/results', this.dataset);
		}
	}
});

</script>

<style>
.navbar {
	background-color: #424242;
}
.navbar {
	background-color: #424242;
}
.nav-arrow {
	color: rgba(255,255,255,1);
	padding-right: 5px;
}
.nav-icon {
	padding: 7px;
	width: 30px;
	height: 30px;
	text-align: center;
	border-radius: 50%;
}
.nav-item .nav-link {
	padding: 2px;
}
.nav-item .navbar-text  {
	letter-spacing: 0.01rem;
}
.navbar-nav .btn  {
	letter-spacing: 0.01rem;
	font-weight: bold;
}
.navbar-nav li a .nav-icon {
	color: white;
	background-color: #616161;
}
.navbar-nav li.active a .nav-icon {
	background-color: #1b1b1b;
}
.navbar-nav li.active a .navbar-text {
	color: rgba(255,255,255,1);
}
.navbar-nav li:hover a .nav-icon {
	transition: 0.5s all ease;
	color: white;
	background-color: #1b1b1b;
}
.navbar-nav li:hover a .navbar-text {
	transition:0.5s all ease;
	color: rgba(255,255,255,1);
}
.navbar-nav li.active ~ li a .nav-icon {
	color: hsla(0,0%,100%,.5);
	background-color: inherit;
}
.navbar-nav li.active ~ li a .navbar-text {
	background-color: inherit;
}
.navbar-nav li.active ~ li a:hover .nav-icon {
	transition:0.5s all ease;
	color: white;
	background-color: #1b1b1b;
}
.navbar-nav li.active ~ li a:hover .navbar-text {
	transition:0.5s all ease;
	color: rgba(255,255,255,1);
}
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
	fill: #c90;
}
.abort-icon {
	vertical-align: middle;
	color: #cf3835;
}
.abort-button {
	margin-left: 20px;
}
.session-label {
	padding-right: 4px;
}
.bottom-shadowed {
	box-shadow: 0 6px 12px 0 rgba(0,0,0,0.10);
}

@media (max-width: 576px) {
	.nav-item .nav-link {
		padding: 5px;
	}
}

</style>
