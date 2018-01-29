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
import Multimap from 'multimap';
import { Highlights } from '../util/highlights';

export default Vue.extend({
	name: 'facets',

	props: {
		groups: Array,
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
			facetFilteredValues: new Multimap(),
			instanceName: _.uniqueId('facet-')
		};
	},

	mounted() {
		const component = this;
		const groups = () => <Group[]>this.groups;

		// Instantiate the external facets widget.  The facets maintain their own copies
		// of group objects which are replaced wholesale on changes.  Elsewhere in the code
		// we modify local copies of the group objects, then replace those in the Facet component
		// with copies.
		this.facets = new Facets(this.$el, groups().map(group => {
			return _.cloneDeep(group);
		}));

		// Call customization hook
		groups().forEach(group => {
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
		injectHTML(group: Group, $elem: JQuery) {

			$elem.click(() => {
				this.$emit('click', group.key);
			});

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

		isHighlightedValue(highlights: Highlights, key: string, value: string | Range): boolean {
			// if not instance, return false
			if (!this.isHighlightedGroup(highlights, key)) {
				return false;
			}
			// if string, check for match
			if (_.isString(value)) {
				return highlights.root.value === value;
			}
			// otherwise, check range
			return highlights.root.value.from === value.from &&
				highlights.root.value.to === value.to;
		},

		injectHighlightsIntoGroup(group: any, highlights: Highlights) {

			// loop through groups ensure that selection is clear on each - not that clear
			// the selection on a categorical facet means set its selection to a full count
			if (_.isEmpty(highlights.values) || !highlights.values[group.key]) {
				group.facets.forEach(facet => {
					if (facet._histogram) {
						facet.deselect();
					} else {
						facet.select(facet.count);
					}
				});
				return;
			}

			const values = highlights.values[group.key];
			for (const facet of group.facets) {
				if (facet._type === 'placeholder') {
					continue;
				}
				if (facet._histogram) {
					// Build up the selection structure to pass to the facets lib.  The facets library doesn't
					// give us a good way to determine the index of a particular numeric value in the set of generated
					// bars (they are non-contiguous), so we just have to check each range ourselves.  To be more efficient
					// we sort the values and do it one pass.
					const sortedValues = Array.from(values).sort((a, b) => <any>a - <any>b);
					const slices: Dictionary<number> = {};
					let lastLabelIdx = 0;
					for (const value of sortedValues) {
						// iterate over the facet bars and find the one that contains the current value
						for (let i = lastLabelIdx; i < facet._histogram.bars.length; i++) {
							const metadata: any[] = facet._histogram.bars[i].metadata;
							const numValue = _.toNumber(value);

							const barMin = _.toNumber(_.first(metadata).label);
							const barMax = _.toNumber(_.last(metadata).toLabel);

							// If the current bar contains the selected value, add it to the slices map so that
							// it gets added to the selection.
							if (numValue >= barMin && numValue <= barMax) {
								// If the current bar is the flagged highlight, make sure we have the selection
								// tag set.
								if (this.isHighlightedGroup(highlights, group.key)) {
									$(facet._histogram.bars[i]._element).addClass('select-highlight');
								}

								const valueMetadata = _.last(metadata);
								slices[valueMetadata.label] = valueMetadata.count;
								lastLabelIdx = i;
								break;
							}
						}
					}
					// Apply the current selection set to the facet.
					facet.select({ selection: { slices: slices } });
				} else {
					// See if this facet is in the values list, marking it as selected if it is.
					const matchedValue = values.find(v => v.toLowerCase() === facet.value.toLowerCase() ? facet.value.toLowerCase(): undefined);
					if (matchedValue) {

						facet.select(facet.count);

						// Check to see if this facet is the root selection, updating its visual state as necesary.
						if (this.isHighlightedGroup(highlights, group.key)) {
							$(facet._element).addClass('select-highlight');
						}

					} else {
						facet.deselect();
					}
				}
			}
		},

		injectHighlights(highlights: Highlights) {
			// Clear highlight state incase it was set via a click on on another
			// component
			$(this.$el).find('.select-highlight').removeClass('select-highlight');
			// if (this.isHighlightedInstance(highlights) || _.isEmpty(highlights.values)) {
			// 	$(this.$el).find('.select-highlight').removeClass('select-highlight');
			// }
			/// Update highlights
			this.groups.forEach(g => {
				const group = this.facets.getGroup(g.key);
				if (!group) {
					return;
				}
				this.injectHighlightsIntoGroup(group, highlights);
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
					// init the internal categorical facet filter state from the supplied facet
					// spec
					group.facets.forEach(facetSpec => {
						if (isCategoricalFacet(facetSpec) && !facetSpec.selected) {
							this.facetFilteredValues.set(group.key, facetSpec.value);
						}
					});
					this.injectHTML(group, this.facets.getGroup(group.key)._element);
					this.injectHighlightsIntoGroup(this.facets.getGroup(group.key), this.highlights);
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
					// init the internal categorical facet filter state from the supplied facet
					// spec
					groupSpec.facets.forEach(facetSpec => {
						if (isCategoricalFacet(facetSpec) && !facetSpec.selected) {
							this.facetFilteredValues.set(groupSpec.key, facetSpec.value);
						}
					});
					this.injectHTML(groupSpec, this.facets.getGroup(groupSpec.key)._element);
					this.injectHighlightsIntoGroup(this.facets.getGroup(groupSpec.key), this.highlights);
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
				if (this.facetFilteredValues.has(key, value)) {
					$icon = $(`<i id=${key}-${value} class="fa fa-circle-o"></i>`);
				} else {
					$icon = $(`<i id=${key}-${value} class="fa fa-circle"></i>`);
				}
				$icon.appendTo($facet);


				$icon.click(e => {
					// get group and current facet
					const group = this.facets.getGroup(key);
					const current = <any>(<CategoricalFacet[]>group.facets).find(facet => facet.value === value);

					// toggle the facet filter state
					if (!this.facetFilteredValues.has(key, value)) {
						// switch to unfilter from filtered
						$icon.removeClass('fa-circle').addClass('fa-circle-o');
						current.deselect();
						this.facetFilteredValues.set(key, value);
					} else {
						// switch from filtered to unfiltered, and restore highlight state if needed
						$icon.removeClass('fa-circle-o').addClass('fa-circle');
						if (_.isEqual(this.facetHighlightValue, { key, value })) {
							current.select({ count: current.count });
						}
						this.facetFilteredValues.delete(key, value);
					}
					// get all currently selected values
					const values = group.facets
						.filter(f => !this.facetFilteredValues.has(f.key, f.value))
						.map(f => f.value);

					this.$emit('facet-toggle', key, values);
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
.facets-facet-horizontal .facet-histogram-bar-highlighted.select-highlight {
	fill: #007bff;
}

.facets-facet-vertical.select-highlight .facet-bar-selected {
	box-shadow: inset 0 0 0 1000px #007bff;
}
</style>
