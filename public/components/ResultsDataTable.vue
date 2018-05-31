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

				<template :slot="targetErrorCol" slot-scope="data">
					<!-- A custom formatted data column cell -->
					<div class="error-bar-container">
						<div class="error-bar" v-bind:style="{ 'background-color': errorBarColor(data.item[targetErrorCol]), width: errorBarWidth(data.item[targetErrorCol]), left: errorBarLeft(data.item[targetErrorCol]) }"></div>
						<div class="error-bar-center"></div>
					</div>
				</template>
			</b-table>
		</div>

	</div>
</template>

<script lang="ts">

import $ from 'jquery';
import _ from 'lodash';
import { spinnerHTML } from '../util/spinner';
import { Extrema } from '../store/dataset/index';
import { TargetRow, TableRow, TableColumn } from '../store/dataset/index';
import { RowSelection } from '../store/highlights/index';
import { getters as resultsGetters } from '../store/results/module';
import { getters as routeGetters } from '../store/route/module';
import { getters as solutionGetters } from '../store/solutions/module';
import { Dictionary } from '../util/dict';
import { removeNonTrainingItems, removeNonTrainingFields, getPredictedCol, getErrorCol } from '../util/data';
import { updateRowSelection, isRowSelected } from '../util/row';
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

		solutionIndex(): number {
			return routeGetters.getSolutionIndex(this.$store);
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		predictedCol(): string {
			return `HEAD_${getPredictedCol(this.target)}`;
		},

		targetErrorCol(): string {
			return getErrorCol(this.target);
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

		items(): TargetRow[] {
			return removeNonTrainingItems(this.dataItems, this.training);
		},

		fields(): Dictionary<TableColumn> {
			return removeNonTrainingFields(this.dataFields, this.training);
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

	updated() {
		if (this.rowSelection) {
			const $rows = $(this.$el).find('table').find('tbody').find('tr');
			$rows.removeClass('selected');
			this.rowSelection.rows.forEach(row => {
				const elem = $rows.get(row.index);
				if (elem) {
					$(elem).addClass('selected');
				}
			});
		}
	},

	methods: {

		onRowClick(row: TableRow) {
			if (!isRowSelected(this.rowSelection, row._key)) {
				// clicked on a different row than last time - new selection
				const r = {
					index: row._key,
					included: true, // TODO: fix this
					cols: _.map(this.fields, (field, key) => {
						return {
							key: key,
							value: row[key]
						};
					})
				};
				updateRowSelection(this, this.instanceName, this.rowSelection, r);
			} else {
				_.remove(this.rowSelection.rows, r => {
					return r.index === row._key;
				});
				updateRowSelection(this, this.instanceName, this.rowSelection, null);
			}
		},

		normalizeError(error: number): number {
			const range = this.residualExtrema.max - this.residualExtrema.min;
			return (error - this.residualExtrema.min) / range;
		},

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

</style>
