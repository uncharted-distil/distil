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

<script lang="ts">

import ResultsDataTable from '../components/ResultsDataTable.vue';
import { getTask } from '../util/pipelines';
import _ from 'lodash';
import Vue from 'vue';
import { getters as dataGetters} from '../store/data/module';
import { getters as routeGetters} from '../store/route/module';
import { actions } from '../store/data/module';
import { PipelineState } from '../store/pipelines/index';
import { getPipelineResults } from '../util/pipelines';

export default Vue.extend({
	name: 'results-comparison',

	components: {
		ResultsDataTable,
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
			const requestId = routeGetters.getRouteCreateRequestId(this.$store);
			const resultId = atob(routeGetters.getRouteResultId(this.$store));
			const pipelineRequest = getPipelineResults(<PipelineState>this.$store.state.pipelineModule, requestId);
			return _.find(pipelineRequest, r => r.pipeline.resultId === resultId);
		},

		dataset() {
			return routeGetters.getRouteDataset(this.$store);
		},

		target() {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		variables() {
			return dataGetters.getVariables(this.$store);
		},

		residualThreshold() {
			return routeGetters.getRouteResidualThreshold(this.$store);
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
			actions.updateResults(this.$store, {
				dataset: this.dataset,
				resultId: atob(this.$store.getters.getRouteResultId()),
				filters: this.$store.getters.getFilters()
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
