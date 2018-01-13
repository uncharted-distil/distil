<template>
	<div class='result-facets'>
		<result-group class="result-group-container" :key="group.name" v-for="group in resultGroups"
			:name="group.groupName"
			:request-id="group.requestId"
			:pipeline-id="group.pipelineId"
			:result-summary="group.resultSummary"
			:residuals-summary="group.residualsSummary"
			:resultHtml="html"
			:residualHtml="html">
		</result-group>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Facets from '../components/Facets';
import ResultGroup from '../components/ResultGroup.vue';
import { VariableSummary } from '../store/data/index';
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { getPipelinesForDatasetAndTarget } from '../util/pipelines';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';
import Vue from 'vue';

/*eslint-disable */
interface SummaryGroup {
	requestId: string;
	pipelineId: string;
	groupName: string;
	resultSummary: VariableSummary;
	residualsSummary: VariableSummary;
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

		// Generate pairs of residuals and results for each pipeline in the numerical case.
		resultGroups(): SummaryGroup[] {

			const pipelines = getPipelinesForDatasetAndTarget(this.$store.state.pipelineModule, this.dataset, this.target);
			const resultSummaries = dataGetters.getResultsSummaries(this.$store);
			const residualsSummaries = this.regression ? dataGetters.getResidualsSummaries(this.$store) : [];

			const summaryGroups = pipelines.map(pipeline => {
				const pipelineId = pipeline.pipelineId;
				const requestId = pipeline.requestId;
				const resultSummary = _.find(resultSummaries, summary => {
					return summary.pipelineId === pipelineId;
				});
				const residualSummary = _.find(residualsSummaries, summary => {
					return summary.pipelineId === pipelineId;
				});
				return {
					requestId: requestId,
					pipelineId: pipelineId,
					groupName: pipeline ? pipeline.name : '',
					resultSummary: resultSummary,
					residualsSummary: residualSummary
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
