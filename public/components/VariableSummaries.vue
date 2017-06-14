<template>
	<div id='variable-summaries'>
	</div>
</template>

<script>

import Facets from '@uncharted.software/stories-facets';
import '@uncharted.software/stories-facets/dist/facets.css';
import 'font-awesome/css/font-awesome.css';

export default {
	name: 'variable-summaries',
	mounted() {
		// instantiate the external facets widget
		const container = document.getElementById('variable-summaries');
		const facets = new Facets(container, []);
		
		// update it's contents when the dataset changes
		// any event handlers would be added here as well
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
	height: 80vh;
	padding: 5px;
}
</style>
