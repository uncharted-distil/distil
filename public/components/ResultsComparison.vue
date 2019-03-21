<template>
	<div class="results-slots" v-bind:class="{ 'one-slot': !hasHighlights, 'two-slots': hasHighlights }">
		<p class="nav-link font-weight-bold">
			<b-nav>
				Samples Modeled
				<b-form-group class="view-button ml-auto">
					<b-form-radio-group buttons v-model="viewType" button-variant="outline-secondary">
						<b-form-radio :value="TABLE_VIEW" class="view-button">
							<i class="fa fa-columns"></i>
						</b-form-radio >
						<b-form-radio :value="IMAGE_VIEW" class="view-button">
							<i class="fa fa-image"></i>
						</b-form-radio >
						<b-form-radio :value="GRAPH_VIEW" class="view-button">
							<i class="fa fa-share-alt"></i>
						</b-form-radio >
						<b-form-radio :value="GEO_VIEW" class="view-button">
							<i class="fa fa-globe"></i>
						</b-form-radio >
						<b-form-radio :value="TIMESERIES_VIEW" class="view-button">
							<i class="fa fa-line-chart"></i>
						</b-form-radio >
					</b-form-radio-group>
				</b-form-group>
			</b-nav>
		</p>
		<template v-if="hasHighlights">
			<results-data-slot
				:title="topSlotTitle"
				:data-fields="includedResultTableDataFields"
				:data-items="includedResultTableDataItems"
				:view-type="viewType"></results-data-slot>
			<br>
			<results-data-slot
				:title="bottomSlotTitle"
				:data-fields="excludedResultTableDataFields"
				:data-items="excludedResultTableDataItems"
				:view-type="viewType"></results-data-slot>
		</template>
		<template v-if="!hasHighlights">
			<results-data-slot
				:title="singleSlotTitle"
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
import { Dictionary } from '../util/dict';
import { getters as datasetGetters } from '../store/dataset/module';
import { getters as resultsGetters } from '../store/results/module';
import { getters as routeGetters } from '../store/route/module';
import { getters as solutionGetters } from '../store/solutions/module';
import { Solution } from '../store/solutions/index';
import { Variable, TableRow, TableColumn } from '../store/dataset/index';
import { getHighlights } from '../util/highlights';

const TABLE_VIEW = 'table';
const IMAGE_VIEW = 'image';
const GRAPH_VIEW = 'graph';
const GEO_VIEW = 'geo';
const TIMESERIES_VIEW = 'timeseries';

export default Vue.extend({
	name: 'results-comparison',

	components: {
		ResultsDataSlot,
	},

	data() {
		return {
			viewType: TABLE_VIEW,
			TABLE_VIEW: TABLE_VIEW,
			IMAGE_VIEW: IMAGE_VIEW,
			GRAPH_VIEW: GRAPH_VIEW,
			GEO_VIEW: GEO_VIEW,
			TIMESERIES_VIEW: TIMESERIES_VIEW
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

		hasHighlights(): boolean {
			const highlights = getHighlights();
			return highlights && highlights.root && highlights.root.value;
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

		topSlotTitle(): string {
			return `${this.numIncludedResultItems} <b class="matching-color">matching</b> samples of ${this.numRows}, including ${this.numIncludedResultErrors} <b class="erroneous-color">erroneous</b> predictions`;
		},

		bottomSlotTitle(): string {
			return `${this.numExcludedResultItems} <b class="other-color">other</b> samples of ${this.numRows}, including ${this.numExcludedResultErrors} <b class="erroneous-color">erroneous</b> predictions`;
		},

		singleSlotTitle(): string {
			return `Displaying ${this.numExcludedResultItems} of ${this.numRows}, including ${this.numExcludedResultErrors} <b>erroneous</b> predictions`;
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
	max-height: 50%;
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
.view-button {
	cursor: pointer;
}
.view-button input[type=radio]{
    display:none;
}
</style>
