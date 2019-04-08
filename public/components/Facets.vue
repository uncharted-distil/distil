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
import { Group, CategoricalFacet, isCategoricalFacet, getCategoricalChunkSize, isNumericalFacet } from '../util/facets';
import { Highlight, RowSelection, Row } from '../store/highlights/index';
import { VariableSummary } from '../store/dataset/index';
import { Dictionary } from '../util/dict';
import { getSelectedRows } from '../util/row';
import Facets from '@uncharted.software/stories-facets';
import ImagePreview from '../components/ImagePreview';
import TypeChangeMenu from '../components/TypeChangeMenu';
import { circleSpinnerHTML } from '../util/spinner';
import { getVarType, isClusterType, isFeatureType, addClusterPrefix, addFeaturePrefix, hasComputedVarPrefix } from '../util/types';

import '@uncharted.software/stories-facets/dist/facets.css';

const INJECT_DEBOUNCE = 200;

/*
In 1989 the japanese-american animated musical film `Little Nemo: Adventures in
Slumberland` was released in North America. In the film a young boy named Nemo
is summoned to Slumberland by King Morpheus to become his heir to the throne.
Morpheus gives Nemo a golden key that opens every door in the kingdom, and warns
him of a door with a dragon insignia that must never be opened as it contains
the dreaded Nightmare King.

This file is the door with the dragon insignia. The facets component is the
Nightmare King. We must keep its evil contained within this file and it must
never be allowed to escape lest the Kingdom of Slumberland fall to darkness.
*/

