<template>
	<div class="results-tables">
		<results-data-table
			class="results-data-table"
			title="Correct Predictions"
			refName="correctTable"
			instanceName="correct-results-data-table"
			:exclude-non-training="excludeNonTraining"
			:filterFunc="correctFilter"
			:decorateFunc="correctDecorate"
			:showError="regressionEnabled"></results-data-table>
		<results-data-table
			class="results-data-table"
			title="Incorrect Predictions"
			refName="incorrectTable"
			instanceName="incorrect-results-data-table"
			:exclude-non-training="excludeNonTraining"
			:filterFunc="incorrectFilter"
			:decorateFunc="incorrectDecorate"
			:showError="regressionEnabled"></results-data-table>
	</div>
</template>

<script lang="ts">

import ResultsDataTable from '../components/ResultsDataTable.vue';
import { getTask } from '../util/pipelines';
import _ from 'lodash';
import Vue from 'vue';
import { getters as dataGetters} from '../store/data/module';
import { getters as routeGetters} from '../store/route/module';
import { actions } from '../store/data/module';
import { getTargetCol, getPredictedCol, getErrorCol } from '../util/data';
import { Variable, TargetRow } from '../store/data/index';

export default Vue.extend({
	name: 'results-comparison',

	components: {
		ResultsDataTable,
	},

	props: {
		excludeNonTraining: Boolean
	},

	mounted() {
		this.fetch();
	},

	watch: {
		// if pipeline id changes, update data
		'$route.query.pipelineId'() {
			this.fetch();
		},
		// if filters change, update data
		'$route.query.filters'() {
			this.fetch();
		}
	},

	computed: {

		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		variables(): Variable[] {
			return dataGetters.getVariables(this.$store);
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

		correctFilter(): (row: TargetRow) => boolean {
			if (this.regressionEnabled) {
				return this.regressionInRangeFilter;
			}
			return this.classificationMatchFilter;
		},

		correctDecorate(): (row: TargetRow) => TargetRow {
			if (this.regressionEnabled) {
				return this.regressionInRangeDecorate;
			}
			return this.classificationMatchDecorate;
		},

		incorrectFilter(): (row: TargetRow) => boolean {
			if (this.regressionEnabled) {
				return this.regressionOutOfRangeFilter;
			}
			return this.classificationNoMatchFilter;
		},

		incorrectDecorate(): (row: TargetRow) => TargetRow {
			if (this.regressionEnabled) {
				return this.regressionOutOfRangeDecorate;
			}
			return this.classificationNoMatchDecorate;
		}
	},

	methods: {
		fetch() {
			actions.updateResults(this.$store, {
				dataset: this.dataset,
				pipelineId: routeGetters.getRoutePipelineId(this.$store),
				filters: routeGetters.getDecodedFilters(this.$store)
			});
		},

		// Methods passed to classification result table instances to filter their displays.
		classificationMatchFilter(row: TargetRow): boolean {
			return row[getTargetCol(this.target)] === row[getPredictedCol(this.target)];
		},

		classificationNoMatchFilter(row: TargetRow): boolean {
			return row[getTargetCol(this.target)] !== row[getPredictedCol(this.target)];
		},

		// Methods passed to classification result table instance to update their row visuals post-filter
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

		// Methods passed to regression result table instances to filter their displays.

		regressionInRangeFilter(row: TargetRow): boolean {
			// grab the residual threshold slider value and update
			const err = row[getErrorCol(this.target)];
			return err >= this.residualThresholdMin && err <= this.residualThresholdMax;
		},

		regressionOutOfRangeFilter(row: TargetRow): boolean {
			return !this.regressionInRangeFilter(row);
		},

		// Methods passed to classification result table instance to update their row visuals post-filter

		regressionInRangeDecorate(row: TargetRow): TargetRow {
			row._cellVariants = {
				[getTargetCol(this.target)]: 'primary',
				[getPredictedCol(this.target)]: 'success',
				[getErrorCol(this.target)]: 'success'
			};
			return row;
		},

		// Methods passed to classification result table instance to update their row visuals post-filter

		regressionOutOfRangeDecorate(row: TargetRow): TargetRow {
			row._cellVariants = {
				[getTargetCol(this.target)]: 'primary',
				[getPredictedCol(this.target)]: 'warning',
				[getErrorCol(this.target)]: 'warning'
			};
			return row;
		}
	}
});
</script>

<style>
.results-tables {
	display: flex;
	flex-direction: column;
	flex: none;
}
.results-data-table {
	display: flex;
	flex-direction: column;
	max-height: 50%;
	min-height: 50%;
}
</style>
