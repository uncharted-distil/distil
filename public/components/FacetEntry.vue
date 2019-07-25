<template>
	<div class="facets-root" v-bind:class="{ 'highlighting-enabled': enableHighlighting }">
		<div class="facet-tooltip" style="display:none;"></div>
	</div>
</template>

<script lang="ts">
import _ from 'lodash';
import $ from 'jquery';
import Vue from 'vue';

import IconFork from './icons/IconFork';
import IconBookmark from './icons/IconBookmark';
import { createIcon } from '../util/icon';
import { createGroup, Group, CategoricalFacet, isCategoricalFacet, getCategoricalChunkSize, isNumericalFacet, isSparklineFacet } from '../util/facets';
import { VariableSummary, Highlight, RowSelection, Row } from '../store/dataset/index';
import { Dictionary } from '../util/dict';
import { getSelectedRows } from '../util/row';
import Facets from '@uncharted.software/stories-facets';
import ImagePreview from '../components/ImagePreview';
import TypeChangeMenu from '../components/TypeChangeMenu';
import { getVarType, isClusterType, isFeatureType, addClusterPrefix, addFeaturePrefix, hasComputedVarPrefix, GEOCOORDINATE_TYPE } from '../util/types';
import { IMPORTANT_VARIABLE_RANKING_THRESHOLD } from '../util/data';
import { getters as datasetGetters } from '../store/dataset/module';

import '@uncharted.software/stories-facets/dist/facets.css';