export default Vue.extend({
	name: 'facets',

	props: {
		groups: Array as () => Group[],
		highlights: Object as () => Highlight,
		rowSelection: Object as () => RowSelection,
		deemphasis: Object as () => any,
		enableTypeChange: Boolean as () => boolean,
		enableHighlighting: Boolean as () => boolean,
		ignoreHighlights: Boolean as () => boolean,
		html: [ String as () => string, Object as () => any, Function as () => Function ],
		instanceName: String as () => string,
		highlightArrows: Boolean as () => boolean,
		solutionId: String as () => string,
		sort: {
			default: (a: { key: string }, b: { key: string }) => {
				const textA = a.key.toLowerCase();
				const textB = b.key.toLowerCase();
				return (textA < textB) ? -1 : (textA > textB) ? 1 : 0;
			},
			type: Function as () => Function
		}
	},

	data() {
		const component = this as any;
		return {
			facets: {} as any,
			debouncedInjection: _.debounce((highlights: Highlight, selection: RowSelection, deemphasis: any) => {
				// we need to guard here because this debounced call can execute
				// after this component is destroyed
				if (!component.facets) {
					return;
				}
				component.injectHighlights(highlights, selection, deemphasis);
			}, INJECT_DEBOUNCE),
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
		this.facets = new Facets(this.$el, this.processedGroups);

		// Call customization hook
		this.processedGroups.forEach(group => {
			this.augmentGroup(group, this.facets.getGroup(group.key));
			this.injectHTML(group, this.facets.getGroup(group.key)._element);
		});

		// proxy events

		this.facets.on('facet-group:expand', (event: Event, key: string) => {
			component.$emit('expand', key);
		});

		this.facets.on('facet-group:collapse', (event: Event, key: string) => {
			component.$emit('collapse', key);
		});

		this.facets.on('facet-histogram:rangechangeduser', (event: Event, key: string, value: any, facet: any) => {
			const range = {
				from: _.toNumber(value.from.label[0]),
				to: _.toNumber(value.to.label[0])
			};
			component.$emit('range-change', this.instanceName, key, range, facet.dataset);
		});

		// hover over events

		this.facets.on('facet-histogram:mouseenter', (event: Event, key: string, value: any) => {
			component.$emit('histogram-mouse-enter', key, value);
		});

		this.facets.on('facet-histogram:mouseleave', (event: Event, key: string) => {
			component.$emit('histogram-mouse-leave', key);
		});

		this.facets.on('facet:mouseenter', (event: Event, key: string, value: number) => {
			component.$emit('facet-mouse-enter', key, value);
		});

		this.facets.on('facet:mouseleave', (event: Event, key: string) => {
			component.$emit('facet-mouse-leave', key);
		});

		// more events

		this.facets.on('facet-group:more', (event: Event, key: string) => {
			const group = _.find(this.groups, g => g.key === key);
			const chunkSize = getCategoricalChunkSize(group.type);
			if (!component.numToDisplay[key]) {
				Vue.set(component.numToDisplay, key, chunkSize);
				Vue.set(component.numAddedToDisplay, key, 0);
			}
			Vue.set(component.numToDisplay, key, component.numToDisplay[key] + chunkSize);
			Vue.set(component.numAddedToDisplay, key, component.numAddedToDisplay[key] + chunkSize);
			component.$emit('facet-more', key);
		});

		this.facets.on('facet-group:less', (event: Event, key: string) => {
			const group = _.find(this.groups, g => g.key === key);
			const chunkSize = getCategoricalChunkSize(group.type);
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
			if (this.isHighlightedValue(this.highlights, key, range)) {
				// clear current selection
				component.$emit('histogram-click', this.instanceName, null, null, facet.dataset);
			} else {
				// set selection
				component.$emit('histogram-click', this.instanceName, key, range, facet.dataset);
			}
		});

		this.facets.on('facet:click', (event: Event, key: string, value: string, count: number, facet: any) => {
			// User clicked on the value that is currently the highlight root
			if (this.isHighlightedValue(this.highlights, key, value)) {
				// clear current selection
				component.$emit('facet-click', this.instanceName, null, null, facet.dataset);
			} else {
				// set selection
				component.$emit('facet-click', this.instanceName, key, value, facet.dataset);
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
	},

	computed: {
		processedGroups(): Group[] {
			const groups = _.cloneDeep(this.groups);
			groups.forEach(group => {
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
			});
			return groups;
		}
	},

	watch: {

		// handle changes to the facet group list
		processedGroups(currGroups: Group[], prevGroups: Group[]) {
			// get map of all existing group keys in facets
			const prevMap: Dictionary<Group> = {};
			prevGroups.forEach(group => {
				prevMap[group.key] = group;
			});
			// update and groups
			const unchangedGroups = this.updateGroups(currGroups, prevMap);
			// for the unchanged, update collapse state
			this.updateCollapsed(unchangedGroups);
		},

		// handle external highlight changes by updating internal facet select states
		highlights(currHighlights: Highlight) {
			this.debouncedInjection(currHighlights, this.rowSelection, this.deemphasis);
			if (this.enableHighlighting) {
				this.addHighlightArrow(currHighlights);
			}
		},

		// handle external highlight changes by updating internal facet select states
		rowSelection(currSelection: RowSelection) {
			this.debouncedInjection(this.highlights, currSelection, this.deemphasis);
		},

		deemphasis(currDemphasis: any) {
			this.debouncedInjection(this.highlights, this.rowSelection, currDemphasis);
		},

		sort(currSort) {
			this.facets.sort(currSort);
		}
	},

	methods: {
		injectHTML(group: Group, $elem: JQuery) {
			const $groupFooter = $('<div class="group-footer"></div>').appendTo($elem.find('.facets-group'));
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
						this.$emit('numerical-click', this.instanceName, group.key, range, group.dataset);
					} else if (isCategoricalFacet(facet)) {
						this.$emit('categorical-click', this.instanceName, group.key, null, group.dataset);
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
						this.$emit('numerical-click', this.instanceName, group.key, range, group.dataset);
					}
				}
			});

			// inject type icon in group header
			this.injectTypeIcon(group, $elem);

			// inject type change header menus
			this.injectTypeChangeHeaders(group, $elem);

			// inject image preview if image type
			this.injectImagePreview(group, $elem);

			if (this.html) {
				if (_.isFunction(this.html)) {
					$groupFooter.append(this.html(group));
				} else {
					$groupFooter.append(this.html);
				}
			}
			this.injectImportantBadge(group, $elem);
		},

		addHighlightArrow(highlights: Highlight) {
			const $elem = $(this.$el);
			// remove previous
			$elem.find('.highlight-arrow').remove();

			// NOTE: first group is a query group, ignore it
			const QUERY_OFFSET = 1;


			const $groups = $elem.find('.facets-group');
			this.groups.forEach((group, index) => {
				// add highlight arrow
				if (this.isHighlightedGroup(highlights, group.key)) {
					const $group = $($groups.get(index + QUERY_OFFSET));
					$group.append('<div class="highlight-arrow"><i class="fa fa-arrow-circle-right fa-2x"></i></div>');
				}
			});
		},

		isHighlightedInstance(highlights: Highlight): boolean {
			return _.get(highlights, 'root.context') === this.instanceName;
		},

		isHighlightedGroup(highlights: Highlight, key: string): boolean {
			return this.isHighlightedInstance(highlights) &&
				_.get(highlights, 'root.key') === key;
		},

		isHighlightedValue(highlights: Highlight, key: string, value: any): boolean {
			// if not instance, return false
			if (!this.isHighlightedGroup(highlights, key)) {
				return false;
			}
			if (_.isArray(highlights.root.value)) {
				return false;
			}
			// if string, check for match
			if (_.isString(highlights.root.value)) {
				return highlights.root.value === value;
			}
			// otherwise, check range
			return highlights.root.value.from === value.from &&
				highlights.root.value.to === value.to;
		},

		getHighlightRootValue(highlights: Highlight): any {
			if (highlights.root && highlights.root.value) {
				return highlights.root.value;
			}
			return null;
		},

		getHighlightSummaries(highlights: Highlight): VariableSummary[] {
			if (highlights.values) {
				return highlights.values.summaries;
			}
			return null;
		},

		getHighlightSummary(highlights: Highlight, key: string): VariableSummary {
			const highlightSummaries = this.getHighlightSummaries(highlights);
			return _.find(highlightSummaries, s => {
				return s.key === key;
			});
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
				if (!facet._histogram) {
					continue;
				}

				facet._histogram.bars.forEach(bar => {
					if (!bar._element.hasClass('row-selected'))  {
						// NOTE: don't trample row selections
						bar._element.css('fill', '');
					}
				});
			}

			// de-emphasis within range
			for (const facet of group.facets) {

				// only emphasize histograms
				if (!facet._histogram) {
					continue;
				}

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
					return c.key === group.key;
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

		removeSpinnerFromGroup(group: any) {
			group._element.find('.facet-highlight-spinner').remove();
		},

		addSpinnerForGroup(group: any) {
			this.removeSpinnerFromGroup(group);
			const $spinner = $(`<div class="facet-highlight-spinner">${circleSpinnerHTML()}</div>`);
			group._element.find('.facets-group').append($spinner);
		},

		injectHighlightDatasetDeemphasis(group: any, highlights: Highlight) {

			// if the dataset of the highlight does not match the dataset of this
			// facet, deemphasis the group

			if (!highlights || !highlights.root || highlights.root.dataset === group.dataset) {
				group._element.removeClass('deemphasis');
				return;
			}
			if (highlights.root.dataset !== group.dataset) {
				group._element.addClass('deemphasis');
			}
		},

		injectHighlightsIntoGroup(group: any, highlights: Highlight) {

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
				} else {
					this.selectCategoricalFacet(facet);
				}
			});

			const highlightRootValue = this.getHighlightRootValue(highlights);
			const highlightSummary = this.getHighlightSummary(highlights, group.key);

			for (const facet of group.facets) {

				// ignore placeholder facets
				if (facet._type === 'placeholder') {
					continue;
				}

				if (facet._histogram) {

					const selection = {} as any;

					// if this is the highlighted group, create filter selection
					if (this.isHighlightedGroup(highlights, group.key)) {

						// NOTE: the `from` / `to` values MUST be strings.
						selection.range = {
							from: `${highlightRootValue.from}`,
							to: `${highlightRootValue.to}`
						};

					} else {

						const bars = facet._histogram.bars;

						if (highlightSummary && highlightSummary.buckets.length === bars.length) {
							this.removeSpinnerFromGroup(group);

							const slices = {};

							highlightSummary.buckets.forEach((bucket, index) => {
								const entry: any = _.last(bars[index].metadata);
								slices[entry.label] = bucket.count;
							});

							selection.slices = slices;
						} else {
							if (highlightRootValue) {
								this.addSpinnerForGroup(group);
							}
						}
					}


					facet.select({
						selection: selection
					});

				} else {

					if (this.isHighlightedGroup(highlights, group.key)) {

						const highlightValue = this.getHighlightRootValue(highlights);
						if (highlightValue.toLowerCase() === facet.value.toLowerCase()) {
							this.selectCategoricalFacet(facet);
							this.selectTimeseriesFacet(facet);
						} else {
							this.deselectCategoricalFacet(facet);
							this.deselectTimeseriesFacet(facet);
						}

					} else {

						if (highlightSummary) {
							this.removeSpinnerFromGroup(group);

							const bucket = _.find(highlightSummary.buckets, b => {
								return b.key === facet.value;
							});

							if (bucket && bucket.count > 0) {
								this.selectCategoricalFacet(facet, bucket.count);
							} else {
								this.deselectCategoricalFacet(facet);
							}

						} else {
							if (highlightRootValue) {
								this.addSpinnerForGroup(group);
							}
						}
					}
				}
			}
		},

		injectHighlights(highlights: Highlight, selection: RowSelection, deemphasis: any) {
			// Clear highlight state incase it was set via a click on on another
			// component
			$(this.$el).find('.select-highlight').removeClass('select-highlight');
			/// Update highlights
			this.processedGroups.forEach(g => {
				const group = this.facets.getGroup(g.key);
				if (!group) {
					return;
				}
				this.injectHighlightsIntoGroup(group, highlights);
				this.injectHighlightDatasetDeemphasis(group, highlights);
				this.injectSelectedRowIntoGroup(group, selection);
				this.injectDeemphasis(group, deemphasis);
			});
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

		updateGroups(currGroups: Group[], prevGroups: Dictionary<Group>): Group[] {
			const toAdd: Group[] = [];
			const unchanged: Group[] = [];
			// get map of all current, to track which groups need to be removed
			const toRemove: Dictionary<boolean> = {};
			_.forIn(prevGroups, group => {
				toRemove[group.key] = true;
			});
			// for each new group
			currGroups.forEach(group => {
				const old = prevGroups[group.key];
				// check if it already exists
				if (old) {
					// remove from existing so we can track which groups to remove
					toRemove[group.key] = false;
					// check if equal, if so, no need to change
					if (this.groupsEqual(group, old)) {
						// add to unchanged
						unchanged.push(group);
						return;
					}
					// replace group if it is existing
					this.facets.replaceGroup(_.cloneDeep(group));
					this.augmentGroup(group, this.facets.getGroup(group.key));
					this.injectHTML(group, this.facets.getGroup(group.key)._element);
					this.injectHighlightsIntoGroup(this.facets.getGroup(group.key), this.highlights);
					this.injectHighlightDatasetDeemphasis(this.facets.getGroup(group.key), this.highlights);
					this.injectSelectedRowIntoGroup(this.facets.getGroup(group.key), this.rowSelection);
					this.injectDeemphasis(this.facets.getGroup(group.key), this.deemphasis);
				} else {
					// add to appends
					toAdd.push(_.cloneDeep(group));
				}
			});
			// remove any old
			_.forIn(toRemove, (remove, key) => {
				if (remove) {
					this.facets.removeGroup(key);
				}
			});
			if (toAdd.length > 0) {
				// append groups
				this.facets.append(toAdd);
				toAdd.forEach(groupSpec => {
					this.augmentGroup(groupSpec, this.facets.getGroup(groupSpec.key));
					this.injectHTML(groupSpec, this.facets.getGroup(groupSpec.key)._element);
					this.injectHighlightsIntoGroup(this.facets.getGroup(groupSpec.key), this.highlights);
					this.injectHighlightDatasetDeemphasis(this.facets.getGroup(groupSpec.key), this.highlights);
					this.injectSelectedRowIntoGroup(this.facets.getGroup(groupSpec.key), this.rowSelection);
					this.injectDeemphasis(this.facets.getGroup(groupSpec.key), this.deemphasis);
				});
			}
			// sort alphabetically
			this.facets.sort(this.sort);

			// update 'important' class
			currGroups.forEach((group: Group) => {
				const $group = this.facets.getGroup(group.key)._element;
				$group.toggleClass('important', Boolean(group.isImportant));
			});

			// return unchanged groups
			return unchanged;
		},

		updateCollapsed(unchangedGroups) {
			unchangedGroups.forEach(group => {
				// get the existing facet
				const existing = this.facets.getGroup(group.key);
				if (existing) {
					if (existing.collapsed !== group.collapsed) {
						existing.collapsed = group.collapsed;
					}
				}
			});
		},

		// inject type icon
		injectTypeIcon(group: Group, $elem: JQuery) {
			if (isCategoricalFacet(group.facets.length > 0 && group.facets[0])) {
				const facetSpecs = (<CategoricalFacet[]>group.facets);
				const typeicon = facetSpecs[0].icon.class;
				const $icon = $(`<i class="${typeicon}"></i>`);
				$elem.find('.group-header').append($icon);
			}
			if (hasComputedVarPrefix(group.key)) {
				const $forkIcon = createIcon(IconFork);
				$elem.find('.group-header').append($forkIcon);
			}
		},

		getGroupSampleValues(group: Group): any[] {
			let values = [];
			group.facets.forEach(facet => {
				if (isNumericalFacet(facet)) {
					values = facet.histogram.slices.slice(0, 10).map(b => _.toNumber(b.label));
				} else if (isCategoricalFacet(facet)) {
					values.push(facet.value);
				}
			});
			return values.filter(v => v !== undefined);
		},

		// inject type headers
		injectTypeChangeHeaders(group: Group, $elem: JQuery) {
			if (this.enableTypeChange) {
				const $slot = $('<span/>');
				$elem.find('.group-header').append($slot);
				const menu = new TypeChangeMenu(
					{
						store: this.$store,
						propsData: {
							dataset: group.dataset,
							field: group.key,
							values: this.getGroupSampleValues(group)
						}
					});
				menu.$mount($slot[0]);
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
.facets-group .group-footer {
	display: flex;
}
.facets-group .group-footer .important-badge {
	align-self: center;
	padding-bottom: 5px;
	display: none;
}
.facets-group-container.important .group-footer .important-badge {
	display: block;
}

.facets-facet-horizontal .select-highlight,
.facets-facet-horizontal .facet-histogram-bar-highlighted.select-highlight {
	fill: #007bff;
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
