<template>

	<sparkline-timeseries-view
		disable-highlighting
		:instance-name="instanceName"
		:include-active="includedActive"
		:variable-summaries="variableSummaries"
		:items="items"
		:fields="fields"
		:predictedCol="predictedCol">
	</sparkline-timeseries-view>

</template>

<script lang="ts">

import Vue from 'vue';
import SparklineTimeseriesView from './SparklineTimeseriesView';
import { Dictionary } from '../util/dict';
import { VariableSummary, TableColumn, TableRow, } from '../store/dataset/index';
import { getters as routeGetters } from '../store/route/module';
import { getters as solutionGetters } from '../store/solutions/module';
import { getters as resultsGetters } from '../store/results/module';
import { Solution } from '../store/solutions/index';

export default Vue.extend({
	name: 'results-timeseries-view',

	components: {
		SparklineTimeseriesView
	},

	props: {
		items: Array as () => TableRow[],
		fields: Object as () => Dictionary<TableColumn>,
		instanceName: String as () => string,
		includedActive: Boolean as () => boolean
	},

	computed: {

		variableSummaries(): VariableSummary[] {
			const training = resultsGetters.getTrainingSummaries(this.$store);
			const target = resultsGetters.getTargetSummary(this.$store);
			return training.concat(target);
		},

		solution(): Solution {
			return solutionGetters.getActiveSolution(this.$store);
		},

		predictedCol(): string {
			return this.solution ? this.solution.predictedKey : '';
		}
	}

});
</script>

<style>
</style>
