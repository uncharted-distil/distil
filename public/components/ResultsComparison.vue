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
		ResultsDataTable,
	},

	watch: {
		// if filters change, update data
		'$route.query'() {
			const dataset = this.$store.getters.getRouteDataset();
			this.$store.dispatch('updateFilteredData', dataset).then(() => {
				this.$store.dispatch('updateResults', {
					dataset: dataset,
					resultId: atob(this.$store.getters.getRouteResultId()),
					generateResiduals: this.regressionEnabled
				});
			});
		}
	},

	data() {
		return {
			correctFilter: this.classificationMatchFilter,
			correctDecorate: this.classificationMatchDecorate,
			incorrectFilter: this.classificationNoMatchFilter,
			incorrectDecorate: this.classificationNoMatchDecorate
		};
	},

	mounted() {
		// get dataset from route - generate residuals if we're dealing with regression
		const dataset = this.$store.getters.getRouteDataset();
		this.$store.dispatch('updateFilteredData', dataset).then(() => {
			this.$store.dispatch('updateResults', {
				dataset: dataset,
				resultId: atob(this.$store.getters.getRouteResultId()),
				// generateResiduals: isRegressionOutput(this.result.pipeline.output)
				generateResiduals: this.regressionEnabled
			});
		});

		// set the filter and decorate functions based on the result type
		if (this.regressionEnabled) {
			this.correctFilter = this.regressionInRangeFilter;
			this.correctDecorate = this.regressionInRangeDecorate;
			this.incorrectFilter = this.regressionOutOfRangeFilter;
			this.incorrectDecorate = this.regressionOutOfRangeDecorate;
		} else {
			this.correctFilter = this.classificationMatchFilter;
			this.correctDecorate = this.classificationMatchDecorate;
			this.incorrectFilter = this.classificationNoMatchFilter;
			this.incorrectDecorate = this.classificationNoMatchDecorate;
		}
	},

	computed: {
		result() {
			const requestId = this.$store.getters.getRouteCreateRequestId();
			const resultId = atob(this.$store.getters.getRouteResultId());
			const pipelineRequest = this.$store.getters.getPipelineResults(requestId);
			return _.find(pipelineRequest, r => r.pipeline.resultUri === resultId);
		},

		regressionEnabled() {
			//return isRegressionOutput(this.result.pipeline.output);
			const targetVarName = this.$store.getters.getRouteTargetVariable();
			const targetVar = _.find(this.$store.getters.getVariables(), v => v.name === targetVarName);
			const task = getTask(targetVar.type);
			return task.schemaName === 'regression';
		}
	},

	methods: {
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
				[dataItem._target.truth]: 'info',
				[dataItem._target.predicted]: 'success'
			};
			return dataItem;
		},

		classificationNoMatchDecorate(dataItem) {
			dataItem._cellVariants = {
				[dataItem._target.truth]: 'info',
				[dataItem._target.predicted]: 'danger'
			};
			return dataItem;
		},

		// Methods passed to regression result table instances to filter their displays.

		regressionInRangeFilter(dataItem) {
			// grab the residual threshold slider value and update
			const residualThreshold = this.$store.getters.getRouteResidualThreshold();
			return Math.abs(dataItem[dataItem._target.error]) <= residualThreshold;
		},

		regressionOutOfRangeFilter(dataItem) {
			const residualThreshold = this.$store.getters.getRouteResidualThreshold();
			return Math.abs(dataItem[dataItem._target.error]) > residualThreshold;
		},

		// Methods passed to classification result table instance to update their row visuals post-filter

		regressionInRangeDecorate(dataItem) {
			dataItem._cellVariants = {
				[dataItem._target.truth]: 'info',
				[dataItem._target.predicted]: 'success',
				[dataItem._target.error]: 'success'
			};
			return dataItem;
		},

		// Methods passed to classification result table instance to update their row visuals post-filter

		regressionOutOfRangeDecorate(dataItem) {
			dataItem._cellVariants = {
				[dataItem._target.truth]: 'info',
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
