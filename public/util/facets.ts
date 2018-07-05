import _ from 'lodash';
import { spinnerHTML } from '../util/spinner';
import { formatValue } from '../util/types';
import { VariableSummary } from '../store/dataset/index';

export const CATEGORICAL_CHUNK_SIZE = 10;
export const IMAGE_CHUNK_SIZE = 5;

export interface PlaceHolderFacet {
	placeholder: boolean;
	html: string;
	filterable: boolean;
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
	countLabel: string;
	filterable: boolean;
	segments: Segment[];
	file: string;
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
	more?: number;
	moreTotal?: number;
	total?: number;
	remaining?: (PlaceHolderFacet | CategoricalFacet | NumericalFacet)[];
}

// creates the set of facets from the supplied summary data
export function createGroups(summaries: VariableSummary[]): Group[] {
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
export function createErrorFacet(summary: VariableSummary): Group {
	return {
		label: summary.label,
		key: summary.key,
		type: summary.varType,
		collapsible: false,
		collapsed: false,
		facets: [{
			placeholder: true,
			html: `<div>${summary.err}</div>`,
			filterable: false
		}],
		numRows: 0
	};
}

// creates a place holder facet to dispay a spinner
export function createPendingFacet(summary: VariableSummary): Group {
	return {
		label: summary.label,
		key: summary.key,
		type: summary.varType,
		collapsible: false,
		collapsed: false,
		facets: [{
			placeholder: true,
			html: spinnerHTML(),
			filterable: false
		}],
		numRows: 0
	};
}

// creates categorical or numerical summary facets based on the input summary type
export function createSummaryFacet(summary: VariableSummary): Group {
	switch (summary.type) {
		case 'categorical':
			return createCategoricalSummaryFacet(summary);
		case 'numerical':
			return createNumericalSummaryFacet(summary);
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

export function getCategoricalChunkSize(type: string):number {
	if (type === 'image') {
		return IMAGE_CHUNK_SIZE;
	}
	return CATEGORICAL_CHUNK_SIZE;
}

// creates a categorical facet
function createCategoricalSummaryFacet(summary: VariableSummary): Group {
	let total = 0;
	const facets =  summary.buckets.map((b, index) => {
		const segments = [];
		const selected = {
			count: b.count
		};
		let countLabel = b.count.toString();

		const facet: CategoricalFacet = {
			icon : {
				class : getGroupIcon(summary)
			},
			value: b.key,
			countLabel: countLabel,
			count: b.count,
			selected: selected,
			segments: segments,
			filterable: false,
			file: summary.files ? summary.files[index] : null
		};
		total += b.count;
		return facet;
	});

	facets.sort((a, b) => {
		return b.count - a.count;
	});

	const chunkSize = getCategoricalChunkSize(summary.varType);
	const top = facets.slice(0, chunkSize)
	const remaining = (facets.length > chunkSize) ? facets.slice(chunkSize) : [];
	let remainingTotal = 0;
	remaining.forEach(facet => {
		remainingTotal += facet.count;
	});

	// Generate a facet group
	return {
		label: summary.label,
		key: summary.key,
		type: summary.varType,
		collapsible: false,
		collapsed: false,
		facets: top,
		total: total,
		numRows: summary.numRows,
		more: remaining.length,
		moreTotal: remainingTotal,
		remaining: remaining
	};
}

function getHistogramSlices(summary: VariableSummary) {
	const buckets = summary.buckets;
	const extrema = summary.extrema;
	const slices = new Array(buckets.length);
	for (let i=0; i<buckets.length; i++) {
		const bucket = buckets[i];
		const from = _.toNumber(bucket.key);
		const to = (i < buckets.length-1) ? _.toNumber(buckets[i+1].key) : extrema.max;
		slices[i] = {
			label: `${formatValue(from, summary.varType)}`,
			toLabel: `${formatValue(to, summary.varType)}`,
			count: bucket.count
		};
	}
	return slices;
}

function createNumericalSummaryFacet(summary: VariableSummary): Group {
	const slices = getHistogramSlices(summary);
	return {
		label: summary.label,
		key: summary.key,
		type: summary.varType,
		collapsible: false,
		collapsed: false,
		facets: [
			{
				histogram: {
					slices: slices
				},
				filterable: false,
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
