<template>
	<div class='variable-summaries'>
		<p class="nav-link font-weight-bold">Feature Summaries</p>
		<result-variable-facets
			enable-search
			enable-highlighting
			instance-name="result-summary-facets"
			:groups="groups"
			:dataset="dataset"
			:enable-type-change="enableTypeChance">
		</result-variable-facets>
	</div>
</template>

<script lang="ts">

import ResultVariableFacets from './ResultsVariableFacets';
import VariableSummaries from './VariableSummaries';
import 'font-awesome/css/font-awesome.css';
import Vue from 'vue';

export default Vue.extend({
	name: 'results-variable-summaries',

	extends: VariableSummaries,

	components: {
		ResultVariableFacets
	}
});
</script>
