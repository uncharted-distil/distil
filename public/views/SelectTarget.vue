<template>

	<div class="container-fluid d-flex flex-column h-100 select-view">
		<div class="row flex-0-nav">
		</div>
		<div class="row flex-shrink-0 align-items-center justify-content-center bg-white">
			<div class="col-12 col-md-10">
				<h5 class="header-label">Select Feature to Predict</h5>
			</div>
		</div>
		<div class="row justify-content-center pb-3">
			<div class="col-12 col-md-10 flex-column d-flex h-100">
				<available-target-variables>
				</available-target-variables>
			</div>
		</div>
	</div>

</template>

<script lang="ts">

import Vue from 'vue';
import AvailableTargetVariables from '../components/AvailableTargetVariables.vue';
import { actions as viewActions } from '../store/view/module';
import { getters as routeGetters } from '../store/route/module';

export default Vue.extend({
	name: 'select-view',

	components: {
		AvailableTargetVariables
	},

	computed: {

		availableTargetVarsPage(): number {
			return routeGetters.getRouteAvailableTargetVarsPage(this.$store);
		},
	},

	watch: {
		availableTargetVarsPage() {
			viewActions.fetchSelectTargetData(this.$store);
		}
	},

	beforeMount() {
		viewActions.fetchSelectTargetData(this.$store);
	}
});
</script>

<style>
.select-view .nav-link {
	padding: 1rem 0 0.25rem 0;
	border-bottom: 1px solid #E0E0E0;
	color: rgba(0,0,0,.87);
}
.select-view .nav-tabs .nav-item a {
	padding-left: 0.5rem;
	padding-right: 0.5rem;
}
.select-view .nav-tabs .nav-link {
	color: #757575;
}
.select-view .nav-tabs .nav-link.active {
	color: rgba(0, 0, 0, 0.87);
}
.header-label {
	padding: 1rem 0 0.5rem 0;
	font-weight: bold;
}
</style>
