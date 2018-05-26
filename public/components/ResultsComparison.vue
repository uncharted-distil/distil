<template>
	<div class="results-tables" v-bind:class="{ 'one-table': !hasHighlights, 'two-tables': hasHighlights }">
		<template v-if="hasHighlights">
			<p class="nav-link font-weight-bold">Samples Modeled</p>
			<results-data-table
				refName="topTable"
				instanceName="top-results-data-table"
				:title="topTableTitle"
				:data-fields="includedResultTableDataFields"
				:data-items="includedResultTableDataItems"
				:showError="regressionEnabled"></results-data-table>
			<br>
			<results-data-table
				refName="bottomTable"
				instanceName="bottom-results-data-table"
				:title="bottomTableTitle"
				:data-fields="excludedResultTableDataFields"
				:data-items="excludedResultTableDataItems"
				:showError="regressionEnabled"></results-data-table>
		</template>
		<template v-if="!hasHighlights">
			<p class="nav-link font-weight-bold">Samples Modeled</p>
			<results-data-table
				refName="singleTable"
				instanceName="single-results-data-table"
				:title="singleTableTitle"
				:data-fields="includedResultTableDataFields"
				:data-items="includedResultTableDataItems"
				:showError="regressionEnabled"></results-data-table>
		</template>
	</div>
</template>

<script lang="ts">

import ResultsDataTable from '../components/ResultsDataTable.vue';
import { getTask } from '../util/solutions';
import _ from 'lodash';
import Vue from 'vue';
import { Dictionary } from '../util/dict';
import { getters as datasetGetters} from '../store/dataset/module';
import { getters as resultsGetters} from '../store/results/module';
import { getters as routeGetters} from '../store/route/module';
import { getErrorCol } from '../util/data';
import { Variable, TargetRow, TableColumn } from '../store/dataset/index';
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

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},

		hasHighlights(): boolean {
			const highlights = getHighlights(this.$store);
			return highlights && highlights.root && highlights.root.value;
		},

		includedResultTableDataItems(): TargetRow[] {
			return resultsGetters.getIncludedResultTableDataItems(this.$store);
		},

		includedResultTableDataFields():  Dictionary<TableColumn> {
			return resultsGetters.getIncludedResultTableDataFields(this.$store);
		},

		includedResultErrors(): number {
			return this.includedResultTableDataItems.filter(item => {
				const err = _.toNumber(item[getErrorCol(this.target)]);
				return err < this.residualThresholdMin || err > this.residualThresholdMax;
			}).length;
		},

		excludedResultTableDataItems(): TargetRow[] {
			return resultsGetters.getExcludedResultTableDataItems(this.$store);
		},

		excludedResultTableDataFields():  Dictionary<TableColumn> {
			return resultsGetters.getExcludedResultTableDataFields(this.$store);
		},

		excludedResultErrors(): number {
			return this.excludedResultTableDataItems.filter(item => {
				const err = _.toNumber(item[getErrorCol(this.target)]);
				return err < this.residualThresholdMin || err > this.residualThresholdMax;
			}).length;
		},

		residualThresholdMin(): number {
			return _.toNumber(routeGetters.getRouteResidualThresholdMin(this.$store));
		},

		residualThresholdMax(): number {
			return _.toNumber(routeGetters.getRouteResidualThresholdMax(this.$store));
		},

		regressionEnabled(): boolean {
			const targetVarName = this.target;
			const variables = this.variables;
			const targetVar = _.find(variables, v => {
				return _.toLower(v.name) === _.toLower(targetVarName);
			});
			if (_.isEmpty(targetVar)) {
				return false;
			}
			const task = getTask(targetVar.type);
			return task.schemaName === 'regression';
		},

		numRows(): number {
			return resultsGetters.getResultDataNumRows(this.$store);
		},

		topTableTitle(): string {
			return `${this.includedResultTableDataItems.length} <b class="matching-color">matching</b> samples of ${this.numRows}, including ${this.includedResultErrors} <b class="erroneous-color">erroneous</b> predictions`;
		},

		bottomTableTitle(): string {
			return `${this.excludedResultTableDataItems.length} <b class="other-color">other</b> samples of ${this.numRows}, including ${this.excludedResultErrors} <b class="erroneous-color">erroneous</b> predictions`;

		},

		singleTableTitle(): string {
			return `Displaying ${this.excludedResultTableDataItems.length} of ${this.numRows}, including ${this.excludedResultErrors} <b>erroneous</b> predictions`;
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
	min-height: 50%;
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
