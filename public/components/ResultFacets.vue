<template>
	<div class='result-facets'>
		<div class="request-group-container" :key="request.requestId" v-for="(request, index) in requestGroups">

			<p class="nav-link font-weight-bold">
				Solution <sup>{{index}}</sup>

				<div v-if="isPending(request.progress)">
					<b-badge variant="info">{{request.progress}}</b-badge>
					<b-button variant="danger" size="sm" class="pull-right abort-search-button" @click="stopRequest(request.requestId)">Stop</b-button>
				</div>

				<div v-if="isErrored(request.progress)">
					<b-badge variant="danger">
						ERROR
					</b-badge>
				</div>
			</p>

			<result-group class="result-group-container" :key="group.solutionId" v-for="(group, index) in request.groups"
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
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import _ from 'lodash';
import moment from 'moment';
import Facets from '../components/Facets';
import ResultGroup from '../components/ResultGroup.vue';
import { VariableSummary } from '../store/dataset/index';
import { REQUEST_COMPLETED, REQUEST_ERRORED } from '../store/solutions/index';
import { getters as resultsGetters } from '../store/results/module';
import { getters as routeGetters } from '../store/route/module';
import { getters as solutionGetters, actions as solutionActions } from '../store/solutions/module';

import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';

/*eslint-disable */
interface SummaryGroup {
	requestId: string;
	solutionId: string;
	groupName: string;
	predictedSummary: VariableSummary;
	residualsSummary: VariableSummary;
	correctnessSummary: VariableSummary;
}

interface RequestGroup {
	requestId: string;
	groups: SummaryGroup[];
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

		requestGroups(): RequestGroup[] {
			const requests = solutionGetters.getSolutionsRequests(this.$store);
			const filtered = requests.filter(request => {
				return request.dataset === this.dataset && request.feature === this.target;
			});
			const predictedSummaries = this.predictedSummaries;
			const residualsSummaries = this.residualSummaries;
			const correctnessSummaries = this.correctnessSummaries;
			return filtered.map(request => {
				return {
					requestId: request.requestId,
					progress: request.progress,
					groups: request.solutions.map(solution => {
						const solutionId = solution.solutionId;
						const requestId = solution.requestId;
						const predictedSummary = _.find(predictedSummaries, summary => summary.solutionId === solutionId);
						const residualSummary = _.find(residualsSummaries, summary => summary.solutionId === solutionId);
						const correctnessSummary = _.find(correctnessSummaries, summary => summary.solutionId === solutionId);
						return {
							requestId: requestId,
							solutionId: solutionId,
							groupName: solution.feature,
							timestamp: moment(solution.timestamp).format('YYYY/MM/DD'),
							scores: solution.scores,
							predictedSummary: predictedSummary,
							residualsSummary: residualSummary,
							correctnessSummary: correctnessSummary
						};
					})
				};
			});
		},
	},

	methods: {

		isPending(status: string): boolean {
			return status !== REQUEST_COMPLETED && status !== REQUEST_ERRORED;
		},

		isCompleted(status: string): boolean {
			return status === REQUEST_COMPLETED;
		},

		isErrored(status: string): boolean {
			return status === REQUEST_ERRORED;
		},

		stopRequest(requestId: string) {
			solutionActions.stopSolutionRequest(this.$store, { requestId: requestId });
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
