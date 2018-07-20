<template>
	<div class="results-tables" v-bind:class="{ 'one-table': !hasHighlights, 'two-tables': hasHighlights }">
		<template v-if="hasHighlights">
			<p class="nav-link font-weight-bold">Samples Modeled</p>
			<results-data-table
				refName="topTable"
				:title="topTableTitle"
				:data-fields="includedResultTableDataFields"
				:data-items="includedResultTableDataItems"
				:showError="regressionEnabled"></results-data-table>
			<br>
			<results-data-table
				refName="bottomTable"
				:title="bottomTableTitle"
				:data-fields="excludedResultTableDataFields"
				:data-items="excludedResultTableDataItems"
				:showError="regressionEnabled"></results-data-table>
		</template>
		<template v-if="!hasHighlights">
			<p class="nav-link font-weight-bold">Samples Modeled</p>
			<results-data-table
				refName="singleTable"
				:title="singleTableTitle"
				:data-fields="includedResultTableDataFields"
				:data-items="includedResultTableDataItems"
				:showError="regressionEnabled"></results-data-table>
		</template>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import ResultsDataTable from '../components/ResultsDataTable.vue';
import { Dictionary } from '../util/dict';
import { getters as datasetGetters } from '../store/dataset/module';
import { getters as resultsGetters } from '../store/results/module';
import { getters as routeGetters } from '../store/route/module';
import { getters as solutionGetters } from '../store/solutions/module';
import { Solution } from '../store/solutions/index';
import { Variable, TableRow, TableColumn } from '../store/dataset/index';
import { getHighlights } from '../util/highlights';

export default Vue.extend({
	name: 'results-comparison',

	components: {
		ResultsDataTable,
	},

	computed: {

		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		solutionId(): string {
			return routeGetters.getRouteSolutionId(this.$store);
		},

		solution(): Solution {
			return solutionGetters.getActiveSolution(this.$store);
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},

		hasHighlights(): boolean {
			const highlights = getHighlights();
			return highlights && highlights.root && highlights.root.value;
		},

		includedResultTableDataItems(): TableRow[] {
			return resultsGetters.getIncludedResultTableDataItems(this.$store);
		},

		includedResultTableDataFields():  Dictionary<TableColumn> {
			return resultsGetters.getIncludedResultTableDataFields(this.$store);
		},

		numIncludedResultItems(): number {
			return this.includedResultTableDataItems ? this.includedResultTableDataItems.length : 0;
		},

		numIncludedResultErrors(): number {
			if (!this.includedResultTableDataItems) {
				return 0;
			}
			return this.includedResultTableDataItems.filter(item => {
				if (this.regressionEnabled) {
					const err = _.toNumber(item[this.solution.errorKey]);
					return err < this.residualThresholdMin || err > this.residualThresholdMax;
				} else {
					return item[this.target] !== item[this.solution.predictedKey];
				}
			}).length;
		},

		excludedResultTableDataItems(): TableRow[] {
			return resultsGetters.getExcludedResultTableDataItems(this.$store);
		},

		excludedResultTableDataFields():  Dictionary<TableColumn> {
			return resultsGetters.getExcludedResultTableDataFields(this.$store);
		},

		numExcludedResultItems(): number {
			return this.excludedResultTableDataItems ? this.excludedResultTableDataItems.length : 0;
		},

		numExcludedResultErrors(): number {
			if (!this.excludedResultTableDataItems) {
				return 0;
			}
			return this.excludedResultTableDataItems.filter(item => {
				if (this.regressionEnabled) {
					const err = _.toNumber(item[this.solution.errorKey]);
					return err < this.residualThresholdMin || err > this.residualThresholdMax;
				} else {
					return item[this.target] !== item[this.solution.predictedKey];
				}
			}).length;
		},

		residualThresholdMin(): number {
			return _.toNumber(routeGetters.getRouteResidualThresholdMin(this.$store));
		},

		residualThresholdMax(): number {
			return _.toNumber(routeGetters.getRouteResidualThresholdMax(this.$store));
		},

		regressionEnabled(): boolean {
			return solutionGetters.isRegression(this.$store);
		},

		numRows(): number {
			return resultsGetters.getResultDataNumRows(this.$store);
		},

		topTableTitle(): string {
			return `${this.numIncludedResultItems} <b class="matching-color">matching</b> samples of ${this.numRows}, including ${this.numIncludedResultErrors} <b class="erroneous-color">erroneous</b> predictions`;
		},

		bottomTableTitle(): string {
			return `${this.numExcludedResultItems} <b class="other-color">other</b> samples of ${this.numRows}, including ${this.numExcludedResultErrors} <b class="erroneous-color">erroneous</b> predictions`;
		},

		singleTableTitle(): string {
			return `Displaying ${this.numExcludedResultItems} of ${this.numRows}, including ${this.numExcludedResultErrors} <b>erroneous</b> predictions`;
		}
	}
});
</script>

<style>
.results-tables,
.one-table,
.two-tables {
	display: flex;
	flex-direction: column;
	flex: none;
}
.results-data-table {
	display: flex;
	flex-direction: column;
}
.two-tables .results-data-table {
	max-height: 50%;
}
.one-table .results-data-table {
	height: 100%;
}
.matching-color {
	color: #00c6e1;
}
.other-color {
	color: #333;
}
.erroneous-color {
	color: #e05353;
}
</style>