export default Vue.extend({
	name: 'facet-entry',

	props: {
		summary: Object as () => VariableSummary,
		highlight: Object as () => Highlight,
		rowSelection: Object as () => RowSelection,
		deemphasis: Object as () => any,
		enableTypeChange: Boolean as () => boolean,
		enableHighlighting: Boolean as () => boolean,
		showOrigin: Boolean as () => boolean,
		ignoreHighlights: Boolean as () => boolean,
		instanceName: String as () => string,
		html: [ String as () => string, Object as () => any, Function as () => Function ],
	},

	data() {
		const component = this as any;
		return {
			menus: {} as any,
			facets: {} as any,
			numToDisplay: {} as Dictionary<number>,
			numAddedToDisplay: {} as Dictionary<number>
		};
	},

	mounted() {

		
		const component = this;

		// Instantiate the external facets widget. The facets maintain their own copies
		// of group objects which are replaced wholesale on changes.  Elsewhere in the code
		// we modify local copies of the group objects, then replace those in the Facet component
		// with copies.
		this.facets = new Facets(this.$el, [ this.groupSpec ]);

		// Call customization hook

		if (!this.isFacetTypeGeocoord) {
			this.injectHTML(this.groupSpec, this.facets.getGroup(this.groupSpec.key)._element);
		this.augmentGroup(this.groupSpec, this.facets.getGroup(this.groupSpec.key));
		this.updateImportantBadge(this.groupSpec);

		// proxy events
		this.facets.on('facet-group:expand', (event: Event, key: string) => {
			component.$emit('expand', this.groupSpec.colName);
		});

		this.facets.on('facet-group:collapse', (event: Event, key: string) => {
			component.$emit('collapse', this.groupSpec.colName);
		});

		this.facets.on('facet-histogram:rangechangeduser', (event: Event, key: string, value: any, facet: any) => {
			const range = {
				from: _.toNumber(value.from.label[0]),
				to: _.toNumber(value.to.label[0])
			};
			component.$emit('range-change', this.instanceName, this.groupSpec.colName, range, facet.dataset);
		});

		// hover over events

		this.facets.on('facet-histogram:mouseenter', (event: Event, key: string, value: any) => {
			component.$emit('histogram-mouse-enter', this.groupSpec.colName, value);
		});

		this.facets.on('facet-histogram:mouseleave', (event: Event, key: string) => {
			component.$emit('histogram-mouse-leave', this.groupSpec.colName);
		});

		this.facets.on('facet:mouseenter', (event: Event, key: string, value: number) => {
			component.$emit('facet-mouse-enter', this.groupSpec.colName, value);
		});

		this.facets.on('facet:mouseleave', (event: Event, key: string) => {
			component.$emit('facet-mouse-leave', this.groupSpec.colName);
		});

		// more events

		this.facets.on('facet-group:more', (event: Event, key: string) => {
			const chunkSize = getCategoricalChunkSize(this.groupSpec.type);
			if (!component.numToDisplay[key]) {
				Vue.set(component.numToDisplay, key, chunkSize);
				Vue.set(component.numAddedToDisplay, key, 0);
			}
			Vue.set(component.numToDisplay, key, component.numToDisplay[key] + chunkSize);
			Vue.set(component.numAddedToDisplay, key, component.numAddedToDisplay[key] + chunkSize);
			component.$emit('facet-more', key);
		});

		this.facets.on('facet-group:less', (event: Event, key: string) => {
			const chunkSize = getCategoricalChunkSize(this.groupSpec.type);
			if (!component.numToDisplay[key]) {
				Vue.set(component.numToDisplay, key, chunkSize);
				Vue.set(component.numAddedToDisplay, key, 0);
			}
			Vue.set(component.numToDisplay, key, component.numToDisplay[key] - chunkSize);
			Vue.set(component.numAddedToDisplay, key, component.numAddedToDisplay[key] - chunkSize);
			component.$emit('facet-less', key);
		});

		// click events

		this.facets.on('facet-histogram:click', (event: Event, key: string, value: any, facet: any) => {
			// if this is a click on value previously used as highlight root, clear
			const range = {
				from: _.toNumber(value.label),
				to: _.toNumber(value.toLabel)
			};
			if (this.isHighlightedValue(this.highlight, this.groupSpec.colName, range)) {
				// clear current selection
				component.$emit('histogram-click', this.instanceName, null, null, facet.dataset);
			} else {
				// set selection
				component.$emit('histogram-click', this.instanceName, this.groupSpec.colName, range, facet.dataset);
			}
		});

		this.facets.on('facet:click', (event: Event, key: string, value: string, count: number, facet: any) => {
			// User clicked on the value that is currently the highlight root
			if (this.isHighlightedValue(this.highlight, this.groupSpec.colName, value)) {
				// clear current selection
				component.$emit('facet-click', this.instanceName, null, null, facet.dataset);
			} else {
				// set selection
				component.$emit('facet-click', this.instanceName, this.groupSpec.colName, value, facet.dataset);
			}
		});

		this.facets.on('facet-histogram:mouseenter', (event: Event, key: string, value: any) => {
			const $target = $(event.target);
			const $parent = $target.parent();
			const $root = $(this.$el);
			const $tooltip = $(this.$el).find('.facet-tooltip');
			const TOOLTIP_BUFFER = 8;

			// set the text now so that the dimensions are correct
			$tooltip.html(`<b>${value.label} - ${value.toLabel}</b>`);
			const posX = $parent.offset().left - $root.offset().left;
			const posY = $parent.offset().top - $root.offset().top + $root.scrollTop();
			const offsetX = posX + ($target.outerWidth() / 2) - ($tooltip.outerWidth() / 2);
			const offsetY = posY - $tooltip.height() - TOOLTIP_BUFFER;

			// ensure it doesnt go outside of root
			const x = Math.min($root.width(), Math.max(0, offsetX));
			const y = Math.max(0, offsetY);

			$tooltip.css('left', `${x}px`);
			$tooltip.css('top', `${y}px`);
			$tooltip.show();
		});

		this.facets.on('facet-histogram:mouseleave', (event: Event, key: string, value: string) => {
			$(this.$el).find('.facet-tooltip').hide();
		});
		}
	},

	computed: {
		isFacetTypeGeocoord(): boolean {
			return this.summary.type === GEOCOORDINATE_TYPE;
		},
		ranking(): number {
			const variables = datasetGetters.getVariables(this.$store);
			const v = variables.find(v => v.colName === this.summary.key);
			if (v && v.ranking !== undefined) {
				return v.ranking;
			}
			return 0;
		},
		groupSpec(): Group {

			const group = createGroup(this.summary);

			const numToDisplay = this.numToDisplay[group.key];
			if (numToDisplay) {
				const show = group.all.slice(0, numToDisplay);
				const hide = group.all.slice(numToDisplay);
				group.facets = show;

				let remainingTotal = 0;
				hide.forEach(facet => {
					if (isCategoricalFacet(facet)) {
						remainingTotal += facet.count;
					}
				});
				group.more = group.all.length - numToDisplay;
				group.moreTotal = remainingTotal;

				// track how many are already added
				if (this.numAddedToDisplay[group.key] && this.numAddedToDisplay[group.key] > 0) {
					group.less = this.numAddedToDisplay[group.key];
				}
			}

			// TODO: move this reference to highlights, as it forces a refresh

			if (this.enableHighlighting &&
				this.highlight &&
				this.highlight.context === this.instanceName) {
				if (group.colName === this.highlight.key) {
					group.facets.forEach(facet => {
						facet.filterable = true;
					});
				}
			}

			if (this.showOrigin) {
				group.facets.forEach((facet: any) => {
					if (facet.histogram) {
						facet.histogram.showOrigin = true;
					}
				});
			}

			return group;
		}
	},

	watch: {

		// handle changes to the facet group list
		groupSpec: {
			handler(currGroup: Group, prevGroup: Group) {
				// update and groups
				this.updateGroups(currGroup, prevGroup);
				this.injectHighlights(this.highlight, this.rowSelection, this.deemphasis);
			},
			deep: true
		},

		// handle external highlight changes by updating internal facet select states
		highlight(currHighlights: Highlight) {
			if (this.enableHighlighting) {
				this.addHighlightArrow(currHighlights);
			}
		},

		// handle external highlight changes by updating internal facet select states
		rowSelection(currSelection: RowSelection) {
			this.injectHighlights(this.highlight, currSelection, this.deemphasis);
		},

		deemphasis(currDemphasis: any) {
			this.injectHighlights(this.highlight, this.rowSelection, currDemphasis);
		},
		html(customHtml: () => HTMLElement) {
			const $elem = this.facets.getGroup(this.groupSpec.key)._element;
			this.updateCustomHtml(this.groupSpec, $elem, customHtml);
		}
	},

	methods: {

		injectHTML(group: Group, $elem: JQuery) {

			const $groupFooter = $('<div class="group-footer"><div class="html-slot"></div></div>').appendTo($elem.find('.facets-group'));
			$elem.click(event => {
				if (group.facets.length >= 1) {
					const facet = group.facets[0];
					if (isNumericalFacet(facet)) {

						const slices = facet.histogram.slices;
						const first = slices[0];
						const last = slices[slices.length - 1];
						const range = {
							from: _.toNumber(first.label),
							to: _.toNumber(last.toLabel)
						};
						this.$emit('numerical-click', this.instanceName, group.colName, range, group.dataset);

					} else if (isSparklineFacet(facet)) {

						const points = facet.sparklines ? facet.sparklines[0] : facet.sparkline;
						const first = points[0][0];
						const last = points[points.length - 1][0];
						const range = {
							from: _.toNumber(first),
							to: _.toNumber(last)
						};
						this.$emit('numerical-click', this.instanceName, group.colName, range, group.dataset);

					} else if (isCategoricalFacet(facet)) {
						this.$emit('categorical-click', this.instanceName, group.colName, null, group.dataset);
					}
				}
			});

			$elem.find('.facet-histogram g').click(event => {
				if (group.facets.length >= 1) {
					const facet = group.facets[0];

					if (isNumericalFacet(facet)) {

						const slices = facet.histogram.slices;
						const first = slices[0];
						const last = slices[slices.length - 1];
						const range = {
							from: _.toNumber(first.label),
							to: _.toNumber(last.toLabel)
						};
						this.$emit('numerical-click', this.instanceName, group.colName, range, group.dataset);

					} else if (isSparklineFacet(facet)) {

						const points = facet.sparklines ? facet.sparklines[0] : facet.sparkline;
						const first = points[0][0];
						const last = points[points.length - 1][0];
						const range = {
							from: _.toNumber(first),
							to: _.toNumber(last)
						};
						this.$emit('numerical-click', this.instanceName, group.colName, range, group.dataset);

					}
				}
			});

			// inject type icon in group header
			this.injectTypeIcon(group, $elem);

			// inject type change header menus
			this.injectTypeChangeHeaders(group, $elem);

			// inject html
			this.updateCustomHtml(group, $elem, this.html);

			// inject image preview if image type
			this.injectImagePreview(group, $elem);

			this.injectImportantBadge(group, $elem);
		},

		updateCustomHtml(group: Group, $elem: JQuery, html: any) {
			if (html) {
				const $htmlSlot = $elem.find('.html-slot');
				const customHtml = _.isFunction(this.html) ? this.html(group) : this.html;
				$htmlSlot.empty();
				$htmlSlot.append(customHtml);
				this.$emit('html-appended', customHtml);
			}

		},

		addHighlightArrow(highlight: Highlight) {
			const $elem = $(this.$el);
			// remove previous
			$elem.find('.highlight-arrow').remove();

			// NOTE: first group is a query group, ignore it
			const QUERY_OFFSET = 1;

			const $groups = $elem.find('.facets-group');
			// add highlight arrow
			if (this.isHighlightedGroup(highlight, this.groupSpec.colName)) {
				const $group = $($groups.get(QUERY_OFFSET));
				$group.append('<div class="highlight-arrow"><i class="fa fa-arrow-circle-right fa-2x"></i></div>');
			}
		},

		isHighlightedInstance(highlight: Highlight): boolean {
			return highlight && highlight.context === this.instanceName;
		},

		isHighlightedGroup(highlight: Highlight, colName: string): boolean {
			return this.isHighlightedInstance(highlight) && highlight.key === colName;
		},

		isHighlightedValue(highlight: Highlight, colName: string, value: any): boolean {
			// if not instance, return false
			if (!this.isHighlightedGroup(highlight, colName)) {
				return false;
			}
			if (_.isArray(highlight.value)) {
				return false;
			}
			// if string, check for match
			if (_.isString(highlight.value)) {
				return highlight.value === value;
			}
			// otherwise, check range
			return highlight.value.from === value.from &&
				highlight.value.to === value.to;
		},

		getHighlightValue(highlight: Highlight): any {
			if (highlight && highlight.value) {
				return highlight.value;
			}
			return null;
		},

		selectCategoricalFacet(facet: any, count?: number) {
			if (count === undefined && facet._spec.segments && facet._spec.segments.length > 0) {
				facet.select(facet._spec.segments);
			} else {
				facet.select(count ? count : facet.count);
			}
		},

		deselectCategoricalFacet(facet: any) {
			if (facet._spec.segments && facet._spec.segments.length > 0) {
				facet.select(0);
			} else {
				facet.deselect();
			}
		},

		selectTimeseriesFacet(facet: any, count?: number) {
			facet._sparklineContainer.parent().css('background-color', 'rgba(0, 198, 225, .2)');
		},

		deselectTimeseriesFacet(facet: any) {
			facet._sparklineContainer.parent().css('background-color', '');
		},

		injectDeemphasis(group: any, deemphasis: any) {
			if (!deemphasis) {
				return;
			}

			// clear deemphasis
			for (const facet of group.facets) {

				// only emphasize histograms
				if (facet._histogram) {
					facet._histogram.bars.forEach(bar => {
						if (!bar._element.hasClass('row-selected'))  {
							// NOTE: don't trample row selections
							bar._element.css('fill', '');
						}
					});
				}
			}

			// de-emphasis within range
			for (const facet of group.facets) {

				// only emphasize histograms
				if (facet._histogram) {
					facet._histogram.bars.forEach(bar => {
						const entry: any = _.last(bar.metadata);
						if (_.toNumber(entry.label) >= deemphasis.min &&
							_.toNumber(entry.toLabel) < deemphasis.max) {
							if (!bar._element.hasClass('row-selected'))  {
								// NOTE: don't trample row selections
								bar._element.css('fill', '#ddd');
							}
						}
					});
				}
			}
		},

		findClusterCol(key: string, row: Row) {
			const clustered = addClusterPrefix(key);
			return _.find(row.cols, c => {
				return c.key === clustered;
			});
		},

		findFeatureCol(key: string, row: Row) {
			const feature = addFeaturePrefix(key);
			return _.find(row.cols, c => {
				return c.key === feature;
			});
		},

		addRowSelectionToFacet(facet: any, col: any) {
			if (facet._histogram) {
				facet._histogram.bars.forEach(bar => {
					const entry: any = _.last(bar.metadata);
					if (col.value >= _.toNumber(entry.label) &&
						col.value < _.toNumber(entry.toLabel)) {
						bar._element.css('fill', '#ff0067');
						bar._element.addClass('row-selected');
					}
				});
			} else if (facet._sparkline) {
				// TODO: sparkline
			} else {
				facet._sparklineContainer.parent().css('box-shadow', 'inset 0 0 0 1000px rgba(255,0,103,.2)');
				facet._barForeground.css('box-shadow', 'inset 0 0 0 1000px #ff0067');
			}

		},

		removeRowSelectionFromFacet(facet: any) {
			if (facet._histogram) {
				facet._histogram.bars.forEach(bar => {
					bar._element.css('fill', '');
					bar._element.removeClass('row-selected');
				});
			} else if (facet._sparkline) {
				// TODO: sparkline
			} else {
				facet._barForeground.css('box-shadow', '');
				facet._sparklineContainer.parent().css('box-shadow', '');
			}
		},

		injectSelectedRowIntoGroup(group: any, selection: RowSelection) {

			// clear existing selections
			for (const facet of group.facets) {

				// ignore placeholder facets
				if (facet._type === 'placeholder') {
					continue;
				}

				this.removeRowSelectionFromFacet(facet);
			}

			// if no selection, exit early
			if (!selection || selection.d3mIndices.length === 0) {
				return;
			}

			const rows = getSelectedRows(selection);
			rows.forEach(row => {

				// get col
				const col = _.find(row.cols, c => {
					return c.key === this.groupSpec.colName;
				});

				// no matching col, exit early
				if (!col) {
					return;
				}

				for (const facet of group.facets) {

					// ignore placeholder facets
					if (facet._type === 'placeholder') {
						continue;
					}

					if (facet._histogram) {

						this.addRowSelectionToFacet(facet, col);

					} else if (facet._sparkline) {

						// TODO: sparkline

					} else {

						const type = getVarType(facet.key);

						if (isClusterType(type)) {
							const clusterCol = this.findClusterCol(facet.key, row);
							if (facet.value === clusterCol.value) {
								this.addRowSelectionToFacet(facet, col);
							}
							continue;
						}

						if (isFeatureType(type)) {
							const featureCol = this.findFeatureCol(facet.key, row);
							const features = featureCol.value.split(',');
							if (_.includes(features, facet.value)) {
								this.addRowSelectionToFacet(facet, col);
							}
							continue;
						}

						if (facet.value === col.value) {
							this.addRowSelectionToFacet(facet, col);
						}

					}
				}
			});
		},

		injectHighlightDatasetDeemphasis(group: any, highlight: Highlight) {

			// if the dataset of the highlight does not match the dataset of this
			// facet, deemphasis the group

			if (!highlight || !highlight || highlight.dataset === group.dataset) {
				group._element.removeClass('deemphasis');
				return;
			}
			if (highlight.dataset !== group.dataset) {
				group._element.addClass('deemphasis');
			}
		},

		injectHighlightsIntoGroup(group: any, highlight: Highlight) {

			if (this.ignoreHighlights) {
				return;
			}

			// loop through groups ensure that selection is clear on each
			group.facets.forEach(facet => {
				if (facet._type === 'placeholder') {
					return;
				}
				if (facet._histogram) {

					facet.deselect();

				} else if (facet._sparkline) {

					// TODO: sparkline

				} else {
					this.selectCategoricalFacet(facet);
				}
			});


			const highlightRootValue = this.getHighlightValue(highlight);
			const highlightSummary = this.groupSpec.summary ? this.groupSpec.summary.filtered : null;

			for (const facet of group.facets) {

				// ignore placeholder facets
				if (facet._type === 'placeholder') {
					continue;
				}

				if (facet._histogram) {

					const selection = {} as any;

					// if this is the highlighted group, create filter selection
					if (this.isHighlightedGroup(highlight, this.groupSpec.colName)) {

						// NOTE: the `from` / `to` values MUST be strings.
						selection.range = {
							from: `${highlightRootValue.from}`,
							to: `${highlightRootValue.to}`
						};

					} else {

						const bars = facet._histogram.bars;

						if (highlightSummary && highlightSummary.buckets.length === bars.length) {

							const slices = {};

							highlightSummary.buckets.forEach((bucket, index) => {
								const entry: any = _.last(bars[index].metadata);
								slices[entry.label] = bucket.count;
							});

							selection.slices = slices;
						}
					}

					facet.select({
						selection: selection
					});

				} else if (facet._sparkline) {

					const selection = {} as any;

					// if this is the highlighted group, create filter selection
					if (this.isHighlightedGroup(highlight, this.groupSpec.colName)) {

						// NOTE: the `from` / `to` values MUST be strings.
						selection.range = {
							from: `${highlightRootValue.from}`,
							to: `${highlightRootValue.to}`
						};

					} else {

						// TODO: impl highlighting

						// const points = facet._sparkline.points;
						// if (highlightSummary && highlightSummary.buckets.length === points.length) {
						//
						// 	const slices = {};
						//
						// 	highlightSummary.buckets.forEach((bucket, index) => {
						// 		const entry: any = _.last(points[index].metadata);
						// 		slices[entry.label] = bucket.count;
						// 	});
						//
						// 	selection.slices = slices;
						// }
					}

					facet.select({
						selection: selection
					});

				} else {

					if (this.isHighlightedGroup(highlight, this.groupSpec.colName)) {

						const highlightValue = this.getHighlightValue(highlight);
						if (highlightValue.toLowerCase() === facet.value.toLowerCase()) {
							this.selectCategoricalFacet(facet);
							this.selectTimeseriesFacet(facet);
						} else {
							this.deselectCategoricalFacet(facet);
							this.deselectTimeseriesFacet(facet);
						}

					} else {

						if (highlightSummary) {

							const bucket = _.find(highlightSummary.buckets, b => {
								return b.key === facet.value;
							});

							if (bucket && bucket.count > 0) {
								this.selectCategoricalFacet(facet, bucket.count);
							} else {
								this.deselectCategoricalFacet(facet);
							}

						}
					}
				}
			}
		},

		injectHighlights(highlight: Highlight, selection: RowSelection, deemphasis: any) {
			// Clear highlight state incase it was set via a click on on another
			// component
			$(this.$el).find('.select-highlight').removeClass('select-highlight');
			// Update highlight
			const group = this.facets.getGroup(this.groupSpec.key);
			if (!group) {
				return;
			}
			this.injectHighlightsIntoGroup(group, highlight);
			this.injectHighlightDatasetDeemphasis(group, highlight);
			this.injectSelectedRowIntoGroup(group, selection);
		},

		augmentGroup(distilGroup: Group, facetsGroup: any) {
			// inject any custom properties required for the distil app
			facetsGroup.dataset = distilGroup.dataset;
			facetsGroup.facets.forEach(f => {
				f.dataset = distilGroup.dataset;
			});
		},

		groupsEqual(a: Group, b: Group): boolean {
			const OMITTED_FIELDS = ['selection', 'selected'];
			// NOTE: we dont need to check key, we assume its already equal
			if (a.label !== b.label) {
				return false;
			}
			if (a.facets.length !== b.facets.length) {
				return false;
			}
			for (let i = 0; i < a.facets.length; i++) {
				if (!_.isEqual(
					_.omit(a.facets[i], OMITTED_FIELDS),
					_.omit(b.facets[i], OMITTED_FIELDS))) {
					return false;
				}
			}
			return true;
		},

		updateGroups(currGroup: Group, prevGroup: Group) {

			const groupSpec = _.cloneDeep(currGroup);

			// check if it already exists
			if (prevGroup) {
				// check if equal, if so, no need to change
				if (this.groupsEqual(currGroup, prevGroup)) {
					return;
				}
				// replace group if it is existing
				this.facets.replaceGroup(groupSpec);
				this.augmentGroup(currGroup, this.facets.getGroup(currGroup.key));
				this.injectHTML(currGroup, this.facets.getGroup(currGroup.key)._element);
				this.injectHighlightsIntoGroup(this.facets.getGroup(currGroup.key), this.highlight);
				this.injectHighlightDatasetDeemphasis(this.facets.getGroup(currGroup.key), this.highlight);
				this.injectSelectedRowIntoGroup(this.facets.getGroup(currGroup.key), this.rowSelection);
				this.injectDeemphasis(this.facets.getGroup(currGroup.key), this.deemphasis);
			} else {
				// add to appends
				this.facets.append(groupSpec);
				this.augmentGroup(currGroup, this.facets.getGroup(currGroup.key));
				this.injectHTML(currGroup, this.facets.getGroup(currGroup.key)._element);
				this.injectHighlightsIntoGroup(this.facets.getGroup(currGroup.key), this.highlight);
				this.injectHighlightDatasetDeemphasis(this.facets.getGroup(currGroup.key), this.highlight);
				this.injectSelectedRowIntoGroup(this.facets.getGroup(currGroup.key), this.rowSelection);
				this.injectDeemphasis(this.facets.getGroup(currGroup.key), this.deemphasis);
			}
			this.updateImportantBadge(currGroup);
		},

		updateImportantBadge(group: Group) {
			// update 'important' class
			const $group = this.facets.getGroup(group.key)._element;
			const isImportant = this.ranking > IMPORTANT_VARIABLE_RANKING_THRESHOLD;
			$group.toggleClass('important', Boolean(isImportant));
		},

		// inject type icon
		injectTypeIcon(group: Group, $elem: JQuery) {
			if (isCategoricalFacet(group.facets.length > 0 && group.facets[0])) {
				const facetSpecs = (<CategoricalFacet[]>group.facets);
				const typeicon = facetSpecs[0].icon.class;
				const $icon = $(`<i class="${typeicon}"></i>`);
				$elem.find('.group-header').append($icon);
			}
			if (hasComputedVarPrefix(group.colName)) {
				const $forkIcon = createIcon(IconFork);
				$elem.find('.group-header').append($forkIcon);
			}
		},


		// inject type headers
		injectTypeChangeHeaders(group: Group, $elem: JQuery) {
			if (this.enableTypeChange) {
				const facetId = `${group.dataset}:${group.colName}`;
				// if we have a menu for this already, destroy it to replace it
				if (this.menus[facetId]) {
					this.menus[facetId].$destroy();
				}
				const $slot = $('<span/>');
				$elem.find('.group-header').append($slot);
				const menu = new TypeChangeMenu(
					{
						store: this.$store,
						router: this.$router,
						propsData: {
							dataset: group.dataset,
							field: group.colName
						}
					});
				menu.$mount($slot[0]);
				this.menus[facetId] = menu;
			}
		},

		injectImagePreview(group: Group, $elem: JQuery) {
			if (group.type === 'image') {
				const $facets = $elem.find('.facet-block');
				group.facets.forEach((facet: any, index) => {
					const $facet = $($facets.get(index));
					const $slot = $('<span/>');
					$facet.append($slot);
					const preview = new ImagePreview(
						{
							store: this.$store,
							router: this.$router,
							propsData: {
								// NOTE: there seems to be an issue with the visibility plugin used
								// when injecting this way. Cancel the visibility flagging for facets.
								preventHiding: true,
								imageUrl: facet.file || facet.value
							}
						});
					preview.$mount($slot[0]);
				});
			}
		},
		injectImportantBadge(group: Group, $elem: JQuery) {
			const $groupFooter = $elem.find('.group-footer');
			const importantBadge = document.createElement('div');
			importantBadge.className += 'important-badge';
			const $bookMarkIcon = createIcon(IconBookmark);
			importantBadge.append($bookMarkIcon);
			$groupFooter.append(importantBadge);
		},
	},

	destroyed: function() {
		this.facets.destroy();
		this.facets = null;

		// specifically destroy menus so because we injected them
		// and so we have to take manual action to destroy them
		_.forIn(this.menus, menu => menu.$destroy());
		this.menus = null;
	},
});
</script>

