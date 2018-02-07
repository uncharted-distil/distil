import _ from 'lodash';
import { spinnerHTML } from '../util/spinner';
import { formatValue } from '../util/types';
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
	numRows: number;
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
		label: summary.label ? summary.label : summary.name,
		key: summary.name,
		type: summary.varType,
		collapsible: enableCollapse,
		collapsed: false,
		facets: [{
			placeholder: true,
			html: `<div>${summary.err}</div>`
		}],
		numRows: 0
	};
}

// creates a place holder facet to dispay a spinner
export function createPendingFacet(summary: VariableSummary, enableCollapse: boolean): Group {
	return {
		label: summary.label ? summary.label : summary.name,
		key: summary.name,
		type: summary.varType,
		collapsible: enableCollapse,
		collapsed: false,
		facets: [{
			placeholder: true,
			html: spinnerHTML()
		}],
		numRows: 0
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

	facets.sort((a, b) => {
		return b.count - a.count;
	});

	// Generate a facet group
	return {
		label: summary.label ? summary.label : summary.name,
		key: summary.name,
		type: summary.varType,
		collapsible: enableCollapse,
		collapsed: false,
		facets: facets,
		numRows: summary.numRows
	};
}

function truncateTowardsZero(num: number): number {
	if (num < 0) {
		return Math.ceil(num);
	}
	return Math.floor(num);
}

function hackyBinning(summary: VariableSummary, extrema: Extrema) {
	const NUM_BUCKETS = 50;
	const range = extrema.max - extrema.min;
	const span = range / NUM_BUCKETS;
	const buckets = new Array(NUM_BUCKETS);
	for (let i=0; i<NUM_BUCKETS; i++) {
		const from = extrema.min + (i * span);
		const to = extrema.min + ((i + 1) * span);
		buckets[i] = {
			label: `${formatValue(from, summary.varType)}`,
			toLabel: `${formatValue(to, summary.varType)}`,
			count: 0
		};
	}
	for (let i=0; i<summary.buckets.length; i++) {
		const bucket = summary.buckets[i];
		if (bucket.count === 0) {
			continue;
		}
		const bucketKey = _.toNumber(bucket.key);
		if (bucketKey < extrema.min || bucketKey > extrema.max) {
			continue;
		}
		const index = truncateTowardsZero((bucketKey / span) - (extrema.min / span));
		buckets[index].count += bucket.count;
	}
	return buckets;
}

function getHistogramSlices(summary: VariableSummary, extrema: Extrema) {
	if (extrema && !_.isNaN(extrema.min) && !_.isNaN(extrema.max)) {
		return hackyBinning(summary, extrema);
	}
	const buckets = summary.buckets;
	const slices = new Array(buckets.length);
	for (let i=0; i<buckets.length; i++) {
		const bucket = buckets[i];
		const from = _.toNumber(bucket.key);
		const to = (i < buckets.length-1) ? _.toNumber(buckets[i+1].key) : summary.extrema.max;
		slices[i] = {
			label: `${formatValue(from, summary.varType)}`,
			toLabel: `${formatValue(to, summary.varType)}`,
			count: bucket.count
		};
	}
	return slices;
}

function createNumericalSummaryFacet(summary: VariableSummary, enableCollapse: boolean, enableFiltering: boolean, extrema: Extrema): Group {
	const slices = getHistogramSlices(summary, extrema);
	return {
		label: summary.label ? summary.label : summary.name,
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
		],
		numRows: summary.numRows
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
