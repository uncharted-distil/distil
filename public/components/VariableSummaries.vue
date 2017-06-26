<template>
	<div id='variable-summaries'>
	</div>
</template>

<script>

import Facets from '@uncharted.software/stories-facets';
import '@uncharted.software/stories-facets/dist/facets.css';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';

export default {
	name: 'variable-summaries',

	mounted() {
		const component = this;

		// instantiate the external facets widget
		const container = document.getElementById('variable-summaries');
		const facets = new Facets(container, []);
		const groups = new Map();
		const pending = new Map();
		const errors = new Map();

		// handle a facet going from collapsed to expanded by updating the state in 
		// the store
		facets.on('facet-group:expand', (evt, key) => {
			component.$store.commit('setVarEnabled', { name: key, enabled: true });
			component.$store.dispatch('updateFilteredData', component.$store.getters.getActiveDataset().name);
		});

		// handle a facet going from expanded to collapsed by updating the state in
		// the store
		facets.on('facet-group:collapse', (evt, key) => {
			component.$store.commit('setVarEnabled', { name: key, enabled: false });
			component.$store.dispatch('updateFilteredData', component.$store.getters.getActiveDataset().name);
		});

		// handle a facet changing its filter range by updating the store
		facets.on(' facet-histogram:rangechangeduser', (evt, key, value) => {			
			component.$store.commit('setVarFilterRange', { 
				name: key,
				min: parseFloat(value.from.label[0]),
				max: parseFloat(value.to.label[0])
			});
			component.$store.dispatch('updateFilteredData', component.$store.getters.getActiveDataset().name);
		});
		

		// on dataset change, clear all the components and reset the filter state
		component.$store.watch(() => component.$store.state.activeDataset, () => {
			groups.clear();
			pending.clear();
			errors.clear();
			facets.replace([]);
			component.$store.commit('setFilterState', {});
		});

		// update it's contents when the dataset changes		
		this.$store.watch(() => this.$store.state.variableSummaries, histograms => {

			const bulk = [];
			// for each histogram
			histograms.forEach(histogram => {

				const key = histogram.name;

				if (histogram.err) {
					// check if already added as error
					if (errors.has(key)) {
						return;
					}
					// add error group
					const group = {
						label: histogram.name,
						key: key,
						facets: [{
							placeholder: true,
							key: 'placeholder',
							html: `<div>${histogram.err}</div>`
						}]
					};
					facets.replaceGroup(group);
					errors.set(key, group);
					pending.delete(key);
					return;
				}

				if (histogram.pending) {
					// check if already added as placeholder
					if (pending.has(key)) {
						return;
					}
					// add placeholder
					const group = {
						label: histogram.name,
						key: key,
						facets: [
							{
								placeholder: true,
								key: 'placeholder',
								html: `
									<div class="bounce1"></div>
									<div class="bounce2"></div>
									<div class="bounce3"></div>`
							}
						]
					};
					bulk.push(group);
					pending.set(key, group);
					return;
				}

				// check if already added
				if (groups.has(key)) {
					return;
				}

				let group;
				switch (histogram.type) {
					case 'categorical':
						group = {
							label: histogram.name,
							key: key,
							facets: histogram.buckets.map(b => {
								return {
									value: b.key,
									count: b.count
								};
							})
						};
						break;

					case 'numerical':
						group = {
							label: histogram.name,
							key: key,
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
						};
						break;

					default:
						console.warn('unrecognized histogram type', histogram.type);
						return;
				}

				// append
				facets.replaceGroup(group);
				// track
				groups.set(key, group);
				pending.delete(key);
				errors.delete(key);
			});

			if (bulk.length > 0) {
				facets.replace(bulk);
			}
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
