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

import _ from 'lodash';
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
			this.defaultTargetHighlight();
		}
	},

	mounted() {
		this.defaultTargetHighlight();
	},

	methods: {
		defaultTargetHighlight() {
			if (this.hasDefaultedHighlight || this.highlights.root) {
				return;
			}

			if (!this.hasDefaultedHighlight && this.targetVariableSummaries.length > 0) {
				updateHighlightRoot(this, {
					context: this.instanceName,
					key: this.target,
					value: this.getValidFacetRangeBounary()
				});
				this.hasDefaultedHighlight = true;
			}
		},
		getValidFacetRangeBounary(): any {
			// facet library is incapable of selecting a range that isnt exactly
			// on a bin boundary, so we need to iterate through and find it
			// manually.
			const summary = this.targetVariableSummaries[0];
			const extrema = summary.extrema;
			const group = this.groups[0];
			const range = extrema.max - extrema.min;
			const from = extrema.min + (range * DEFAULT_HIGHLIGHT_PERCENTILE);
			const to = extrema.max;
			const facet = group.facets[0] as any;
			const slices = facet.histogram.slices;
			let fromSlice = null;
			let toSlice = null;
			for (let i=0; i<slices.length; i++) {
				const slice = _.toNumber(slices[i].label);
				if (from <= slice) {
					fromSlice = slice;
					break;
				}
			}
			for (let i=slices.length-1;  i >= 0; i--) {
				const slice = _.toNumber(slices[i].toLabel);
				if (to >= slice) {
					toSlice = slice;
					break;
				}
			}
			return {
				from: fromSlice,
				to: toSlice
			};
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
