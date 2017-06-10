<template>
	<div id='search-bar'>
		<input v-model="terms" type='search' name='datasetsearch'>
	</div>
</template>

<script>

import _ from 'lodash';
import axios from 'axios';

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
		terms: _.throttle(function(newTerms) {
			const component = this;
			axios.get('/distil/datasets?search=' + newTerms)
				.then(response => {
					if (!_.isEmpty(response.data.datasets)) {
						 component.$store.commit('setDatasets', response.data.datasets);
					} else {
						component.$store.commit('setDatasets', []);
					}
				})
				.catch(error => console.log(error));
		}, 500)
	},
};
</script>

<style>

</style>

