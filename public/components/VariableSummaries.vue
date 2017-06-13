<template>
	<div id='variable-summaries'>
		<!--<div v-for='variable in variables'>
						{{variable.name}}:{{variable.type}}
					</div>-->
	</div>
</template>

<script>

import Facets from '@uncharted.software/stories-facets';
import '@uncharted.software/stories-facets/dist/facets.css';
import 'font-awesome/css/font-awesome.css';
//import * as getters from '../store/getters';

export default {
	name: 'variable-summaries',
	mounted() {
		const container = document.getElementById('variable-summaries');
		const facets = new Facets(container, []);
		const component = this;
		this.$store.watch(() => component.$store.state.variableSummaries, (data) => {
			// convert the histo data into facets data
			const groups = data.histograms.map(d => {
				return ({
					label: d.name,
					key: 'float',
					facets: [
						{
							histogram: {
								slices: d.buckets.map(b => {
									return ({
										label: b.key,
										count: b.count
									});
								})
							}
						}
					]
				});
			});
			facets.replace(groups);
		});
	}
};
</script>

<style>
#variable-summaries {
	width: 240px;
	padding: 5px;
	font-family: Helvetica;
}
</style>
