<template>
	<div class="facets" v-once ref="facets"></div>
</template>

<script lang="ts">
import _ from 'lodash';
import $ from 'jquery';
import Vue from 'vue';
import { Group, CategoricalFacet, isCategoricalFacet } from '../util/facets';
import { Dictionary } from '../util/dict';
import Facets from '@uncharted.software/stories-facets';
import TypeChangeMenu from '../components/TypeChangeMenu';
import '@uncharted.software/stories-facets/dist/facets.css';
import { Highlights, Range } from '../util/highlights';
import { CATEGORICAL_FILTER, NUMERICAL_FILTER } from '../util/filters';

export default Vue.extend({
	name: 'facets',

	props: {
		groups: Array,
		filters: Array,
		highlights: Object, // ValueHighlights
		typeChange: Boolean,
		html: [ String, Object, Function ],
		sort: {
			default: (a: { key: string }, b: { key: string }) => {
				const textA = a.key.toLowerCase();
				const textB = b.key.toLowerCase();
				return (textA < textB) ? -1 : (textA > textB) ? 1 : 0;
			},
			type: Function
		}
	},

	data() {
		return {
			facets: <any>{},
			instanceName: _.uniqueId('facet-')
		};
	},

	mounted() {
		const component = this;

		// Instantiate the external facets widget. The facets maintain their own copies
		// of group objects which are replaced wholesale on changes.  Elsewhere in the code
		// we modify local copies of the group objects, then replace those in the Facet component
		// with copies.
		this.facets = new Facets(this.$el, this.groups.map(group => {
			return _.cloneDeep(group);
		}));

		// Call customization hook
		this.groups.forEach(group => {
			this.injectHTML(group, this.facets.getGroup(group.key)._element);
		});

		// proxy events

		this.facets.on('facet-group:expand', (event: Event, key: string) => {
			component.$emit('expand', key);
		});

		this.facets.on('facet-group:collapse', (event: Event, key: string) => {
			component.$emit('collapse', key);
		});

		this.facets.on('facet-histogram:rangechangeduser', (event: Event, key: string, value: any) => {
			component.$emit('range-change', key, value);
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

		// click events

		this.facets.on('facet-histogram:click', (event: Event, key: string, value: any) => {
			// if this is a click on value previously used as highlight root, clear
			const range = {
				from: _.toNumber(value.label),
				to: _.toNumber(value.toLabel)
			};
			if (this.isHighlightedValue(this.highlights, key, range)) {
				// clear current selection
				component.$emit('histogram-click', this.instanceName);
			} else {
				// set selection
				component.$emit('histogram-click', this.instanceName, key, range);
			}
		});

		this.facets.on('facet:click', (event: Event, key: string, value: string) => {
			// User clicked on the value that is currently the highlight root
			if (this.isHighlightedValue(this.highlights, key, value)) {
				// clear current selection
				component.$emit('facet-click', this.instanceName);
			} else {
				// set selection
				component.$emit('facet-click', this.instanceName, key, value);
			}
		});
	},

	computed: {
		facetFiltersByKey(): Dictionary<string[]> {
			const m = {};
			this.filters.forEach(filter => {
				if (filter.enabled && filter.type === CATEGORICAL_FILTER) {
					const categories = {};
					filter.categories.forEach(category => {
						categories[category] = true;
					});
					m[filter.name] = categories;
				}
			});
			return m;
		},
		histogramFiltersByKey(): Dictionary<Range> {
			const m = {};
			this.filters.forEach(filter => {
				if (filter.enabled && filter.type === NUMERICAL_FILTER) {
					m[filter.name] = {
						from: filter.min,
						to: filter.max
					};
				}
			});
			return m;
		}
	},

	watch: {
		// handle changes to the facet group list
		groups(currGroups: Group[], prevGroups: Group[]) {
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
		highlights(currHighlights: Highlights) {
			this.injectHighlights(currHighlights);
		},

		sort(currSort) {
			this.facets.sort(currSort);
		}
	},

	methods: {
		isFiltered(key, value): boolean {
			return this.facetFiltersByKey[key] ? !this.facetFiltersByKey[key][value] : false;
		},

		injectHTML(group: Group, $elem: JQuery) {

			$elem.click(() => {
				this.$emit('click', group.key);
			});

			// inject type icon in group header
			this.injectTypeIcon(group, $elem);

			// inject type change header menus
			this.injectTypeChangeHeaders(group, $elem);

			// inject category toggle buttons
			this.injectCategoryToggleButtons(group, $elem);

			if (!this.html) {
				return;
			}
			const $group = $elem.find('.facets-group');
			if (_.isFunction(this.html)) {
				$group.append(this.html(group));
			} else {
				$group.append(this.html);
			}
		},

		isHighlightedInstance(highlights: Highlights): boolean {
			return _.get(highlights, 'root.context') === this.instanceName;
		},

		isHighlightedGroup(highlights: Highlights, key: string): boolean {
			return this.isHighlightedInstance(highlights) &&
				_.get(highlights, 'root.key') === key;
		},

		isHighlightedValue(highlights: any, key: string, value: any): boolean {
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

		isValueInBar(bar: any, value: number): boolean {
			const metadata: any[] = bar.metadata;
			const barMin = _.toNumber(_.first(metadata).label);
			const barMax = _.toNumber(_.last(metadata).toLabel);
			const num = _.toNumber(value);
			return (num >= barMin && num < barMax);
		},

		getHighlightRootValue(highlights: Highlights): any {
			if (highlights.root) {
				if (highlights.root.value) {
					if (_.isArray(highlights.root.value)) {
						return null;
					}
					if (_.isString(highlights.root.value)) {
						return highlights.root.value;
					}
					// take middle value
					return _.toNumber((highlights.root.value.to + highlights.root.value.from) / 2);
				}
			}
			return null;
		},

		getHighlightValuesForGroup(highlights: Highlights, key: string): any[] {
			if (highlights.values) {
				return highlights.values[key] ? highlights.values[key] : [];
			}
			return null;
		},

		getGroupNumRows(key: string): number {
			const groups = this.groups.filter(g => {
				return g.key === key;
			});
			return groups.length > 0 ? groups[0].numRows : 0;
		},

		injectSelectionIntoGroup(group: any, highlights: Highlights) {

			if (!this.isHighlightedGroup(highlights, group.key)) {
				return;
			}

			const highlightValue = this.getHighlightRootValue(highlights);

			for (const facet of group.facets) {

				// ignore placeholder facets
				if (facet._type === 'placeholder') {
					continue;
				}

				if (facet._histogram) {

					const bars = facet._histogram.bars;
					for (let i = 0; i < bars.length; i++) {
						const bar = bars[i];
						if (this.isValueInBar(bar, highlightValue)) {
							$(bar._element).addClass('select-highlight');
							return;
						}
					}

				} else {

					if (highlightValue.toLowerCase() === facet.value.toLowerCase()) {
						$(facet._element).addClass('select-highlight');
					}

				}
			}
		},

		selectCategoricalFacet(facet: any, count?: number) {
			if (!count && facet._spec.segments && facet._spec.segments.length > 0) {
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

		getSampleScale(numRows: number ): number {
			const NUM_SAMPLES = 100;
			return 1 / (NUM_SAMPLES / numRows);
		},

		scaleSlicesBySampleSize(slices: Dictionary<number>, numRows: number, bars: any) {
			const count = {};
			for (let i = 0; i < bars.length; i++) {
				const bar = bars[i];
				const entry: any = _.last(bar.metadata);
				count[entry.label] = entry.count;
			}
			_.forIn(slices, (slice, key) => {
				slices[key] = Math.min(count[key], slice * this.getSampleScale(numRows));
			});
		},

		scaleCountBySampleSize(count: number, numRows: number, facet: any) {
			return Math.min(facet.count, count * this.getSampleScale(numRows));
		},

		ensureMinHeight(slices: Dictionary<number>, bars: any) {
			const MIN_PERCENT = 0.1;
			// get counts per entry, and max of all
			const count = {};
			let max = 0;
			for (let i = 0; i < bars.length; i++) {
				const bar = bars[i];
				const entry: any = _.last(bar.metadata);
				count[entry.label] = entry.count;
				max = Math.max(max, entry.count);
			}
			// set count to ensure min height
			const minCount = MIN_PERCENT * max;
			_.forIn(slices, (slice, key) => {
				if (slice < minCount) {
					slices[key] = Math.min(count[key], minCount);
				}
			});
		},

		injectHighlightsIntoGroup(group: any, highlights: Highlights) {

			const highlightValues = this.getHighlightValuesForGroup(highlights, group.key);
			const filter = this.histogramFiltersByKey[group.key];

			// loop through groups ensure that selection is clear on each
			group.facets.forEach(facet => {
				if (facet._type === 'placeholder') {
					return;
				}
				if (facet._histogram) {
					facet.deselect();
				}
				this.selectCategoricalFacet(facet);
			});

			for (const facet of group.facets) {

				// ignore placeholder facets
				if (facet._type === 'placeholder') {
					continue;
				}

				if (facet._histogram) {
					// Build up the selection structures ||  to pass to the facets lib.  The facets library doesn't
					// give us a good way to determine the index of a particular numeric value in the set of generated
					// bars (they are non-contiguous), so we just have to check each range ourselves.  To be more efficient
					// we sort the values and do it one pass.

					let slices: Dictionary<number> = null;

					if (highlightValues) {

						const values = Array.from(highlightValues) as number[];

						// if we have values, set the slices object so we can filter by the values.
						slices = {};

						const bars = facet._histogram.bars;

						// if this is the root highlight, highlight only the selected bar
						if (this.isHighlightedGroup(highlights, group.key)) {

							const highlightValue = this.getHighlightRootValue(highlights);
							for (let i = 0; i < bars.length; i++) {
								const bar = bars[i];
								if (this.isValueInBar(bar, highlightValue)) {
									const entry: any = _.last(bar.metadata);
									slices[entry.label] = entry.count;
									break;
								}
							}

						} else {

							// otherwise go through all values in highlights
							const sortedValues: number[] = values.sort((a, b) => a - b) as number[];

							let lastIndex = 0;
							sortedValues.forEach(value => {
								// iterate over the facet bars and find the one that contains the current value
								for (let i = lastIndex; i < bars.length; i++) {
									const bar = bars[i];
									if (this.isValueInBar(bar, value)) {
										const entry: any = _.last(bar.metadata);
										if (!slices[entry.label]) {
											slices[entry.label] = 0;
										}
										slices[entry.label]++;
										lastIndex = i;
										break;
									}
								}
							});

							this.scaleSlicesBySampleSize(slices, this.getGroupNumRows(group.key), bars);

							// ensure min height
							this.ensureMinHeight(slices, bars);
						}

					}

					// create selection
					const selection: any = {};

					if (filter) {
						// NOTE: the `from` / `to` values MUST be strings.
						selection.range = {
							from: `${filter.from}`,
							to: `${filter.to}`
						};
					}

					if (slices) {
						selection.slices = slices;
					}

					if (slices || filter) {
						facet.select({
							selection: selection
						});
					}

				} else {

					if (highlightValues) {

						if (this.isHighlightedGroup(highlights, group.key)) {
							const highlightValue = this.getHighlightRootValue(highlights);
							if (highlightValue.toLowerCase() === facet.value.toLowerCase()) {
								this.selectCategoricalFacet(facet);
							} else {
								this.deselectCategoricalFacet(facet);
							}

						} else {

							const values = Array.from(highlightValues) as string[];
							const matches = _.filter(values, v => v.toLowerCase() === (facet.value.toLowerCase ? facet.value.toLowerCase() : undefined));

							if (matches.length > 0) {
								const count = this.scaleCountBySampleSize(matches.length, this.getGroupNumRows(group.key), facet);
								this.selectCategoricalFacet(facet, count);
							} else {
								this.deselectCategoricalFacet(facet);
							}
						}

					} else {

						const highlightValue = this.getHighlightRootValue(highlights);
						if (highlightValue) {

							if (this.isHighlightedGroup(highlights, group.key) &&
								highlightValue.toLowerCase() === facet.value.toLowerCase()) {
								this.selectCategoricalFacet(facet);
							} else {
								this.deselectCategoricalFacet(facet);
							}
						}

					}
				}
			}
		},

		injectHighlights(highlights: Highlights) {
			// Clear highlight state incase it was set via a click on on another
			// component
			$(this.$el).find('.select-highlight').removeClass('select-highlight');
			/// Update highlights
			this.groups.forEach(g => {
				const group = this.facets.getGroup(g.key);
				if (!group) {
					return;
				}
				this.injectHighlightsIntoGroup(group, highlights);
				this.injectSelectionIntoGroup(group, highlights);
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
			for (let i=0; i<a.facets.length; i++) {
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
					this.injectHTML(group, this.facets.getGroup(group.key)._element);
					this.injectHighlightsIntoGroup(this.facets.getGroup(group.key), this.highlights);
					this.injectSelectionIntoGroup(this.facets.getGroup(group.key), this.highlights);
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
					this.injectHTML(groupSpec, this.facets.getGroup(groupSpec.key)._element);
					this.injectHighlightsIntoGroup(this.facets.getGroup(groupSpec.key), this.highlights);
					this.injectSelectionIntoGroup(this.facets.getGroup(groupSpec.key), this.highlights);
				});
			}
			// sort alphabetically
			this.facets.sort(this.sort);
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
			if (isCategoricalFacet(group.facets[0])) {
				const facetSpecs = (<CategoricalFacet[]>group.facets);
				const typeicon = facetSpecs[0].icon.class;
				const $icon = $(`<i class="${typeicon}"></i>`);
				$elem.find('.group-header').append($icon);
			}
		},

		// inject type headers
		injectTypeChangeHeaders(group: Group, $elem: JQuery) {
			if (this.typeChange) {
				const $slot = $('<span/>');
				$elem.find('.group-header').append($slot);
				const menu = new TypeChangeMenu(
					{
						store: this.$store,
						propsData: {
							field: group.key
						}
					});
				menu.$mount($slot[0]);
			}
		},

		// inject category filter buttons
		injectCategoryToggleButtons(groupSpec: Group, $elem: JQuery) {
			if (!isCategoricalFacet(groupSpec.facets[0])) {
				return;
			}

			// find the facet nodes in the DOM
			const $verticalFacets = $elem.find('.facets-facet-vertical');

			// Add a clickable filter state button to each facet.
			for (const facetElement of $verticalFacets) {

				const $facet = $(facetElement).find('.facet-query-close');
				const label = $facet.parent().find('.facet-label').text().trim();
				const facetSpec = (<CategoricalFacet[]>groupSpec.facets).find(f => f.value === label);

				// only add controls for filterable facets
				if (!facetSpec.filterable) {
					continue;
				}

				// setup based on the initial filter state
				const key = groupSpec.key;
				const value = facetSpec.value;

				let $icon = null;

				if (this.isFiltered(key, value)) {
					$icon = $(`<i id=${key}-${value} class="fa fa-circle-o"></i>`);
				} else {
					$icon = $(`<i id=${key}-${value} class="fa fa-circle"></i>`);
				}
				$icon.appendTo($facet);


				$icon.click(e => {
					// get group and current facet
					const group = this.facets.getGroup(key);
					const current = <any>(<CategoricalFacet[]>group.facets).find(facet => facet.value === value);

					// selected values
					const values = [];

					// toggle the facet filter state
					if (!this.isFiltered(key, value)) {
						// switch to unfilter from filtered
						$icon.removeClass('fa-circle').addClass('fa-circle-o');
						// add newly selected value
						this.deselectCategoricalFacet(current);
					} else {
						// switch from filtered to unfiltered, and restore highlight state if needed
						$icon.removeClass('fa-circle-o').addClass('fa-circle');
						if (this.isHighlightedValue(this.highlights, key, value)) {
							this.selectCategoricalFacet(current);
						}
						values.push(value);
					}

					// add all currently selected values
					const selected = group.facets
						.filter(f => !this.isFiltered(f.key, f.value) && value !== f.value)
						.map(f => f.value);

					this.$emit('facet-toggle', key, values.concat(selected));
				});

				$icon.mouseenter(e => {
					$icon.addClass('text-primary');
				});

				$icon.mouseleave(e => {
					$icon.removeClass('text-primary');
				});
			}
		}
	},

	destroyed: function() {
		this.facets.destroy();
		this.facets = null;
	},
});
</script>

<style>
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
.facets-facet-horizontal .select-highlight,
.facets-facet-horizontal .facet-histogram-bar-highlighted.select-highlight {
	fill: #007bff;
}

.facets-facet-vertical.select-highlight .facet-bar-selected {
	box-shadow: inset 0 0 0 1000px #007bff;
}
</style>
