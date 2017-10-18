import { spinnerHTML } from '../util/spinner';

// creates the set of facets from the supplied summary data
export function createGroups(summaries, enableCollapse, enableFiltering, predictedValues) {
	return summaries.map(summary => {
		if (summary.err) {
			// create error facet
			return createErrorFacet(summary, enableCollapse, predictedValues);
		}
		if (summary.pending) {
			// create pending facet
			return createPendingFacet(summary, enableCollapse, predictedValues);
		}
		// create facet
		return createSummaryFacet(summary, enableCollapse, enableFiltering, predictedValues);
	}).filter(group => {
		// remove null groups
		return group;
	});
}

// creates a facet to display a data fetch error
export function createErrorFacet(summary, enableCollapse, predictedValues) {
	return {
		label: summary.name,
		key: predictedValues ? `Predicted ${summary.feature}` : summary.name,
		collapsible: enableCollapse,
		facets: [{
			placeholder: true,
			html: `<div>${summary.err}</div>`
		}]
	};
}

// creates a place holder facet to dispay a spinner
export function createPendingFacet(summary, enableCollapse, predictedValues) {
	return {
		label: summary.name,
		key: predictedValues ? `Predicted ${summary.feature}` : summary.name,
		collapsible: enableCollapse,
		facets: [{
			placeholder: true,
			html: spinnerHTML()
		}]
	};
}

// creates categorical or numerical summary facets
export function createSummaryFacet(summary, enableCollapse, enableFiltering, predictedValues) {
	switch (summary.type) {
		case 'categorical':
			return {
				label: summary.name,
				key: predictedValues ? `Predicted ${summary.feature}` : summary.name,
				collapsible: enableCollapse,
				facets: summary.buckets.map(b => {
					return {
						icon : {
							class : 'fa fa-info'
						},
						value: b.key,
						count: b.count,
						selected: {
							count: b.count
						},
						filterable: enableFiltering
					};
				})
			};

		case 'numerical':
			return {
				label: summary.name,
				key: predictedValues ? `Predicted ${summary.feature}` : summary.name,
				collapsible: enableCollapse,
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
						},
						filterable: enableFiltering
					}
				]
			};
	}
	console.warn('unrecognized summary type', summary.type);
	return null;
}
