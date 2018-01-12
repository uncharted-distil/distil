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

export default Vue.extend({
	name: 'facets',

	props: {
		groups: Array,
		highlights: Object,
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
			facets: {} as any
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
			component.$emit('histogram-click', key, value);
		});
		this.facets.on('facet:click', (event: Event, key: string, value: string) => {
			component.$emit('facet-click', key, value);
		});
	},

	watch: {
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
			// for the unchanged, update selection
			this.updateSelections(unchangedGroups, prevMap);

		},

		highlights(currHighlights: Dictionary<any>) {
			if (_.isEmpty(currHighlights)) {
				(this.groups as Group[]).forEach(groupSpec => {
					const group = this.facets.getGroup(groupSpec.key);
					const facetSpecs = groupSpec.facets;
					group.facets.forEach((facet, index) => {
						const facetSpec = <any>facetSpecs[index];
						const selection = facetSpec.selection || facetSpec.selected;
						if (selection) {
							facet.select(facetSpec.selected ? facetSpec.selected : facetSpec);
						} else {
							facet.deselect();
						}
					});
				});
				return;
			}
			_.forIn(currHighlights, (value, name) => {
				const group = this.facets.getGroup(name);
				if (group) {
					group.facets.forEach(facet => {
						if (facet._histogram && facet._histogram.highlightRange) {
							// histogram facet
							facet._histogram.highlightValueRange({
								from: value,
								to: value
							});
						} else {
							// vertical facet
							if (facet.value === value) {
								facet.select(facet.count);
							} else {
								facet.deselect();
							}
						}
					});
				}
			});
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
				toAdd.forEach(group => {
					this.injectHTML(group, this.facets.getGroup(group.key)._element);
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

		updateSelections(unchangedGroups, prevGroups) {
			unchangedGroups.forEach(groupSpec => {
				// get the existing facets from the Facet lib
				const existing = this.facets.getGroup(groupSpec.key);
				if (existing) {
					// get the local facet specs
					const currFacets = groupSpec.facets;
					const prevFacets = prevGroups[groupSpec.key].facets;
					existing.facets.forEach((facet, index) => {
						// update the values in the Facet lib from the local specs if there's a delta
						const currSelection = currFacets[index].selection || currFacets[index].selected;
						const prevSelection = prevFacets[index].selection || prevFacets[index].selected;
						if (_.isEqual(currSelection, prevSelection)) {
							// selection is the same, no need to change
							return;
						}
						if (currSelection) {
							const facetSpec = currFacets[index];
							facet.select(facetSpec.selected ? facetSpec.selected : facetSpec);
						} else {
							facet.deselect();
						}
					});
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
		injectCategoryToggleButtons(fgroup: Group, $elem: JQuery) {
			if (!isCategoricalFacet(fgroup.facets[0])) {
				return;
			}

			// find the facet nodes in the DOM
			const $verticalFacets = $elem.find('.facets-facet-vertical');

			// Add a clickable state button to each facet.
			for (const facetElement of $verticalFacets) {

				const $facet = $(facetElement).find('.facet-query-close');
				const label = $facet.parent().find('.facet-label').text().trim();
				const ffacet = (<CategoricalFacet[]>fgroup.facets).find(f => f.value === label);

				// only add controls for filterable facets
				if (!ffacet.filterable) {
					continue;
				}

				// setup based on initial state
				let $icon = null;
				if (!ffacet.selected) {
					$icon = $('<i class="fa fa-circle-o"></i>');
				} else {
					$icon = $('<i class="fa fa-circle"></i>');
				}
				$icon.appendTo($facet);

				const key = fgroup.key;
				const value = ffacet.value;

				$icon.click(e => {
					// get group and current facet
					const group = this.facets.getGroup(key);
					const current = <any>(<CategoricalFacet[]>group.facets).find(facet => facet.value === value);

					// toggle facet
					if (current._spec.selected) {
						$icon.removeClass('fa-circle').addClass('fa-circle-o');
						current.deselect();
					} else {
						$icon.removeClass('fa-circle-o').addClass('fa-circle');
						current.select({ count: current.count });
					}
					// get all currently selected values
					const values = group.facets.filter(f => f._spec.selected).map(f => f.value);

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
.facets-facet-vertical .facet-label-count,
.facets-facet-vertical .facet-label,
.facets-group .group-header {
	font-family: inherit;
}
</style>
