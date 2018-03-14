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
			:dataset="dataset"
			:instance-name="instanceName"
			@facet-click="onCategoricalClick"
			@numerical-click="onNumericalClick"></variable-facets>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import 'font-awesome/css/font-awesome.css';
import VariableFacets from '../components/VariableFacets';
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters} from '../store/route/module';
import { Group, createGroups } from '../util/facets';
import { Highlight } from '../store/data/index';
import { getHighlights, updateHighlightRoot, clearHighlightRoot } from '../util/highlights';

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
			const groups =  createGroups(summaries, false, false);
			if (this.highlights.root) {
				groups.forEach(group => {
					if (group) {
						if (group.key === this.highlights.root.key) {
							group.facets.forEach(facet => {
								facet.filterable = true;
							});
						}
					}
				});
			}
			return groups;
		},
		highlights(): Highlight {
			return getHighlights(this.$store);
		},
		instanceName(): string {
			return 'targetVar';
		}
	},

	methods: {

		onCategoricalClick(context: string, key: string, value: string) {
			if (key && value) {
				// extract the var name from the key
				updateHighlightRoot(this, {
					context: context,
					key: key,
					value: value
				});
			} else {
				clearHighlightRoot(this);
			}
		},

		onNumericalClick(key: string) {
			console.log(key);
			if (!this.highlights.root || this.highlights.root.key !== key) {
				console.log('derp');
				updateHighlightRoot(this, {
					context: this.instanceName,
					key: key,
					value: null
				});
			}
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
