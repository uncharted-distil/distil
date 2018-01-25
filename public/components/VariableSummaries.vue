<template>
	<div class='variable-summaries'>
		<p class="nav-link font-weight-bold">Feature Summaries</p>
		<variable-facets
			enable-search
			enable-toggle
			enable-group-collapse
			enable-facet-filtering
			instance-name="variable-summary-facets"
			:variables="variables"
			:dataset="dataset"
			:type-change="typeChange">
		</variable-facets>
	</div>
</template>

<script lang="ts">

import VariableFacets from '../components/VariableFacets';
import 'font-awesome/css/font-awesome.css';
import Vue from 'vue';

export default Vue.extend({
	name: 'variable-summaries',

	components: {
		VariableFacets
	},

	props: {
		'variables': Array,
		'dataset': String,
		'typeChange': Boolean
	}

});
</script>

<style>
.variable-summaries {
	display: flex;
	flex-direction: column;
}
</style>
