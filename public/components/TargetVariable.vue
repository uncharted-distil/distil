<template>
	<div>
		<variable-facets class="target-summary"
			enable-highlighting
			:groups="groups"
			:dataset="dataset"
			:instance-name="instanceName"></variable-facets>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import VariableFacets from '../components/VariableFacets';
import { getters as routeGetters} from '../store/route/module';
import { Group, createGroups } from '../util/facets';
import { Highlight } from '../store/highlights/index';
import { VariableSummary } from '../store/dataset/index';
import { getHighlights, updateHighlightRoot } from '../util/highlights';

import 'font-awesome/css/font-awesome.css';

const DEFAULT_HIGHLIGHT_PERCENTILE = 0.8;

export default Vue.extend({
	name: 'target-variable',

	components: {
		VariableFacets
	},

	data() {
		return {
			hasDefaultedHighlight: false
		};
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
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
	},

	watch: {
		targetVariableSummaries() {

			if (this.hasDefaultedHighlight || this.highlights.root) {
				return;
			}

			if (!this.hasDefaultedHighlight && this.targetVariableSummaries.length > 0) {
				const summary = this.targetVariableSummaries[0];
				const extrema = summary.extrema;
				const range = extrema.min + (extrema.max - extrema.min);
				updateHighlightRoot(this, {
					context: this.instanceName,
					key: this.target,
					value: {
						from: extrema.min + (range * DEFAULT_HIGHLIGHT_PERCENTILE),
						to: extrema.max
					}
				});
				this.hasDefaultedHighlight = true;
			}
		}
	}


});
</script>

<style>
.target-summary .variable-facets-container .facets-root-container .facets-group-container .facets-group {
	box-shadow: none;
}

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