<style>

.group-facet-container {
	position: relative;
}

.facet-highlight-spinner {
	position: absolute;
	right: 0;
	bottom: 0;
	margin-right: 8px;
	margin-bottom: 6px;
}

.facets-root {
	position: relative;
}

.facets-group-container.deemphasis {
	opacity: 0.5;
}
.facets-root.highlighting-enabled {
	padding-left: 32px;
}
.highlighting-enabled .facets-group {
	cursor: pointer !important;
	border: 1px solid rgba(0,0,0,0);
}
.highlighting-enabled .facets-group:hover {
	border: 1px solid #00c6e1;
}

.highlighting-enabled .group-header,
.highlighting-enabled .facet-range-controls,
.highlighting-enabled .facets-facet-horizontal {
	cursor: pointer !important;
}

.facets-facet-horizontal .facet-histogram-bar-transform {
	transition: none;
}

.facet-icon {
	display: none;
}
.facets-root-container {
	font-size: inherit;
}
.facets-facet-vertical .facet-label-count,
.facets-facet-vertical .facet-label {
	font-family: inherit;
	font-size: 0.733rem;
	color: rgba(0,0,0,.54);
}
.facets-group .group-header {
	font-family: inherit;
	font-size: 0.867rem;
	font-weight: bold;
	text-transform: uppercase;
	color: rgba(0,0,0,.54);
}
.facets-group .group-header i {
	margin-left: 5px;
}
.facets-group .group-header .svg-icon {
	height: 14px;
	margin-left: 2px;
}
/* .facets-group .group-footer {
	display: flex;
} */
.facets-group .group-footer .important-badge {
	align-self: center;
	padding-bottom: 5px;
	display: none;
}
.facets-group .group-facet-container {
    width: 100%;
	max-height: 240px;
	overflow-y: auto;
    overflow-x: hidden;
}
.facets-group-container.important .group-footer .important-badge {
	display: block;
}

