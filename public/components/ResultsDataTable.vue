<template>
	<div class="results-data-table">
		<p><small v-html="title"></small></p>
		<div class="results-data-table-container">
			<div class="results-data-no-results" v-if="!hasData">
				<div v-html="spinnerHTML"></div>
			</div>
			<div class="results-data-no-results" v-if="hasData && items.length===0">
				No results available
			</div>
			<b-table v-if="items.length>0"
				bordered
				hover
				small
				responsive
				:ref="refName"
				:items="items"
				:fields="fields"
				@row-clicked="onRowClick">

				<template :slot="predictedCol" slot-scope="data">
					<!-- A custom formatted header cell for predicted field -->
					{{target}}<sup>{{solutionIndex}}</sup>
				</template>

				<template :slot="errorCol" slot-scope="data">
					<!-- A custom formatted data column cell -->
					<div class="error-bar-container">
						<div class="error-bar" v-bind:style="{ 'background-color': errorBarColor(data.item[errorCol]), width: errorBarWidth(data.item[errorCol]), left: errorBarLeft(data.item[errorCol]) }"></div>
						<div class="error-bar-center"></div>
					</div>
					{{data.item[errorCol]}}
				</template>
			</b-table>
		</div>

	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import { spinnerHTML } from '../util/spinner';
import { Extrema } from '../store/dataset/index';
import { TableRow, TableColumn, D3M_INDEX_FIELD } from '../store/dataset/index';
import { RowSelection } from '../store/highlights/index';
import { getters as resultsGetters } from '../store/results/module';
import { getters as routeGetters } from '../store/route/module';
import { getters as solutionGetters } from '../store/solutions/module';
import { Solution } from '../store/solutions/index';
import { Dictionary } from '../util/dict';
import { addRowSelection, removeRowSelection, isRowSelected, updateTableRowSelection } from '../util/row';
import Vue from 'vue';

export default Vue.extend({
	name: 'results-data-table',

	props: {
		title: String,
		refName: String,
		dataItems: Array,
		dataFields: Object,
		instanceName: { type: String, default: 'results-table-table' }
	},

	computed: {
		solutionId(): string {
			return routeGetters.getRouteSolutionId(this.$store);
		},

		solution(): Solution {
			return solutionGetters.getActiveSolution(this.$store);
		},

		solutionIndex(): number {
			return routeGetters.getActiveSolutionIndex(this.$store);
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		predictedCol(): string {
			return `HEAD_${this.solution.predictedKey}`;
		},

		errorCol(): string {
			return this.solution.errorKey;
		},

		residualExtrema(): Extrema {
			return resultsGetters.getResidualExtrema(this.$store);
		},

		training(): Dictionary<boolean> {
			return solutionGetters.getActiveSolutionTrainingMap(this.$store);
		},

		hasData(): boolean {
			return !!this.dataItems;
		},

		items(): TableRow[] {
			return updateTableRowSelection(this.dataItems, this.rowSelection, this.instanceName);
		},

		fields(): Dictionary<TableColumn> {
			return this.dataFields;
		},

		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},

		spinnerHTML(): string {
			return spinnerHTML();
		},

		residualThresholdMin(): number {
			return _.toNumber(routeGetters.getRouteResidualThresholdMin(this.$store));
		},

		residualThresholdMax(): number {
			return _.toNumber(routeGetters.getRouteResidualThresholdMax(this.$store));
		},
	},

	methods: {

		onRowClick(row: TableRow) {
			if (!isRowSelected(this.rowSelection, row[D3M_INDEX_FIELD])) {
				addRowSelection(this, this.instanceName, this.rowSelection, row[D3M_INDEX_FIELD]);
			} else {
				removeRowSelection(this, this.instanceName, this.rowSelection, row[D3M_INDEX_FIELD]);
			}
		},

		normalizeError(error: number): number {
			const range = this.residualExtrema.max - this.residualExtrema.min;
			return (error - this.residualExtrema.min) / range;
		},

		// TODO: fix these to work for correctness values too

		errorBarWidth(error: number): string {
			return `${Math.abs((this.normalizeError(error)*50))}%`;
		},

		errorBarLeft(error: number): string {
			const nerr = this.normalizeError(error);
			if (nerr > 0) {
				return '50%';
			}
			return `${50 + nerr * 50}%`;
		},

		errorBarColor(error: number): string {
			if (error < this.residualThresholdMin || error > this.residualThresholdMax) {
				return '#e05353';
			}
			return '#9e9e9e';
		}

	}

});
</script>

<style>

.results-data-table {
	display: flex;
	flex-direction: column;
}
.results-data-table-container {
	display: flex;
	overflow: auto;
}
.results-data-no-results {
	width: 100%;
	background-color: #eee;
	padding: 8px;
	text-align: center;
}
table tr {
	cursor: pointer;
}

.error-bar-container {
	position: relative;
	width: 80px;
	height: 18px;
}

.error-bar {
	position: absolute;
	height: 80%;
	bottom: 0;
}

.error-bar-center {
	position: absolute;
	width: 1px;
	height: 90%;
	left: 50%;
	bottom: 0;
	background-color: #666;
}

.table-selected-row {
	border-left: 4px solid #ff0067;
	background-color: rgba(255, 0, 103, 0.2);
}
</style>
