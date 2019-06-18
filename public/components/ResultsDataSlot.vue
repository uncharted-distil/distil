<template>
	<div class="results-data-slot">
		<p class="results-data-slot-summary" v-if="hasResults">
			<small v-html="title"></small>
		</p>

		<div class="results-data-slot-container" v-bind:class="{ pending: !hasData }">
			<div class="results-data-no-results" v-if="isPending">
				<div v-html="spinnerHTML"></div>
			</div>
			<div class="results-data-no-results" v-if="hasNoResults">
				No results available
			</div>

			<template>
				<results-data-table v-if="viewType===TABLE_VIEW" :data-fields="dataFields" :data-items="dataItems" :instance-name="instanceName"></results-data-table>
				<results-timeseries-view v-if="viewType===TIMESERIES_VIEW" :fields="dataFields" :items="dataItems" :instance-name="instanceName"></results-timeseries-view>
				<results-geo-plot v-if="viewType===GEO_VIEW" :data-fields="dataFields" :data-items="dataItems"  :instance-name="instanceName"></results-geo-plot>
			</template>
		</div>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import _ from 'lodash';
import ResultsDataTable from './ResultsDataTable';
import ResultsTimeseriesView from './ResultsTimeseriesView';
import ResultsGeoPlot from './ResultsGeoPlot';
import { spinnerHTML } from '../util/spinner';
import { TableRow, TableColumn, Variable, RowSelection } from '../store/dataset/index';
import { getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { getters as solutionGetters } from '../store/solutions/module';
import { Solution, SOLUTION_ERRORED } from '../store/solutions/index';
import { Dictionary } from '../util/dict';
import { updateTableRowSelection } from '../util/row';

const TABLE_VIEW = 'table';
const IMAGE_VIEW = 'image';
const GRAPH_VIEW = 'graph';
const GEO_VIEW = 'geo';
const TIMESERIES_VIEW = 'timeseries';

export default Vue.extend({
	name: 'results-data-slot',

	components: {
		ResultsDataTable,
		ResultsTimeseriesView,
		ResultsGeoPlot
	},

	props: {
		title: String as () => string,
		dataItems: Array as () => any[],
		dataFields: Object as () => Dictionary<TableColumn>,
		instanceName: String as () => string,
		viewType: String as () => string
	},

	data() {
		return {
			includedActive: true,
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

		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},

		solution(): Solution {
			return solutionGetters.getActiveSolution(this.$store);
		},

		solutionHasErrored(): boolean {
			return this.solution ? this.solution.progress === SOLUTION_ERRORED : false;
		},

		isPending(): boolean {
			return !this.hasData && !this.solutionHasErrored;
		},

		hasNoResults(): boolean {
			return this.solutionHasErrored || (this.hasData && this.items.length === 0);
		},

		hasResults(): boolean {
			return this.hasData && this.items.length > 0;
		},

		hasData(): boolean {
			return !!this.dataItems;
		},

		items(): TableRow[] {
			return updateTableRowSelection(this.dataItems, this.rowSelection, this.instanceName);
		},

		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},

		spinnerHTML(): string {
			return spinnerHTML();
		}
	}

});
</script>

<style>

.results-data-slot-summary {
	margin: 0;
	flex-shrink: 0;
}

.results-data-slot {
	display: flex;
	flex-direction: column;
}

.results-data-slot-container {
	position: relative;
	display: flex;
	flex-grow: 1;
	overflow: auto;
	background-color: white;
}

.results-data-no-results {
	position: absolute;
	display: block;
	top: 0;
	height: 100%;
	width: 100%;
	padding: 32px;
	text-align: center;
	opacity: 1;
	z-index: 1;
}

.pending {
	opacity: 0.5;
}

</style>
