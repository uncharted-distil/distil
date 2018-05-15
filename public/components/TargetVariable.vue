<template>
	<div>
		<div class="target-no-target" v-if="groups.length===0">
			<div class="text-danger">
				<i class="fa fa-times missing-icon"></i><strong>No Target Feature Selected</strong>
			</div>
		</div>
		<variable-facets v-if="groups.length>0" class="target-summary"
			enable-highlighting
			:groups="groups"
			:dataset="dataset"
			:instance-name="instanceName"></variable-facets>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import 'font-awesome/css/font-awesome.css';
import VariableFacets from '../components/VariableFacets';
import { getters as routeGetters} from '../store/route/module';
import { Group, createGroups } from '../util/facets';
import { Highlight } from '../store/data/index';
import { getHighlights } from '../util/highlights';

export default Vue.extend({
	name: 'target-variable',

	components: {
		VariableFacets
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		targetVariableSummaries(): VariableSummary[] {
			return routeGetters.getTargetVariableSummaries(this.$store);
		},

		groups(): Group[] {
			return createGroups(this.targetVariableSummaries);
		},
		highlights(): Highlight {
			return getHighlights(this.$store);
		},
		instanceName(): string {
			return 'targetVar';
		}
	}

});
</script>

<style>
.target-summary .variable-facets-container .facets-root-container .facets-group-container .facets-group {
	box-shadow: none;
}

/*
.target-summary .facet-range {
	height: 45px;
}
.target-summary .facet-range-controls {
	display: none;
}
*/

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
