<template>
	<div class='variable-summaries'>
		<p class="nav-link font-weight-bold">Feature Summaries</p>
		<variable-facets
			enable-search
			:instance-name="instanceName"
			:groups="groups"
			:dataset="dataset"
			:enable-highlighting="enableHighlighting"
			:enable-type-change="enableTypeChange">
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
		groups: Array,
		dataset: String,
		instanceName: String,
		enableHighlighting: Boolean,
		enableTypeChange: Boolean
	}

});
</script>

<style>
.variable-summaries {
	display: flex;
	flex-direction: column;
}
</style>