.facets-facet-horizontal .select-highlight,
.facets-facet-horizontal .facet-histogram-bar-highlighted.select-highlight {
	fill: #007bff;
}

.excluded .facet-range-filter {
	box-shadow: inset 0 0 0 1000px rgba(220, 53, 0, 0.15) !important;
}
.excluded .facets-facet-horizontal .facet-histogram-bar-highlighted {
	fill: #333;
}
.excluded .facet-bar-base.facet-bar-selected {
	box-shadow: inset 0 0 0 1000px #333;
}
.facet-histogram {
	cursor: pointer !important;
}

.facets-facet-vertical.select-highlight .facet-bar-selected {
	box-shadow: inset 0 0 0 1000px #007bff;
}

.highlight-arrow {
	position: absolute;
	left: -28px;
	top: calc(50% - 14px);
	color: #00c6e1;
}
.facet-tooltip {
	position: absolute;
	padding: 4px 8px;
	border-radius: 4px;
	background-color: #333;
	color: #fff;
	z-index: 2;
	pointer-events: none;
}
.facet-tooltip::after {
	content: " ";
	position: absolute;
	top: 100%; /* At the bottom of the tooltip */
	left: 50%;
	margin-left: -5px;
	border-width: 5px;
	border-style: solid;
	border-color: #333 transparent transparent transparent;
}
.facet-sparkline,
.facet-sparkline-container {
	overflow: visible !important;
}
</style>
