import { spinnerHTML } from '../util/spinner';
import { VariableSummary } from '../store/data/index';

export interface PlaceHolderFacet {
	placeholder: boolean;
	html: string;
}

export interface CategoricalFacet {
	icon: { class: string };
	selected: { count: number };
	value: string;
	count: number;
	filterable: boolean;
}

export interface Slice {
	label: string;
	toLabel: string;
	count: number;
}

export interface Selection {
	range: {
		to: string;
		from: string;
	}
}

export interface NumericalFacet {
	histogram: { slices: Slice[] };
	filterable: boolean;
	selection: Selection;
}

export interface Group {
	label: string;
	key: string;
	collapsible: boolean;
	collapsed: boolean;
	facets: (PlaceHolderFacet | CategoricalFacet | NumericalFacet)[];
}

// creates the set of facets from the supplied summary data
export function createGroups(summaries: VariableSummary[], enableCollapse: boolean, enableFiltering: boolean): Group[] {
	return summaries.map(summary => {
		if (summary.err) {
			// create error facet
			return createErrorFacet(summary, enableCollapse);
		}
		if (summary.pending) {
			// create pending facet
			return createPendingFacet(summary, enableCollapse);
		}
		// create facet
		return createSummaryFacet(summary, enableCollapse, enableFiltering);
	}).filter(group => {
		// remove null groups
		return group;
	});
}

// creates a facet to display a data fetch error
export function createErrorFacet(summary: VariableSummary, enableCollapse: boolean): Group {
	return {
		label: summary.name,
		key: summary.name,
		collapsible: enableCollapse,
		collapsed: false,
		facets: [{
			placeholder: true,
			html: `<div>${summary.err}</div>`
		}]
	};
}

// creates a place holder facet to dispay a spinner
export function createPendingFacet(summary: VariableSummary, enableCollapse: boolean): Group {
	return {
		label: summary.name,
		key: summary.name,
		collapsible: enableCollapse,
		collapsed: false,
		facets: [{
			placeholder: true,
			html: spinnerHTML()
		}]
	};
}

// creates categorical or numerical summary facets
export function createSummaryFacet(summary: VariableSummary, enableCollapse: boolean, enableFiltering: boolean): Group {
	switch (summary.type) {

		case 'categorical':
			return {
				label: summary.name,
				key: summary.name,
				collapsible: enableCollapse,
				collapsed: false,
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
				key: summary.name,
				collapsible: enableCollapse,
				collapsed: false,
				facets: [
					{
						histogram: {
							slices: summary.buckets.map((b, i) => {
								let toLabel: string;
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
						filterable: enableFiltering,
						selection: {} as any
					}
				]
			};
	}
	console.warn('unrecognized summary type', summary.type);
	return null;
}
