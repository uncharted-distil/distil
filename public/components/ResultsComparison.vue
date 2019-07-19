<template>
	<div class="results-slots" v-bind:class="{ 'one-slot': !hasHighlights, 'two-slots': hasHighlights }">

		<view-type-toggle class="flex-shrink-0" v-model="viewTypeModel" :variables="variables">
			Samples Modeled
		</view-type-toggle>

		<div v-if="hasHighlights" class="flex-grow-1">
			<results-data-slot
				instance-name="results-slot-top"
				:title="topSlotTitle"
				:data-fields="includedResultTableDataFields"
				:data-items="includedResultTableDataItems"
				:view-type="viewType"></results-data-slot>
			<results-data-slot
				instance-name="results-slot-bottom"
				:title="bottomSlotTitle"
				:data-fields="excludedResultTableDataFields"
				:data-items="excludedResultTableDataItems"
				:view-type="viewType"></results-data-slot>
		</div>
		<template v-if="!hasHighlights">
			<results-data-slot
				:title="singleSlotTitle"
				instance-name="results-slot"
				:data-fields="includedResultTableDataFields"
				:data-items="includedResultTableDataItems"
				:view-type="viewType"></results-data-slot>
		</template>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import ResultsDataSlot from '../components/ResultsDataSlot';
import ViewTypeToggle from '../components/ViewTypeToggle';
import { Dictionary } from '../util/dict';
import { getters as datasetGetters } from '../store/dataset/module';
import { getters as resultsGetters } from '../store/results/module';
import { getters as routeGetters } from '../store/route/module';
import { getters as solutionGetters } from '../store/solutions/module';
import { Solution } from '../store/solutions/index';
import { Variable, TableRow, TableColumn } from '../store/dataset/index';

const TABLE_VIEW = 'table';
const TIMESERIES_VIEW = 'timeseries';

export default Vue.extend({
	name: 'results-comparison',

	components: {
		ResultsDataSlot,
		ViewTypeToggle
	},

	data() {
		return {
			viewTypeModel: TABLE_VIEW
		};
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

		isTimeseriesAnalysis(): boolean {
			return !!routeGetters.getRouteTimeseriesAnalysis(this.$store);
		},

		viewType(): string {
			if (this.isTimeseriesAnalysis) {
				return TIMESERIES_VIEW;
			}
			return this.viewTypeModel;
		},

		hasHighlights(): boolean {
			const highlight = routeGetters.getDecodedHighlight(this.$store);
			return highlight && highlight.value;
		},

		includedResultTableDataItems(): TableRow[] {
			return resultsGetters.getIncludedResultTableDataItems(this.$store);
		},

		includedResultTableDataFields(): Dictionary<TableColumn> {
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

		excludedResultTableDataFields(): Dictionary<TableColumn> {
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


		isForecasting(): boolean {
			return solutionGetters.isForecasting(this.$store);
		},

		topSlotTitle(): string {
			const matchesLabel = `${this.numIncludedResultItems} <b class="matching-color">matching</b> samples of ${this.numRows}`;
			const erroneousLabel = `, including ${this.numIncludedResultErrors} <b class="erroneous-color">erroneous</b> predictions`;
			return this.isForecasting ? matchesLabel : matchesLabel + erroneousLabel;
		},

		bottomSlotTitle(): string {
			const matchesLabel = `${this.numExcludedResultItems} <b class="other-color">other</b> samples of ${this.numRows}`;
			const erroneousLabel = `, including ${this.numExcludedResultErrors} <b class="erroneous-color">erroneous</b> predictions`;
			return this.isForecasting ? matchesLabel : matchesLabel + erroneousLabel;
		},

		singleSlotTitle(): string {
			const matchesLabel = `Displaying ${this.numExcludedResultItems} of ${this.numRows}`;
			const erroneousLabel = `, including ${this.numExcludedResultErrors} <b>erroneous</b> predictions`;
			return this.isForecasting ? matchesLabel : matchesLabel + erroneousLabel;
		}
	}
});
</script>

<style>
.results-slots {
	display: flex;
	flex-direction: column;
	flex: none;
}
.two-slots .results-data-slot {
	padding-top: 10px;
	height: 50%;
}
.one-slot .results-data-slot {
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
