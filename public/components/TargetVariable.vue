<template>
	<div>
		<variable-facets class="target-summary"
			enable-highlighting
			:groups="groups"
			:instance-name="instanceName"></variable-facets>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import VariableFacets from '../components/VariableFacets';
import { getters as routeGetters } from '../store/route/module';
import { Group, createGroups, getNumericalFacetValue, getCategoricalFacetValue, getTimeseriesFacetValue, TOP_RANGE_HIGHLIGHT } from '../util/facets';
import { TARGET_VAR_INSTANCE } from '../store/route/index';
import { Variable, VariableSummary, Highlight } from '../store/dataset/index';
import { updateHighlight } from '../util/highlights';
import { isNumericType } from '../util/types';

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

		targetSummaries(): VariableSummary[] {
			return routeGetters.getTargetVariableSummaries(this.$store);
		},

		groups(): Group[] {
			return createGroups(this.targetSummaries);
		},

		highlight(): Highlight {
			return routeGetters.getDecodedHighlight(this.$store);
		},

		hasFilters(): boolean {
			return routeGetters.getDecodedFilters(this.$store).length > 0;
		},

		instanceName(): string {
			return TARGET_VAR_INSTANCE;
		},

		defaultHighlightType(): string {
			return TOP_RANGE_HIGHLIGHT;
		},

		isTimeseriesAnalysis(): boolean {
			return !!routeGetters.getRouteTimeseriesAnalysis(this.$store);
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
			if (this.highlight || this.hasFilters || this.hasDefaultedAlready) {
				return;
			}

			if (this.targetSummaries.length > 0 && !this.targetSummaries[0].pending) {
				if (isNumericType(this.targetVariable.colType)) {
					this.selectDefaultNumerical();
				} else {
					this.selectDefaultCategorical();
				}
				this.hasDefaultedAlready = true;
			}
		},

		selectDefaultNumerical() {

			if (this.isTimeseriesAnalysis) {

				updateHighlight(this.$router, {
					context: this.instanceName,
					dataset: this.dataset,
					key: this.target,
					value: getTimeseriesFacetValue(this.targetSummaries[0], this.groups[0], this.defaultHighlightType)
				});

			} else {

				updateHighlight(this.$router, {
					context: this.instanceName,
					dataset: this.dataset,
					key: this.target,
					value: getNumericalFacetValue(this.targetSummaries[0], this.groups[0], this.defaultHighlightType)
				});
			}

		},

		selectDefaultCategorical() {
			updateHighlight(this.$router, {
				context: this.instanceName,
				dataset: this.dataset,
				key: this.target,
				value: getCategoricalFacetValue(this.targetSummaries[0])
			});
		}
	}


});
</script>

<style>
.target-summary .variable-facets-container .facets-root-container .facets-group-container .facets-group {
	box-shadow: none;
}

</style>
