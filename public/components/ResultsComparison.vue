<template>
	<div class="results-tables" v-bind:class="{ 'one-table': !hasHighlights, 'two-tables': hasHighlights }">
		<template v-if="hasHighlights">
			<p class="nav-link font-weight-bold">Samples Modeled</p>
			<results-data-table
				refName="topTable"
				instanceName="top-results-data-table"
				:title="topTableTitle"
				:data-fields="highlightedResultDataFields"
				:data-items="highlightedResultDataItems"
				:decorateFunc="topDecorate"
				:showError="regressionEnabled"></results-data-table>
			<results-data-table
				refName="bottomTable"
				instanceName="bottom-results-data-table"
				:title="bottomTableTitle"
				:data-fields="unhighlightedResultDataFields"
				:data-items="unhighlightedResultDataItems"
				:decorateFunc="bottomDecorate"
				:showError="regressionEnabled"></results-data-table>
		</template>
		<template v-if="!hasHighlights">
			<p class="nav-link font-weight-bold">Samples Modeled</p>
			<results-data-table
				refName="singleTable"
				instanceName="single-results-data-table"
				:title="singleTableTitle"
				:data-fields="unhighlightedResultDataFields"
				:data-items="unhighlightedResultDataItems"
				:decorateFunc="bottomDecorate"
				:showError="regressionEnabled"></results-data-table>
		</template>
	</div>
</template>

<script lang="ts">

import ResultsDataTable from '../components/ResultsDataTable.vue';
import { getTask } from '../util/pipelines';
import _ from 'lodash';
import Vue from 'vue';
import { Dictionary } from '../util/dict';
import { getters as dataGetters} from '../store/data/module';
import { getters as routeGetters} from '../store/route/module';
import { getTargetCol, getPredictedCol, getErrorCol } from '../util/data';
import { FilterParams } from '../util/filters';
import { Variable, TargetRow, FieldInfo } from '../store/data/index';
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

		pipelineId(): string {
			return routeGetters.getRoutePipelineId(this.$store);
		},

		filters(): FilterParams {
			return routeGetters.getDecodedFilterParams(this.$store);
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		variables(): Variable[] {
			return dataGetters.getVariables(this.$store);
		},

		hasHighlights(): boolean {
			const highlights = getHighlights(this.$store);
			return highlights && highlights.root && highlights.root.value;
		},

		highlightedResultDataItems(): TargetRow[] {
			return dataGetters.getHighlightedResultDataItems(this.$store);
		},

		highlightedResultDataFields():  Dictionary<FieldInfo> {
			return dataGetters.getHighlightedResultDataFields(this.$store);
		},

		highlightedResultErrors(): number {
			return -1;
		},

		unhighlightedResultDataItems(): TargetRow[] {
			return dataGetters.getUnhighlightedResultDataItems(this.$store);
		},

		unhighlightedResultDataFields():  Dictionary<FieldInfo> {
			return dataGetters.getUnhighlightedResultDataFields(this.$store);
		},

		unhighlightedResultErrors(): number {
			return -1;
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

		topDecorate(): (row: TargetRow) => TargetRow {
			if (this.regressionEnabled) {
				return this.regressionInRangeDecorate;
			}
			return this.classificationMatchDecorate;
		},

		bottomDecorate(): (row: TargetRow) => TargetRow {
			if (this.regressionEnabled) {
				return this.regressionOutOfRangeDecorate;
			}
			return this.classificationNoMatchDecorate;
		},

		numRows(): number {
			return dataGetters.getResultDataNumRows(this.$store);
		},

		topTableTitle(): string {
			return `${this.highlightedResultDataItems.length} <b class="matching-color">matching</b> samples of ${this.numRows}, including ${this.highlightedResultErrors} <b class="erroneous-color">erroneous</b> predictions`;
		},

		bottomTableTitle(): string {
			return `${this.unhighlightedResultDataItems.length} <b class="matching-color">matching</b> samples of ${this.numRows}, including ${this.unhighlightedResultErrors} <b class="erroneous-color">erroneous</b> predictions`;

		},

		singleTableTitle(): string {
			return `Displaying ${this.unhighlightedResultDataItems.length} of ${this.numRows}, including ${this.unhighlightedResultErrors} <b>erroneous</b> predictions`;
		}
	},

	methods: {

		classificationMatchDecorate(row: TargetRow): TargetRow {
			row._cellVariants = {
				[getTargetCol(this.target)]: 'primary',
				[getPredictedCol(this.target)]: 'success'
			};
			return row;
		},

		classificationNoMatchDecorate(row: TargetRow): TargetRow {
			row._cellVariants = {
				[getTargetCol(this.target)]: 'primary',
				[getPredictedCol(this.target)]: 'danger'
			};
			return row;
		},

		regressionInRangeDecorate(row: TargetRow): TargetRow {
			row._cellVariants = {
				[getTargetCol(this.target)]: 'success',
				[getPredictedCol(this.target)]: 'primary',
				[getErrorCol(this.target)]: 'danger'
			};
			return row;
		},

		regressionOutOfRangeDecorate(row: TargetRow): TargetRow {
			row._cellVariants = {
				[getTargetCol(this.target)]: 'success',
				[getPredictedCol(this.target)]: 'primary',
				[getErrorCol(this.target)]: 'danger'
			};
			return row;
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
.erroneous-color {
	color: #e05353;
}
</style>
