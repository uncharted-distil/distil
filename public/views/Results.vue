<template>
	<div class="container-fluid d-flex flex-column h-100 results-view">
		<div class="row flex-0-nav"></div>

		<div class="row align-items-center justify-content-center bg-white">

			<div class="col-12 col-md-6 d-flex flex-column">
				<h5 class="header-label">Select Model That Best Predicts {{target.toUpperCase()}}</h5>

				<div class="row col-12 pl-4">
					<div>
						{{target.toUpperCase()}} is being modeled as a {{targetType}}
					</div>
				</div>
				<div class="row col-12 pl-4">
					<p>
						Use interactive feature highlighting to analyze models. Go back to revise features, if needed.
					</p>
				</div>
			</div>

			<div class="col-12 col-md-6 d-flex flex-column">
				<result-target-variable class="col-12 d-flex flex-column select-target-variables"></result-target-variable>
			</div>
		</div>

		<div class="row flex-12 pb-3">

			<div class='variable-summaries col-12 col-md-3 border-gray-right results-variable-summaries'>
				<p class="nav-link font-weight-bold">Feature Summaries</p>
				<variable-facets
					enable-search
					enable-highlighting
					instance-name="resultTrainingVars"
					:groups="trainingGroups">
				</variable-facets>
			</div>

			<results-comparison
				class="col-12 col-md-6 results-result-comparison"></results-comparison>
			<result-summaries
				class="col-12 col-md-3 border-gray-left results-result-summaries"></result-summaries>
		</div>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import VariableFacets from '../components/VariableFacets.vue';
import ResultsComparison from '../components/ResultsComparison.vue';
import ResultSummaries from '../components/ResultSummaries.vue';
import ResultTargetVariable from '../components/ResultTargetVariable.vue';
import { actions as viewActions } from '../store/view/module';
import { getters as datasetGetters } from '../store/dataset/module';
import { getters as resultGetters } from '../store/results/module';
import { getters as routeGetters } from '../store/route/module';
import { Group, createGroups } from '../util/facets';

export default Vue.extend({
	name: 'results-view',

	components: {
		VariableFacets,
		ResultTargetVariable,
		ResultsComparison,
		ResultSummaries
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},
		targetType(): string {
			const variables = datasetGetters.getVariablesMap(this.$store);
			if (variables && variables[this.target]) {
				return variables[this.target].colType;
			}
			return '';
		},
		trainingGroups(): Group[] {
			const summaries = resultGetters.getTrainingSummaries(this.$store);
			return createGroups(summaries);
		},
		solutionId(): string {
			return routeGetters.getRouteSolutionId(this.$store);
		},
		highlightRootStr(): string {
			return routeGetters.getRouteHighlightRoot(this.$store);
		},
		resultTrainingVarsPage(): number {
			return routeGetters.getRouteResultTrainingVarsPage(this.$store);
		}
	},

	beforeMount() {
		viewActions.fetchResultsData(this.$store);
	},

	watch: {
		highlightRootStr() {
			viewActions.updateResultsHighlights(this.$store);
		},
		solutionId() {
			viewActions.updateResultsSolution(this.$store);
		},
		resultTrainingVarsPage() {
			viewActions.updateResultsHighlights(this.$store);
		}
	}
});
</script>

<style>
.variable-summaries {
	display: flex;
	flex-direction: column;
}
.variable-summaries .facets-group {
	padding-bottom: 20px;
}
.results-view .nav-link {
	padding: 1rem 0 0.25rem 0;
	border-bottom: 1px solid #E0E0E0;
	color: rgba(0,0,0,.87);
}
.header-label {
	padding: 1rem 0 0.5rem 0;
	font-weight: bold;
}
.results-data-table-container {
	background-color: white;
}
.results-view .table td {
	text-align: right;
}
.result-facets {
	margin-bottom: 12px;
}
</style>
