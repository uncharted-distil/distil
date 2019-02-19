<template>
	<div class="results-data-table">
		<p class="result-data-table-summary" v-if="hasResults"><small v-html="title"></small></p>
		<div class="results-data-table-container flex-1">
			<div class="results-data-no-results" v-if="isPending">
				<div v-html="spinnerHTML"></div>
			</div>
			<div class="results-data-no-results" v-if="hasNoResults">
				No results available
			</div>
			<fixed-header-table v-if="hasResults">
				<b-table
					bordered
					hover
					small
					:ref="refName"
					:items="items"
					:fields="fields"
					:sort-by="errorCol"
					:sort-compare="sortingByResidualError ? sortingByErrorFunction : undefined"
					@row-clicked="onRowClick"
					@sort-changed="onSortChanged">

					<template :slot="predictedCol" slot-scope="data">
						{{target}}<sup>{{solutionIndex}}</sup>
					</template>

					<template v-for="imageField in imageFields" :slot="imageField" slot-scope="data">
						<image-preview :key="imageField" :image-url="data.item[imageField]"></image-preview>
					</template>

					<template v-for="timeseriesField in timeseriesFields" :slot="timeseriesField" slot-scope="data">
						<sparkline-preview :key="timeseriesField" :timeseries-url="data.item[timeseriesField]" :timeseries-col-name="timeseriesField"></sparkline-preview>
					</template>

					<template :slot="errorCol" slot-scope="data">
						<!-- residual error -->
						<div class="error-bar-container" v-if="isTargetNumerical">
							<div class="error-bar" v-bind:style="{ 'background-color': errorBarColor(data.item[errorCol]), width: errorBarWidth(data.item[errorCol]), left: errorBarLeft(data.item[errorCol]) }"></div>
							<div class="error-bar-center"></div>
						</div>

						<!-- correctness error -->
						<div v-if="isTargetCategorical">
							<div v-if="data.item[predictedCol]==data.item[this.target]">
								Correct
							</div>
							<div v-if="data.item[predictedCol]!=data.item[this.target]">
								Incorrect
							</div>
						</div>
					</template>
				</b-table>
			</fixed-header-table>
		</div>

	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import FixedHeaderTable from './FixedHeaderTable'
import SparklinePreview from './SparklinePreview';
import ImagePreview from './ImagePreview';
import { spinnerHTML } from '../util/spinner';
import { Extrema, TableRow, TableColumn, D3M_INDEX_FIELD } from '../store/dataset/index';
import { RowSelection } from '../store/highlights/index';
import { getters as resultsGetters } from '../store/results/module';
import { getters as routeGetters } from '../store/route/module';
import { getters as solutionGetters } from '../store/solutions/module';
import { Solution, SOLUTION_ERRORED } from '../store/solutions/index';
import { Dictionary } from '../util/dict';
import { getVarType, isTextType, IMAGE_TYPE, TIMESERIES_TYPE } from '../util/types';
import { addRowSelection, removeRowSelection, isRowSelected, updateTableRowSelection } from '../util/row';
import Vue from 'vue';

export default Vue.extend({
	name: 'results-data-table',

	components: {
		ImagePreview,
		SparklinePreview,
		FixedHeaderTable,
	},

	data() {
		return {
			sortingBy: undefined
		};
	},

	props: {
		title: String as () => string,
		refName: String as () => string,
		dataItems: Array as () => any[],
		dataFields: Object as () => Dictionary<TableColumn>,
		instanceName: {
			type: String as () => string,
			default: 'results-table-table'
		}
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

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		isTargetCategorical(): boolean {
			return isTextType(getVarType(this.target));
		},

		isTargetNumerical(): boolean {
			return !this.isTargetCategorical;
		},

		predictedCol(): string {
			return `HEAD_${this.solution.predictedKey}`;
		},

		errorCol(): string {
			return this.solution.errorKey;
		},

		residualExtrema(): Extrema {
			return resultsGetters.getResidualsExtrema(this.$store);
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

		imageFields(): string[] {
			return _.map(this.fields, (field, key) => {
				return {
					key: key,
					type: field.type
				};
			})
			.filter(field => field.type === IMAGE_TYPE)
			.map(field => field.key);
		},

		timeseriesFields(): string[] {
			return _.map(this.fields, (field, key) => {
				return {
					key: key,
					type: field.type
				};
			})
			.filter(field => field.type === TIMESERIES_TYPE)
			.map(field => field.key);
		},

		isRegression(): boolean {
			return solutionGetters.isRegression(this.$store);
		},

		sortingByResidualError(): boolean {
			if (this.isRegression && (this.sortingBy === this.errorCol || this.sortingBy === undefined)) {
				return true;
			}
			return false;
		}
	},

	methods: {

		onRowClick(row: TableRow) {
			if (!isRowSelected(this.rowSelection, row[D3M_INDEX_FIELD])) {
				addRowSelection(this.$router, this.instanceName, this.rowSelection, row[D3M_INDEX_FIELD]);
			} else {
				removeRowSelection(this.$router, this.instanceName, this.rowSelection, row[D3M_INDEX_FIELD]);
			}
		},

		normalizeError(error: number): number {
			const range = this.residualExtrema.max - this.residualExtrema.min;
			return ((error - this.residualExtrema.min) / range) * 2 - 1;
		},

		// TODO: fix these to work for correctness values too

		errorBarWidth(error: number): string {
			return `${Math.abs((this.normalizeError(error) * 50))}%`;
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
		},

		sortingByErrorFunction(a, b, key): number {
			return Math.abs(_.toNumber(a[key])) - Math.abs(_.toNumber(b[key]));
		},

		onSortChanged(event) {
			this.sortingBy = event.sortBy;
		}
	}

});
</script>

<style>

.result-data-table-summary {
	margin: 0;
	flex-shrink: 0;
}
.results-data-table {
	display: flex;
	flex-direction: column;
}
.results-data-table-container {
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
