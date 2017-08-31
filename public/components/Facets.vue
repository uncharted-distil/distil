<template>
	<div class="facets" v-once></div>
</template>

<script>
import _ from 'lodash';

import Facets from '@uncharted.software/stories-facets';
import '@uncharted.software/stories-facets/dist/facets.css';

export default {
	name: 'facets',

	props: [
		'groups'
	],

	mounted() {
		const component = this;
		// instantiate the external facets widget
		this.facets = new Facets(document.querySelector('.facets'), this.groups.map(group => {
			return _.cloneDeep(group);
		}));
		this.facets.getGroupIndices().forEach(key => {
			const group = component.facets.getGroup(key);
			// initialize selection
			group.initializeSelection();
		});
		// proxy events
		this.facets.on('facet-group:expand', (event, key) => {
			component.$emit('expand', key);
		});
		this.facets.on('facet-group:collapse', (event, key) => {
			component.$emit('collapse', key);
		});
		this.facets.on('facet-histogram:rangechangeduser', (event, key, value) => {
			component.$emit('range-change', key, value);
		});
		this.facets.on('facet:click', (event, key, value) => {
			// get group
			const group = component.facets.getGroup(key);
			// get facet
			const current = group.facets.filter(facet => {
				return facet.value === value;
			})[0];
			// toggle facet
			if (current._spec.selected) {
				current.deselect();
			} else {
				current.select({
					count: current.count
				});
			}
			// get all currently selected values
			const values = [];
			group.facets.forEach(facet => {
				if (facet._spec.selected) {
					// facet is selected
					values.push(facet.value);
				}
			});
			component.$emit('facet-toggle', key, values);
		});
	},

	watch: {
		groups: function(currGroups, prevGroups) {
			// get map of all existing group keys in facets
			const prevMap = {};
			prevGroups.forEach(group => {
				prevMap[group.key] = group;
			});
			// update and groups
			const unchangedGroups = this.updateGroups(currGroups, prevMap);
			// for the unchanged, update collapse state
			this.updateCollapsed(unchangedGroups);
			// for the unchanged, update selection
			this.updateSelections(unchangedGroups, prevMap);
		}
	},

	methods: {
		groupsEqual(a, b) {
			const OMITTED_FIELDS = ['selection', 'selected'];
			// NOTE: we dont need to check key, we assume its already equal
			if (a.label !== b.label) {
				return false;
			}
			if (a.facets.length !== b.facets.length) {
				return false;;
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
		updateGroups(currGroups, prevGroups) {
			const toAdd = [];
			const unchanged = [];
			// get map of all current, to track which groups need to be removed
			const toRemove = {};
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
			}
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
			unchangedGroups.forEach(group => {
				// get the existing facet
				const existing = this.facets.getGroup(group.key);
				if (existing) {
					const currFacets = group.facets;
					const prevFacets = prevGroups[group.key].facets;
					existing.facets.forEach((facet, index) => {
						const currSelection = currFacets[index].selection || currFacets[index].selected;
						const prevSelection = prevFacets[index].selection || prevFacets[index].selected;
						if (_.isEqual(currSelection, prevSelection)) {
							// selection is the same, no need to change
							return;
						}
						if (currSelection) {
							const facetSpec = group.facets[index];
							facet.select(facetSpec.selected ? facetSpec.selected : facetSpec);
						} else {
							facet.deselect();
						}
					});
				}
			});
		}
	},

	destroyed: function() {
		this.facets.destroy();
		this.facets = null;
	}
};
</script>

<style>
</style>
