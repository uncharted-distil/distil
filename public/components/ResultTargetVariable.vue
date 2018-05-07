<template>
	<div>
		<facets class="result-summaries-target"
			@facet-click="onCategoricalClick"
			@numerical-click="onNumericalClick"
			@range-change="onRangeChange"
			:instanceName="instanceName"
			:groups="targetGroups"
			:highlights="highlights"></facets>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import Facets from '../components/Facets';
import { getters as routeGetters } from '../store/route/module';
import { getters as dataGetters } from '../store/data/module';
import { Group, createGroups } from '../util/facets';
import _ from 'lodash';
import { getHighlights, updateHighlightRoot, clearHighlightRoot } from '../util/highlights';
import { isTarget, getVarFromTarget, getTargetCol } from '../util/data';
import { VariableSummary, Highlight } from '../store/data/index';

export default Vue.extend({
	name: 'result-target-variable',

	components: {
		Facets
	},

	data() {
		return {
			instanceName: 'resultTargetVariable'
		};
	},

	computed: {

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},
		targetSummary() : VariableSummary {
			const varSummaries = dataGetters.getResultSummaries(this.$store);
			return _.find(varSummaries, v => _.toLower(v.name) === _.toLower(this.target));
		},
		targetGroups(): Group[] {
			if (this.targetSummary) {
				const target = createGroups([ this.targetSummary ]);
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
			// find var marked as 'target' and set associated values as highlights
			const highlights = _.cloneDeep(getHighlights(this.$store));
			if (_.isEmpty(highlights)) {
				return highlights;
			}
			_.forEach(highlights.values.samples, (values, varName) => {
				if (isTarget(varName)) {
					highlights.values.samples[getVarFromTarget(varName)] = values;
				}
			});
			if (highlights.root && isTarget(highlights.root.key)) {
				highlights.root.key = getVarFromTarget(highlights.root.key);
			}
			return highlights;
		},
	},

	methods: {

		onCategoricalClick(context: string, key: string, value: string) {
			if (key && value) {
				// extract the var name from the key
				const colKey = getTargetCol(this.target);
				updateHighlightRoot(this, {
					context: context,
					key: colKey,
					value: value
				});
			} else {
				clearHighlightRoot(this);
			}
		},

		onNumericalClick(context: string, key: string, value: { from: number, to: number }) {
			if (!this.highlights.root || this.highlights.root.key !== key) {
				const colKey = getTargetCol(this.target);
				updateHighlightRoot(this, {
					context: this.instanceName,
					key: colKey,
					value: value
				});
			}
		},

		onRangeChange(context: string, key: string, value: { from: { label: string[] }, to: { label: string[] } }) {
			const colKey = getTargetCol(this.target);
			updateHighlightRoot(this, {
				context: context,
				key: colKey,
				value: value
			});
			this.$emit('range-change', key, value);
		},

		capitalize(str) {
			return str.toUpperCase();
		}
	}

});
</script>

<style>

.result-summaries-target .facets-group {
	box-shadow: none;
}

/*
.result-summaries-target .facets-facet-horizontal .facet-histogram-bar-highlighted {
	fill: #00C851;
}

.result-summaries-target .facets-facet-horizontal .facet-histogram-bar-highlighted:hover {
	fill: #007E33;
}

.result-summaries-target .facets-facet-horizontal .facet-histogram-bar-highlighted.select-highlight {
	fill: #007bff;
}

.result-summaries-target .facets-facet-vertical .facet-bar-selected {
	box-shadow: inset 0 0 0 1000px #00C851;
}

.result-summaries-target .facets-facet-horizontal .facet-range-filter {
	box-shadow: inset 0 0 0 1000px rgba(0, 225, 11, 0.15);
}
*/

</style>
