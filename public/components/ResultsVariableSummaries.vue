<template>
	<div class='variable-summaries'>
		<h6 class="nav-link">Feature Summaries</h6>
		<result-variable-facets
			enable-search
			enable-toggle
			enable-group-collapse
			enable-facet-filtering
			instance-name="varSummaries"
			:variables="variables"
			:dataset="dataset"
			:type-change="typeChange">
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
