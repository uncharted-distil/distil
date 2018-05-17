<template>
	<div class='variable-summaries'>
		<p class="nav-link font-weight-bold">Feature Summaries</p>
		<variable-facets
			enable-search
			:instance-name="instanceName"
			:rows-per-page="numRowsPerPage"
			:groups="groups"
			:dataset="dataset"
			:enable-highlighting="enableHighlighting"
			:enable-type-change="enableTypeChange">
		</variable-facets>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import VariableFacets from '../components/VariableFacets';
import { NUM_PER_PAGE } from '../util/data';

import 'font-awesome/css/font-awesome.css';

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
	},

	computed: {
		numRowsPerPage(): number {
			return NUM_PER_PAGE;
		}
	}

});
</script>

<style>
.variable-summaries {
	display: flex;
	flex-direction: column;
}
</style>
