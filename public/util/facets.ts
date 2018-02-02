import _ from 'lodash';
import { spinnerHTML } from '../util/spinner';
import { VariableSummary, Extrema } from '../store/data/index';

export const CATEGORY_NO_MATCH_COLOR = "#e05353";
export const CATEGORY_MATCH_COLOR = "#03c6e1";

export interface PlaceHolderFacet {
	placeholder: boolean;
	html: string;
}

export interface Segment {
	color: string;
	count: number;
}

export interface SelectedSegments {
	selected: number;
	segments: Segment[];
}

export interface CategoricalFacet {
	icon: { class: string };
	selected: { count: number } | SelectedSegments;
	value: string;
	count: number;
	filterable: boolean;
	segments: Segment[];
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
	type: string;
	collapsible: boolean;
	collapsed: boolean;
	facets: (PlaceHolderFacet | CategoricalFacet | NumericalFacet)[];
	total?: number,
	count?: number
}

// creates the set of facets from the supplied summary data
export function createGroups(summaries: VariableSummary[], enableCollapse: boolean, enableFiltering: boolean, extrema: Extrema = null): Group[] {
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
		return createSummaryFacet(summary, enableCollapse, enableFiltering, extrema);
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
		type: summary.varType,
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
		type: summary.varType,
		collapsible: enableCollapse,
		collapsed: false,
		facets: [{
			placeholder: true,
			html: spinnerHTML()
		}]
	};
}

// creates categorical or numerical summary facets based on the input summary type
export function createSummaryFacet(summary: VariableSummary, enableCollapse: boolean, enableFiltering: boolean, extrema: Extrema): Group {
	switch (summary.type) {
		case 'categorical':
			return createCategoricalSummaryFacet(summary, enableCollapse, enableFiltering, extrema);
		case 'numerical':
			return createNumericalSummaryFacet(summary, enableCollapse, enableFiltering, extrema);
	}
	console.warn('unrecognized summary type', summary.type);
	return null;
}

export function getGroupIcon(summary: VariableSummary): string {
	switch (summary.varType) {
		case 'categorical':
		case 'ordinal':
		case 'boolean':
			return 'fa fa-info';

		case 'address':
		case 'city':
		case 'state':
		case 'country':
			return 'fa fa-globe';

		case 'email':
		case 'postal_code':
			return 'fa fa-envelope';

		case 'phone':
			return 'fa fa-phone';

		case 'uri':
		case 'keyword':
			return 'fa fa-book';

		case 'dateTime':
			return 'fa fa-calendar';

		default:
			return 'fa fa-info';
	}
}

// creates a categorical facet with segments based on nest buckets counts, or no segments if buckets aren't nested
function createCategoricalSummaryFacet(summary: VariableSummary, enableCollapse: boolean, enableFiltering: boolean, extrema: Extrema): Group {

	// generate facets from the supplied variable summary
	const facets = summary.buckets.map(b => {

		let segments = [];
		let selected = null;

		// Populate segments if buckets are nested.  If a nested bucket's key matches the parent bucket key, values
		// are given a colour to signify a match, all other nested buckets are summed and displayed as not matching.
		if (b.buckets) {
			segments.push( { color: CATEGORY_MATCH_COLOR, count: 0 });
			segments.push( { color: CATEGORY_NO_MATCH_COLOR, count: 0 });
			for (const subBucket of b.buckets) {
				if (subBucket.key === b.key) {
					segments[0].count = subBucket.count;
				} else {
					segments[1].count += subBucket.count;
				}
			}
			// TODO: Add proper highlight state visuals once highlighting is cleaned up
			selected = { segments: segments, selected: b.count };
		} else {
			// if no segments, just use basic count selection
			selected = { count: b.count };
		}

		const facet: CategoricalFacet = {
			icon : { class : getGroupIcon(summary) },
			value: b.key,
			count: b.count,
			selected: selected,
			segments: segments,
			filterable: enableFiltering
		};

		return facet;
	})

	let total = undefined;
	let count = undefined;
	// if (extrema) {
	// 	total = extrema.max;
	// 	count = extrema.max;
	// }

	// Generate a facet group
	return {
		label: summary.name,
		key: summary.name,
		type: summary.varType,
		collapsible: enableCollapse,
		collapsed: false,
		facets: facets,
		total: total,
		count: count
	};
}

function createNumericalSummaryFacet(summary: VariableSummary, enableCollapse: boolean, enableFiltering: boolean, extrema: Extrema): Group {

	const span = _.toNumber(summary.buckets[1].key) - _.toNumber(summary.buckets[0].key);

	const slices = summary.buckets.map((b, i) => {
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
	});

	if (extrema) {

		const minDist = summary.extrema.min - extrema.min;
		if (minDist > 0) {
			for (let i=0; i<minDist; i+=span) {
				const prev = slices[0];
				const label = `${Math.max(extrema.min, _.toNumber(prev.label) - span)}`
				const toLabel = prev.label;
				slices.unshift({
					label: label,
					toLabel: toLabel,
					count: 0
				});
			}
		}
		const maxDist = extrema.max - summary.extrema.max;
		if (maxDist > 0) {
			for (let i=0; i<maxDist; i+=span) {
				const next = slices[slices.length - 1];
				const label = next.toLabel;
				const toLabel = `${Math.min(extrema.max, _.toNumber(next.toLabel) + span)}`;
				slices.push({
					label: label,
					toLabel: toLabel,
					count: 0
				});
			}
		}
	}

	return {
		label: summary.name,
		key: summary.name,
		type: summary.varType,
		collapsible: enableCollapse,
		collapsed: false,
		facets: [
			{
				histogram: {
					slices: slices
				},
				filterable: enableFiltering,
				selection: {} as any
			}
		]
	};
}

export function isCategoricalFacet(facet: PlaceHolderFacet | CategoricalFacet | NumericalFacet): facet is CategoricalFacet {
	return (<CategoricalFacet>facet).value !== undefined;
}

export function isNumericalFacet(facet: PlaceHolderFacet | CategoricalFacet | NumericalFacet): facet is NumericalFacet {
	return (<NumericalFacet>facet).histogram !== undefined;
}

export function isPlaceHolderFacet(facet: PlaceHolderFacet | CategoricalFacet | NumericalFacet): facet is PlaceHolderFacet {
	return (<PlaceHolderFacet>facet).placeholder !== undefined;
}
