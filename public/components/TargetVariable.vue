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
import { Variable, VariableSummary } from '../store/dataset/index';
import { getHighlights, updateHighlightRoot } from '../util/highlights';
import { isNumericType } from '../util/types';

import 'font-awesome/css/font-awesome.css';

const DEFAULT_HIGHLIGHT_PERCENTILE = 0.6;

export default Vue.extend({
	name: 'target-variable',

	components: {
		VariableFacets
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		targetVariable(): Variable {
			return routeGetters.getTargetVariable(this.$store);
		},

		targetVariableSummaries(): VariableSummary[] {
			return routeGetters.getTargetVariableSummaries(this.$store);
		},

		groups(): Group[] {
			return createGroups(this.targetVariableSummaries);
		},

		highlights(): Highlight {
			return getHighlights();
		},

		hasFilters(): boolean {
			return routeGetters.getDecodedFilters(this.$store).length > 0;
		},

		instanceName(): string {
			return 'targetVar';
		}
	},

	data() {
		return {
			hasDefaultedAlready: false
		};
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
			// only default higlight numeric types
			if (!this.targetVariable) {
				return;
			}

			// if we have no current highlight, and no filters, highlight default range
			if (this.highlights.root || this.hasFilters || this.hasDefaultedAlready) {
				return;
			}

			if (this.targetVariableSummaries.length > 0 && !this.targetVariableSummaries[0].pending) {
				if (isNumericType(this.targetVariable.type)) {
					this.selectDefaultNumerical();
				} else {
					this.selectDefaultCategorical();
				}
				this.hasDefaultedAlready = true;
			}
		},
		selectDefaultNumerical() {
			updateHighlightRoot(this.$router, {
				context: this.instanceName,
				key: this.target,
				value: this.getNumericalFacetValue()
			});
		},
		selectDefaultCategorical() {
			updateHighlightRoot(this.$router, {
				context: this.instanceName,
				key: this.target,
				value: this.getCategoricalFacetValue()
			});
		},
		getCategoricalFacetValue(): string {
			const summary = this.targetVariableSummaries[0];
			return summary.buckets[0].key;
		},
		getNumericalFacetValue(): any {
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
			// case case set to full range
			let fromSlice = _.toNumber(slices[0].label);
			let toSlice = _.toNumber(slices[slices.length - 1].toLabel);
			// try to narrow into percentile
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
