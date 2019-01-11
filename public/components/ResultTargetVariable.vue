<template>
	<div>
		<variable-facets class="result-target-summary"
			enable-highlighting
			:groups="groups"
			:dataset="dataset"
			:instance-name="instanceName"></variable-facets>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import VariableFacets from '../components/VariableFacets.vue';
import { getters as routeGetters } from '../store/route/module';
import { getters as resultsGetters } from '../store/results/module';
import { Group, createGroups, getNumericalFacetValue, getCategoricalFacetValue, TOP_RANGE_HIGHLIGHT } from '../util/facets';
import { getHighlights, updateHighlightRoot, clearHighlightRoot } from '../util/highlights';
import { RESULT_TARGET_VAR_INSTANCE } from '../store/route/index';
import { Variable, VariableSummary } from '../store/dataset/index';
import { Highlight, RowSelection } from '../store/highlights/index';
import { isNumericType } from '../util/types';

export default Vue.extend({
	name: 'result-target-variable',

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

		resultTargetSummary(): VariableSummary {
			return resultsGetters.getTargetSummary(this.$store);
		},

		groups(): Group[] {
			if (this.resultTargetSummary) {
				const target = createGroups([ this.resultTargetSummary ]);
				if (this.highlights.root) {
					const group = target[0];
					if (group.key === this.highlights.root.key) {
						group.facets.forEach(facet => {
							facet.filterable = true;
						});
					}
				}
				return target;
			}
			return [];
		},

		highlights(): Highlight {
			return getHighlights();
		},

		hasFilters(): boolean {
			return routeGetters.getDecodedFilters(this.$store).length > 0;
		},

		instanceName(): string {
			return RESULT_TARGET_VAR_INSTANCE;
		},

		defaultHighlightType(): string {
			return TOP_RANGE_HIGHLIGHT;
		}
	},

	data() {
		return {
			hasDefaultedAlready: false
		};
	},

	watch: {
		targetSummaries() {
			this.defaultTargetHighlight();
		},
		targetVariable() {
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

			if (this.resultTargetSummary && !this.resultTargetSummary.pending) {
				if (isNumericType(this.targetVariable.colType)) {
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
				value: getNumericalFacetValue(this.resultTargetSummary, this.groups[0], this.defaultHighlightType)
			});
		},

		selectDefaultCategorical() {
			updateHighlightRoot(this.$router, {
				context: this.instanceName,
				key: this.target,
				value: getCategoricalFacetValue(this.resultTargetSummary)
			});
		}
	}

});
</script>

<style>

.result-target-summary .variable-facets-container .facets-root-container .facets-group-container .facets-group {
	box-shadow: none;
}

</style>
