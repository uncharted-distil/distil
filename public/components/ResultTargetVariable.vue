<template>
	<div>
		<facets class="result-summaries-target"
			enable-highlighting
			@facet-click="onCategoricalClick"
			@numerical-click="onNumericalClick"
			@range-change="onRangeChange"
			:row-selection="rowSelection"
			:instanceName="instanceName"
			:groups="targetGroups"
			:highlights="highlights"></facets>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import Facets from '../components/Facets';
import { getters as routeGetters } from '../store/route/module';
import { getters as resultsGetters } from '../store/results/module';
import { Group, createGroups } from '../util/facets';
import { getHighlights, updateHighlightRoot, clearHighlightRoot } from '../util/highlights';
import { VariableSummary } from '../store/dataset/index';
import { Highlight, RowSelection } from '../store/highlights/index';

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
			const varSummaries = resultsGetters.getResultSummaries(this.$store);
			return _.find(varSummaries, v => _.toLower(v.key) === _.toLower(this.target));
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
			return getHighlights(this.$store);
		},
		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		}
	},

	methods: {

		onCategoricalClick(context: string, key: string, value: string) {
			if (key && value) {
				updateHighlightRoot(this, {
					context: context,
					key: this.target,
					value: value
				});
			} else {
				clearHighlightRoot(this);
			}
		},

		onNumericalClick(context: string, key: string, value: { from: number, to: number }) {
			if (!this.highlights.root || this.highlights.root.key !== key) {
				updateHighlightRoot(this, {
					context: this.instanceName,
					key: this.target,
					value: value
				});
			}
		},

		onRangeChange(context: string, key: string, value: { from: { label: string[] }, to: { label: string[] } }) {
			updateHighlightRoot(this, {
				context: this.instanceName,
				key: this.target,
				value: value
			});
			this.$emit('range-change', key, value);
		}
	}

});
</script>

<style>

.result-summaries-target .facets-group {
	box-shadow: none;
}

</style>
