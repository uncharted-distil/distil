<template>
	<div id='search-bar'>
		<b-form-input v-model="terms" type="text" placeholder="Search datasets" name="datasetsearch"/>
	</div>
</template>

<script>

import _ from 'lodash';

export default {
	name: 'search-bar',

	// control local data
	data() {
		return {
			terms: ''
		};
	},

	// data change handlers
	watch: {
		// issues a debounced search request to the server
		terms: _.throttle(function (newTerms) {
			this.$store.dispatch('searchDatasets', newTerms);
		}, 500)
	},
};
</script>

<style>

</style>
