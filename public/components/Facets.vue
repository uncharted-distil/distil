<template>
	<div class="facets" v-once ref="facets"></div>
</template>

<script lang="ts">
import _ from 'lodash';
import $ from 'jquery';
import Vue from 'vue';
import { Group, CategoricalFacet, isCategoricalFacet, CATEGORICAL_CHUNK_SIZE } from '../util/facets';
import { Highlight } from '../store/data/index';
import { Dictionary } from '../util/dict';
import Facets from '@uncharted.software/stories-facets';
import TypeChangeMenu from '../components/TypeChangeMenu';
import '@uncharted.software/stories-facets/dist/facets.css';

export default Vue.extend({
	name: 'facets',

	props: {
		groups: Array,
		highlights: Object,
		enableTypeChance: Boolean,
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
			instanceName: _.uniqueId('facet-'),
			more: {}
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
			const range = {
				from: _.toNumber(value.from.label[0]),
				to: _.toNumber(value.to.label[0])
			};
			component.$emit('range-change', this.instanceName, key, range);
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
			component.$emit('facet-more', key);
			if (!component.more[key]) {
				Vue.set(component.more, key, 0);
			}
			Vue.set(component.more, key, component.more[key] + CATEGORICAL_CHUNK_SIZE);
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
		processedGroups(): Group[] {
			const groups = _.cloneDeep(this.groups);
			groups.forEach(group => {
				const more = this.more[group.key];
				if (more) {
					group.facets = group.facets.concat(group.remaining.slice(0, more));
					group.remaining = group.remaining.slice(more);
					let remainingTotal = 0;
					group.remaining.forEach(facet => {
						remainingTotal += facet.count;
					});
					group.more = group.remaining.length;
					group.moreTotal = remainingTotal;
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
			this.injectHighlights(currHighlights);
		},

		sort(currSort) {
			this.facets.sort(currSort);
		}
	},

	methods: {

		isCategorical(group: any): boolean {
			if (group.facets.length === 0) {
				return false;
			}
			if (group.facets[0].histogram) {
				return false;
			}
			return true;
		},

		isNumerical(group: any): boolean {
			if (group.facets.length === 0) {
				return false;
			}
			if (group.facets[0].histogram) {
				return true;
			}
			return false;
		},

		injectHTML(group: Group, $elem: JQuery) {
			$elem.click(event => {
				if (this.isNumerical(group)) {
					this.$emit('numerical-click', group.key);
				} else if (this.isCategorical(group)) {
					this.$emit('categorical-click', group.key);
				}
			});

			$elem.find('.facet-histogram g').click(event => {
				this.$emit('numerical-click', group.key);
			});

			// inject type icon in group header
			this.injectTypeIcon(group, $elem);

			// inject type change header menus
			this.injectTypeChangeHeaders(group, $elem);

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

		isHighlightedInstance(highlights: Highlight): boolean {
			return _.get(highlights, 'root.context') === this.instanceName;
		},

		isHighlightedGroup(highlights: Highlight, key: string): boolean {
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

		getHighlightRootValue(highlights: Highlight): any {
			if (highlights.root) {
				if (highlights.root.value) {
					if (_.isArray(highlights.root.value)) {
						return null;
					}
					if (_.isString(highlights.root.value)) {
						return highlights.root.value;
					}
					return highlights.root.value;
				}
			}
			return null;
		},

		getHighlightSummaries(highlights: Highlight): any {
			if (highlights.values) {
				return highlights.values.summaries;
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

		injectHighlightsIntoGroup(group: any, highlights: Highlight) {

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

			const highlightRootValue = this.getHighlightRootValue(highlights);

			if (!highlightRootValue) {
				// no value to highlight, exit early
				return;
			}

			const summaries = this.getHighlightSummaries(highlights);

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

						const summary = _.find(summaries, s => {
							return s.name === group.key;
						});

						if (summary) {

							const bars = facet._histogram.bars;

							const slices = {};

							summary.buckets.forEach((bucket, index) => {
								const entry: any = _.last(bars[index].metadata);
								slices[entry.label] = bucket.count;
							});

							// ensure min height
							//this.ensureMinHeight(slices, bars);

							selection.slices = slices;
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
						} else {
							this.deselectCategoricalFacet(facet);
						}

					} else {

						const summary = _.find(summaries, s => {
							return s.name === group.key;
						});

						if (summary) {

							const bucket = _.find(summary.buckets, b => {
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

		injectHighlights(highlights: Highlight) {
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
			if (isCategoricalFacet(group.facets.length > 0 && group.facets[0])) {
				const facetSpecs = (<CategoricalFacet[]>group.facets);
				const typeicon = facetSpecs[0].icon.class;
				const $icon = $(`<i class="${typeicon}"></i>`);
				$elem.find('.group-header').append($icon);
			}
		},

		getGroupSampleValues(group: Group): any[] {
			let values = [];
			group.facets.forEach((facet: any) => {
				if (facet.histogram) {
					values = facet.histogram.slices.slice(0, 10).map(b => _.toNumber(b.label));
				} else {
					values.push(facet.value);
				}
			});
			return values.filter(v => v !== undefined);
		},

		// inject type headers
		injectTypeChangeHeaders(group: Group, $elem: JQuery) {
			if (this.enableTypeChance) {
				const $slot = $('<span/>');
				$elem.find('.group-header').append($slot);
				const menu = new TypeChangeMenu(
					{
						store: this.$store,
						propsData: {
							field: group.key,
							values: this.getGroupSampleValues(group)
						}
					});
				menu.$mount($slot[0]);
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

.facet-histogram {
	cursor: pointer !important;
}

.facets-facet-vertical.select-highlight .facet-bar-selected {
	box-shadow: inset 0 0 0 1000px #007bff;
}
</style>
