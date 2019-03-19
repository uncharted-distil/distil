<template>
	<div class="results-data-slot">
		<p class="results-data-slot-summary" v-if="hasResults">
			<b-nav tabs>
				<b-nav-item>
					<small v-html="title"></small>
				</b-nav-item>
				<b-form-group class="view-button ml-auto">
					<b-form-radio-group buttons v-model="viewType" button-variant="outline-secondary">
						<b-form-radio :value="IMAGE_VIEW" v-if="isImageDataset" class="view-button">
							<i class="fa fa-image"></i>
						</b-form-radio >
						<b-form-radio :value="TABLE_VIEW" class="view-button">
							<i class="fa fa-columns"></i>
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

		<div class="results-data-slot-container flex-1">
			<div class="results-data-no-results" v-if="isPending">
				<div v-html="spinnerHTML"></div>
			</div>
			<div class="results-data-no-results" v-if="hasNoResults">
				No results available
			</div>

			<template v-if="hasData">

				<results-data-table v-if="viewType===TABLE_VIEW" :data-fields="dataFields" :data-items="dataItems" :instance-name="instanceName"></results-data-table>
				<results-timeseries-view v-if="viewType===TIMESERIES_VIEW" :fields="dataFields" :items="dataItems" :instance-name="instanceName"></results-timeseries-view>
				<!--
				<select-image-mosaic v-if="viewType===IMAGE_VIEW" :included-active="includedActive" :instance-name="instanceName"></select-image-mosaic>
				<select-graph-view v-if="viewType===GRAPH_VIEW" :included-active="includedActive" :instance-name="instanceName"></select-graph-view>
				<select-geo-plot v-if="viewType===GEO_VIEW" :included-active="includedActive" :instance-name="instanceName"></select-geo-plot>
				<select-timeseries-view v-if="viewType===TIMESERIES_VIEW" :included-active="includedActive" :instance-name="instanceName"></select-timeseries-view>
				-->
			</template>
		</div>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import _ from 'lodash';
import ResultsDataTable from './ResultsDataTable';
import ResultsTimeseriesView from './ResultsTimeseriesView';
import { spinnerHTML } from '../util/spinner';
import { TableRow, TableColumn, Variable } from '../store/dataset/index';
import { RowSelection } from '../store/highlights/index';
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
		ResultsTimeseriesView
	},

	data() {
		return {
			viewType: TABLE_VIEW,
			includedActive: true,
			TABLE_VIEW: TABLE_VIEW,
			IMAGE_VIEW: IMAGE_VIEW,
			GRAPH_VIEW: GRAPH_VIEW,
			GEO_VIEW: GEO_VIEW,
			TIMESERIES_VIEW: TIMESERIES_VIEW
		};
	},

	props: {
		title: String as () => string,
		dataItems: Array as () => any[],
		dataFields: Object as () => Dictionary<TableColumn>,
		instanceName: String as () => string
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
		},

		isImageDataset(): boolean {
			return this.variables.filter(v => v.colType === 'image').length  > 0;
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
	display: flex;
	overflow: auto;
	background-color: white;
}

.results-data-no-results {
	width: 100%;
	background-color: #eee;
	padding: 8px;
	text-align: center;
}

</style>
