<template>
	<b-navbar toggleable="md" type="dark"  class="nav-bar-fixed-top bottom-shadowed">

		<b-nav-toggle target="nav-collapse"></b-nav-toggle>

		<img src="/images/uncharted.svg" class="app-icon">
		<span class="navbar-brand">Distil</span>

		<b-collapse v-if="!isAborted" is-nav id="nav-collapse">
			<b-navbar-nav>
				<b-nav-item @click="onHome" v-if="!isTask1 && !isTask2" :active="isActive(HOME_ROUTE)" v-bind:class="{ active: isActive(HOME_ROUTE) }">
					<i class="fa fa-home nav-icon"></i>
					<b-nav-text>Home</b-nav-text>
				</b-nav-item>
				<b-nav-item @click="onSearch" v-if="!isTask1 && !isTask2" :active="isActive(SEARCH_ROUTE)" v-bind:class="{ active: isActive(SEARCH_ROUTE) }">
					<i class="fa fa-angle-right nav-arrow"></i>
					<i class="fa fa-file-text-o nav-icon"></i>
					<b-nav-text>Select Data</b-nav-text>
				</b-nav-item>
				<b-nav-item @click="onJoinDatasets" v-if="isJoinDatasets" :active="isActive(JOIN_DATASETS_ROUTE)" v-bind:class="{ active: isActive(JOIN_DATASETS_ROUTE) }">
					<i class="fa fa-angle-right nav-arrow"></i>
					<i class="fa fa-database nav-icon"></i>
					<b-nav-text>Join Datasets</b-nav-text>
				</b-nav-item>
				<b-nav-item @click="onSelectTarget" v-if="!isTask2" :active="isActive(SELECT_TARGET_ROUTE)" :disabled="!hasSelectTargetView()" v-bind:class="{ active: isActive(SELECT_TARGET_ROUTE) }">
					<i class="fa fa-angle-right nav-arrow"></i>
					<i class="fa fa-dot-circle-o  nav-icon"></i>
					<b-nav-text>Select Target</b-nav-text>
				</b-nav-item>
				<b-nav-item @click="onSelectData" :active="isActive(SELECT_TRAINING_ROUTE)" :disabled="!hasSelectTrainingView()" v-bind:class="{ active: isActive(SELECT_TRAINING_ROUTE) }">
					<i class="fa fa-angle-right nav-arrow"></i>
					<i class="fa fa-code-fork  nav-icon"></i>
					<b-nav-text>Create Models</b-nav-text>
				</b-nav-item>
				<b-nav-item @click="onResults" :active="isActive(RESULTS_ROUTE)" :disabled="!hasResultView()" v-bind:class="{ active: isActive(RESULTS_ROUTE) }">
					<i class="fa fa-angle-right nav-arrow"></i>
					<i class="fa fa-line-chart nav-icon"></i>
					<b-nav-text>View Models</b-nav-text>
				</b-nav-item>
			</b-navbar-nav>
			<b-navbar-nav class="ml-auto">
				<b-nav-item href="/help">
					<b-nav-text>
					Help
					</b-nav-text>
				</b-nav-item>
				<b-btn v-b-modal.abort size="sm" variant="danger" class="abort-button">Abort</b-btn>
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
import { gotoHome, gotoSearch, gotoJoinDatasets, gotoSelectTarget,
	gotoSelectData, gotoResults } from '../util/nav';
import { actions as appActions,  getters as appGetters } from '../store/app/module';
import { getters as routeGetters } from '../store/route/module';
import { HOME_ROUTE, SEARCH_ROUTE, JOIN_DATASETS_ROUTE, SELECT_TARGET_ROUTE,
	SELECT_TRAINING_ROUTE, RESULTS_ROUTE, ABORT_SUCCESS_ROUTE } from '../store/route/index';
import { restoreView } from '../util/view';
import Vue from 'vue';

export default Vue.extend({
	name: 'nav-bar',

	data() {
		return {
			HOME_ROUTE: HOME_ROUTE,
			SEARCH_ROUTE: SEARCH_ROUTE,
			JOIN_DATASETS_ROUTE: JOIN_DATASETS_ROUTE,
			SELECT_TARGET_ROUTE: SELECT_TARGET_ROUTE,
			SELECT_TRAINING_ROUTE: SELECT_TRAINING_ROUTE,
			RESULTS_ROUTE: RESULTS_ROUTE
		};
	},

	computed: {
		path(): string {
			return routeGetters.getRoutePath(this.$store);
		},
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		joinDatasets(): string[] {
			return routeGetters.getRouteJoinDatasets(this.$store);
		},
		joinDatasetsHash(): string {
			return routeGetters.getRouteJoinDatasetsHash(this.$store);
		},
		isAborted(): boolean {
			return appGetters.isAborted(this.$store);
		},
		isTask1(): boolean {
			return appGetters.isTask1(this.$store);
		},
		isTask2(): boolean {
			return appGetters.isTask2(this.$store);
		},
		isJoinDatasets(): boolean {
			return this.joinDatasets.length === 2 || this.hasJoinDatasetView();
		}
	},

	methods: {
		isActive(view) {
			return view === this.path;
		},
		onHome() {
			gotoHome(this.$router);
		},
		onSearch() {
			gotoSearch(this.$router);
		},
		onJoinDatasets() {
			gotoJoinDatasets(this.$router);
		},
		onSelectTarget() {
			gotoSelectTarget(this.$router);
		},
		onSelectData() {
			gotoSelectData(this.$router);
		},
		onResults() {
			gotoResults(this.$router);
		},
		onAbort() {
			this.$router.replace(ABORT_SUCCESS_ROUTE);
			appActions.abort(this.$store);
		},
		hasJoinDatasetView(): boolean {
			return !!restoreView(JOIN_DATASETS_ROUTE, this.joinDatasetsHash);
		},
		hasSelectTargetView(): boolean {
			return !!restoreView(SELECT_TARGET_ROUTE, this.dataset);
		},
		hasSelectTrainingView(): boolean {
			return !!restoreView(SELECT_TRAINING_ROUTE, this.dataset);
		},
		hasResultView(): boolean {
			return !!restoreView(RESULTS_ROUTE, this.dataset);
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
.nav-bar-fixed-top {
	position: fixed;
	right: 0;
	left: 0;
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
.nav-item {
	white-space: nowrap;
	overflow: hidden;
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
	white-space: nowrap;
	overflow: hidden;
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
