<template>
	<div class='variable-summaries'>
		<p class="nav-link font-weight-bold">Feature Summaries</p>
		<variable-facets
			enable-search
			enable-toggle
			instance-name="variable-summary-facets"
			:groups="groups"
			:dataset="dataset"
			:enable-type-change="enableTypeChance">
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
		'groups': Array,
		'dataset': String,
		'enableTypeChance': Boolean
	}

});
</script>

<style>
.variable-summaries {
	display: flex;
	flex-direction: column;
}
</style>
