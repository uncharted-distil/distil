<template>
	<div>
		<p class="nav-link font-weight-bold">Target Feature</p>
		<div class="target-no-target" v-if="groups.length===0">
			<div class="text-danger">
				<i class="fa fa-times missing-icon"></i><strong>No Target Feature Selected</strong>
			</div>
		</div>
		<variable-facets v-if="groups.length>0"
			type-change
			:groups="groups"
			:dataset="dataset"></variable-facets>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import 'font-awesome/css/font-awesome.css';
import VariableFacets from '../components/VariableFacets';
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters} from '../store/route/module';
import { Group, createGroups } from '../util/facets';

export default Vue.extend({
	name: 'target-variables',

	components: {
		VariableFacets
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		groups(): Group[] {
			const summaries = dataGetters.getTargetVariableSummaries(this.$store);
			return createGroups(summaries, false, false);
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
