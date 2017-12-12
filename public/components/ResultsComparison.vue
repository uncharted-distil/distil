<template>
	<div class="results-tables">
		<results-data-table
			class="results-data-table"
			title="Correct Predictions"
			:exclude-non-training="excludeNonTraining"
			:filterFunc="correctFilter"
			:decorateFunc="correctDecorate"
			:showError="regressionEnabled"></results-data-table>
		<results-data-table
			class="results-data-table"
			title="Incorrect Predictions"
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
import { PipelineState, PipelineInfo } from '../store/pipelines/index';
import { Variable, TargetRow } from '../store/data/index';
import { getPipelineResults } from '../util/pipelines';

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
		// if filters change, update data
		// TODO: watch needs to be finer grained
		'$route.query'() {
			this.fetch();
		}
	},

	computed: {
		result(): PipelineInfo {
			const requestId = routeGetters.getRouteCreateRequestId(this.$store);
			const resultId = routeGetters.getRouteResultId(this.$store);
			const pipelineRequest = getPipelineResults(<PipelineState>this.$store.state.pipelineModule, requestId);
			return _.find(pipelineRequest, r => r.pipeline.resultId === resultId);
		},

		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		variables(): Variable[] {
			return dataGetters.getVariables(this.$store);
		},

		residualThreshold(): string {
			return routeGetters.getRouteResidualThreshold(this.$store);
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

		correctFilter(): (dataItem: TargetRow) => boolean {
			if (this.regressionEnabled) {
				return this.regressionInRangeFilter;
			}
			return this.classificationMatchFilter;
		},

		correctDecorate(): (dataItem: TargetRow) => TargetRow {
			if (this.regressionEnabled) {
				return this.regressionInRangeDecorate;
			}
			return this.classificationMatchDecorate;
		},

		incorrectFilter(): (dataItem: TargetRow) => boolean {
			if (this.regressionEnabled) {
				return this.regressionOutOfRangeFilter;
			}
			return this.classificationNoMatchFilter;
		},

		incorrectDecorate(): (dataItem: TargetRow) => TargetRow {
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
				resultId: routeGetters.getRouteResultId(this.$store),
				filters: routeGetters.getDecodedFilters(this.$store)
			});
		},

		// Methods passed to classification result table instances to filter their displays.
		classificationMatchFilter(dataItem: TargetRow): boolean {
			return dataItem[dataItem._target.truth] === dataItem[dataItem._target.predicted];
		},

		classificationNoMatchFilter(dataItem: TargetRow): boolean {
			return dataItem[dataItem._target.truth] !== dataItem[dataItem._target.predicted];
		},

		// Methods passed to classification result table instance to update their row visuals post-filter
		classificationMatchDecorate(dataItem: TargetRow): TargetRow {
			dataItem._cellVariants = {
				[dataItem._target.truth]: 'primary',
				[dataItem._target.predicted]: 'success'
			};
			return dataItem;
		},

		classificationNoMatchDecorate(dataItem: TargetRow): TargetRow {
			dataItem._cellVariants = {
				[dataItem._target.truth]: 'primary',
				[dataItem._target.predicted]: 'danger'
			};
			return dataItem;
		},

		// Methods passed to regression result table instances to filter their displays.

		regressionInRangeFilter(dataItem: TargetRow): boolean {
			// grab the residual threshold slider value and update
			return Math.abs(dataItem[dataItem._target.error]) <= _.toNumber(this.residualThreshold);
		},

		regressionOutOfRangeFilter(dataItem: TargetRow): boolean {
			return Math.abs(dataItem[dataItem._target.error]) > _.toNumber(this.residualThreshold);
		},

		// Methods passed to classification result table instance to update their row visuals post-filter

		regressionInRangeDecorate(dataItem: TargetRow): TargetRow {
			dataItem._cellVariants = {
				[dataItem._target.truth]: 'primary',
				[dataItem._target.predicted]: 'success',
				[dataItem._target.error]: 'success'
			};
			return dataItem;
		},

		// Methods passed to classification result table instance to update their row visuals post-filter

		regressionOutOfRangeDecorate(dataItem: TargetRow): TargetRow {
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
