<template>
	<div class="results-tables">
		<results-data-table
			class="results-data-table"
			title="Correct Predictions"
			:filterFunc="correctFilter"
			:decorateFunc="correctDecorate"
			:showError="regressionEnabled"
		></results-data-table>
		<results-data-table
			class="results-data-table"
			title="Incorrect Predictions"
			:filterFunc="incorrectFilter"
			:decorateFunc="incorrectDecorate"
			:showError="regressionEnabled"
		></results-data-table>
	</div>
</template>

<script>

import ResultsDataTable from '../components/ResultsDataTable';
import { getTask } from '../util/pipelines';
import _ from 'lodash';

export default {
	name: 'results-comparison',

	components: {
		ResultsDataTable
	},

	mounted() {
		this.fetch();
	},

	watch: {
		// if filters change, update data
		// TODO: watch needs to be finer grained
		'$route.query'() {
			this.fetch();
		}
	},

	computed: {
		result() {
			const requestId = this.$store.getters.getRouteCreateRequestId();
			const resultId = atob(this.$store.getters.getRouteResultId());
			const pipelineRequest = this.$store.getters.getPipelineResults(requestId);
			return _.find(pipelineRequest, r => r.pipeline.resultUri === resultId);
		},

		dataset() {
			return this.$store.getters.getRouteDataset();
		},

		target() {
			return this.$store.getters.getRouteTargetVariable();
		},

		variables() {
			return this.$store.getters.getVariables();
		},

		residualThreshold() {
			return this.$store.getters.getRouteResidualThreshold();
		},

		regressionEnabled() {
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

		correctFilter() {
			if (this.regressionEnabled) {
				return this.regressionInRangeFilter;
			}
			return this.classificationMatchFilter;
		},

		correctDecorate() {
			if (this.regressionEnabled) {
				return this.regressionInRangeDecorate;
			}
			return this.classificationMatchDecorate;
		},

		incorrectFilter() {
			if (this.regressionEnabled) {
				return this.regressionOutOfRangeFilter;
			}
			return this.classificationNoMatchFilter;
		},

		incorrectDecorate() {
			if (this.regressionEnabled) {
				return this.regressionOutOfRangeDecorate;
			}
			return this.classificationNoMatchDecorate;
		}
	},

	methods: {
		fetch() {
			this.$store.dispatch('updateFilteredData', {
				dataset: this.dataset
			}).then(() => {
				this.$store.dispatch('updateResults', {
					dataset: this.dataset,
					resultId: atob(this.$store.getters.getRouteResultId())
				});
			});
		},

		// Methods passed to classification result table instances to filter their displays.
		classificationMatchFilter(dataItem) {
			return dataItem[dataItem._target.truth] === dataItem[dataItem._target.predicted];
		},

		classificationNoMatchFilter(dataItem) {
			return dataItem[dataItem._target.truth] !== dataItem[dataItem._target.predicted];
		},

		// Methods passed to classification result table instance to update their row visuals post-filter
		classificationMatchDecorate(dataItem) {
			dataItem._cellVariants = {
				[dataItem._target.truth]: 'primary',
				[dataItem._target.predicted]: 'success'
			};
			return dataItem;
		},

		classificationNoMatchDecorate(dataItem) {
			dataItem._cellVariants = {
				[dataItem._target.truth]: 'primary',
				[dataItem._target.predicted]: 'danger'
			};
			return dataItem;
		},

		// Methods passed to regression result table instances to filter their displays.

		regressionInRangeFilter(dataItem) {
			// grab the residual threshold slider value and update
			return Math.abs(dataItem[dataItem._target.error]) <= this.residualThreshold;
		},

		regressionOutOfRangeFilter(dataItem) {
			return Math.abs(dataItem[dataItem._target.error]) > this.residualThreshold;
		},

		// Methods passed to classification result table instance to update their row visuals post-filter

		regressionInRangeDecorate(dataItem) {
			dataItem._cellVariants = {
				[dataItem._target.truth]: 'primary',
				[dataItem._target.predicted]: 'success',
				[dataItem._target.error]: 'success'
			};
			return dataItem;
		},

		// Methods passed to classification result table instance to update their row visuals post-filter

		regressionOutOfRangeDecorate(dataItem) {
			dataItem._cellVariants = {
				[dataItem._target.truth]: 'primary',
				[dataItem._target.predicted]: 'warning',
				[dataItem._target.error]: 'warning'
			};
			return dataItem;
		}
	}
};
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
