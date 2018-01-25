<template>

	<div class="container-fluid d-flex flex-column h-100 select-view">
		<div class="row flex-0-nav">
		</div>
		<div class="row flex-1 align-items-center justify-content-center bg-white">
			<div class="col-12 col-md-10">
				<h5 class="header-label">Select Feature to Predict</h5>
			</div>
		</div>
		<div class="row flex-10 justify-content-center pb-3">
			<div class="col-12 col-md-10 d-flex">
				<available-target-variables>
				</available-target-variables>
			</div>
		</div>
	</div>

</template>

<script lang="ts">

import AvailableTargetVariables from '../components/AvailableTargetVariables.vue';
import { getters as dataGetters, actions } from '../store/data/module';
import { getters as routeGetters} from '../store/route/module';
import { Variable } from '../store/data/index';
import Vue from 'vue';

export default Vue.extend({
	name: 'select-view',

	components: {
		AvailableTargetVariables
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		variables(): Variable[] {
			return dataGetters.getVariables(this.$store);
		},
		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		}
	},

	mounted() {
		this.fetch();
	},

	methods: {
		fetch() {
			actions.fetchVariablesAndVariableSummaries(this.$store, {
				dataset: this.dataset
			});
		}
	}
});
</script>

<style>
.select-view .nav-link {
	padding: 1rem 0 0.25rem 0;
	border-bottom: 1px solid #E0E0E0;
}
.header-label {
	padding: 1rem 0 0.5rem 0;
	font-weight: bold;
}

</style>
