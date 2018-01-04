<template>
	<div>
		<h6 class="nav-link">Target Feature</h6>
		<div class="target-no-target" v-if="variables.length===0">
			<div class="text-danger">
				<i class="fa fa-times missing-icon"></i><strong>No Target Feature Selected</strong>
			</div>
		</div>
		<variable-facets v-if="variables.length>0"
			type-change
			:variables="variables"
			:dataset="dataset"></variable-facets>
	</div>
</template>

<script lang="ts">

import VariableFacets from '../components/VariableFacets';
import 'font-awesome/css/font-awesome.css';
import Vue from 'vue';
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters} from '../store/route/module';
import { VariableSummary } from '../store/data/index';

export default Vue.extend({
	name: 'target-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		variables(): VariableSummary[] {
			return dataGetters.getTargetVariableSummaries(this.$store);
		}
	}
});
</script>

<style>
.target-no-target {
	width: 100%;
	background-color: #eee;
	padding: 8px;
	font-size: 1rem;
}
.missing-icon {
	padding-right: 4px;
}
</style>
