<template>
	<div class='result-facets'>
		<result-group class="result-group-container" :key="group.name" v-for="group in resultGroups"
			:name="group.groupName"
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
import { getters as dataGetters} from '../store/data/module';
import { getters as routeGetters} from '../store/route/module';
import { getPipelineResult } from '../util/pipelines';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';
import Vue from 'vue';

/*eslint-disable */
interface SummaryGroup {
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
		resultGroups(): SummaryGroup[] {
			// Generate pairs of residuals and results for each pipeline in the numerical case.  Categorical
			//
			const resultSummaries = dataGetters.getResultsSummaries(this.$store);
			const residualsSummaries = this.regression ? dataGetters.getResidualsSummaries(this.$store) : [];

			const requestId = routeGetters.getRouteCreateRequestId(this.$store);
			const summaryGroups = resultSummaries.map(s => {
				const residuals = _.find(residualsSummaries, f => s.pipelineId === f.pipelineId);
				const result = getPipelineResult(this.$store.state.pipelineModule, requestId, s.pipelineId);
				return {
					groupName: result ? result.name : '',
					resultSummary: s,
					residualsSummary: residuals
				};
			});

			// Sort alphabetically
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
