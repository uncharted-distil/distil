import { spinnerHTML } from '../util/spinner';

// creates the set of facets from the supplied summary data
export function createGroups(summaries) {
	return summaries.map(summary => {
		if (summary.err) {
			// create error facet
			return createErrorFacet(summary);
		}
		if (summary.pending) {
			// create pending facet
			return createPendingFacet(summary);
		}
		// create facet
		return createSummaryFacet(summary);
	}).filter(group => {
		// remove null groups
		return group;
	});
}

// creates a facet to display a data fetch error
export function createErrorFacet(summary) {
	return {
		label: summary.name,
		key: summary.name,
		facets: [{
			placeholder: true,
			html: `<div>${summary.err}</div>`
		}]
	};
}

// creates a place holder facet to dispay a spinner
export function createPendingFacet(summary) {
	return {
		label: summary.name,
		key: summary.name,
		facets: [{
			placeholder: true,
			html: spinnerHTML()
		}]
	};
}

// creates categorical or numerical summary facets
export function createSummaryFacet(summary) {
	switch (summary.type) {

		case 'categorical':
			return {
				label: summary.name,
				key: summary.name,
				facets: summary.buckets.map(b => {
					return {
						icon : {
							class : 'fa fa-info'
						},
						value: b.key,
						count: b.count,
						selected: {
							count: b.count
						}
					};
				})
			};

		case 'numerical':
			return {
				label: summary.name,
				key: summary.name,
				facets: [
					{
						histogram: {
							slices: summary.buckets.map((b, i) => {
								let toLabel;
								if (i < summary.buckets.length-1) {
									toLabel = summary.buckets[i+1].key;
								} else {
									toLabel = `${summary.extrema.max}`;
								}
								return {
									label: b.key,
									toLabel: toLabel,
									count: b.count
								};
							})
						}
					}
				]
			};
	}
	console.warn('unrecognized summary type', summary.type);
	return null;
}
