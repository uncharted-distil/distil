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
		this.$store.watch(() => component.$store.state.variableSummaries, histograms => {

			// convert the histo data into facets data
			const groups = [];

			histograms.forEach(d => {

				if (d.err) {
					// error
					groups.push({
						label: `Error: ${d.err}`,
						facets: []
					});
					return;
				}

				if (d.pending) {
					// pending
					groups.push({
						label: d.histogram.name,
						facets: [
							{
								placeholder: true
							}
						]
					});
					return;
				}

				const histogram = d.histogram;

				switch (histogram.type) {
					case 'categorical':
						groups.push({
							label: histogram.name,
							key: 'category',
							facets: histogram.buckets.map(b => {
								return {
									value: b.key,
									count: b.count
								};
							})
						});
						return;

					case 'numerical':
						groups.push({
							label: histogram.name,
							key: 'float',
							facets: [
								{
									histogram: {
										slices: histogram.buckets.map(b => {
											return {
												label: b.key,
												count: b.count
											};
										})
									}
								}
							]
						});
						return;

					default:
						console.warn('unrecognized histogram type', histogram.type);
						return;
				}
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
