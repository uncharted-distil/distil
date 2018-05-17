<template>
	<div class='result-facets'>
		<result-group class="result-group-container" :key="group.name" v-for="(group, index) in resultGroups"
			:name="group.groupName"
			:index="index"
			:timestamp="group.timestamp"
			:request-id="group.requestId"
			:solution-id="group.solutionId"
			:scores="group.scores"
			:predicted-summary="group.predictedSummary"
			:residuals-summary="group.residualsSummary"
			:correctness-summary="group.correctnessSummary"
			:resultHtml="html"
			:residualHtml="html">
		</result-group>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import moment from 'moment';
import Facets from '../components/Facets';
import ResultGroup from '../components/ResultGroup.vue';
import { VariableSummary } from '../store/dataset/index';
import { getters as resultsGetters } from '../store/results/module';
import { getters as routeGetters } from '../store/route/module';
import { getters as solutionGetters } from '../store/solutions/module';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';
import Vue from 'vue';

/*eslint-disable */
interface SummaryGroup {
	requestId: string;
	solutionId: string;
	groupName: string;
	predictedSummary: VariableSummary;
	residualsSummary: VariableSummary;
	correctnessSummary: VariableSummary;
}
/*eslint-enable */

export default Vue.extend({
	name: 'result-facets',

	components: {
		Facets,
		ResultGroup
	},

	props: {
		html: String,
		regression: Boolean
	},

	computed: {

		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		predictedSummaries(): VariableSummary[] {
			return resultsGetters.getPredictedSummaries(this.$store);
		},

		residualSummaries(): VariableSummary[] {
			return this.regression ? resultsGetters.getResidualsSummaries(this.$store) : [];
		},

		correctnessSummaries(): VariableSummary[] {
			return !this.regression ? resultsGetters.getCorrectnessSummaries(this.$store) : [];
		},

		// Generate pairs of residuals and results for each solution in the numerical case.
		resultGroups(): SummaryGroup[] {

			const solutions = solutionGetters.getSolutions(this.$store).filter(solution => solution.feature === this.target);
			const predictedSummaries = this.predictedSummaries;
			const residualsSummaries = this.residualSummaries;
			const correctnessSummaries = this.correctnessSummaries;

			const summaryGroups = solutions.map(solution => {
				const solutionId = solution.solutionId;
				const requestId = solution.requestId;
				const predictedSummary = _.find(predictedSummaries, summary => summary.solutionId === solutionId);
				const residualSummary = _.find(residualsSummaries, summary => summary.solutionId === solutionId);
				const correctnessSummary = _.find(correctnessSummaries, summary => summary.solutionId === solutionId);
				return {
					requestId: requestId,
					solutionId: solutionId,
					groupName: solution ? solution.name : '',
					timestamp: solution ? moment(solution.timestamp).format('YYYY/MM/DD') : '',
					scores: solution ? solution.scores : [],
					predictedSummary: predictedSummary,
					residualsSummary: residualSummary,
					correctnessSummary: correctnessSummary
				};
			});

			// sort alphabetically
			summaryGroups.sort((a, b) => {
				const textA = a.groupName.toLowerCase();
				const textB = b.groupName.toLowerCase();
				return (textA < textB) ? -1 : (textA > textB) ? 1 : 0;
			});

			return summaryGroups;
		}
	}
});
</script>

<style>
button {
	cursor: pointer;
}

.result-group-container {
	overflow-x: hidden;
	overflow-y: hidden;
}
</style>
