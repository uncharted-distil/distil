<template>
	<div class='row h-100 variable-facets'>
		<div class="col-12 flex-column d-flex">
			<div v-if="enableSearch" class="row flex-1 align-items-center facet-filters">
				<div class="col-12 flex-column d-flex">
					<b-form-input size="sm" v-model="filter" placeholder="Search" />
				</div>
			</div>
			<div v-if="enableTitle" class="row flex-1 align-items-center">
				<div class="col-12 flex-column d-flex">
					<p>Select one of the following feature summaries (sorted by interestingness) showing count of records by feature value.</p>
				</div>
			</div>
			<div class="pl-2">
				<slot></slot>
			</div>
			<div class="row flex-11">
				<facets class="col-12 flex-column d-flex variable-facets-container"
					:groups="sortedGroups"
					:highlights="highlights"
					:row-selection="rowSelection"
					:html="html"
					:sort="sort"
					:enable-type-change="enableTypeChange"
					:enable-highlighting="enableHighlighting"
					:instanceName="instanceName"
					@numerical-click="onNumericalClick"
					@categorical-click="onCategoricalClick"
					@range-change="onRangeChange"
					@facet-click="onFacetClick">
				</facets>
			</div>
			<div v-if="numRows > rowsPerPage" class="row flex-1 align-items-center variable-page-nav">
				<div class="col-12 flex-column">
					<b-pagination size="sm" align="center" :total-rows="numRows" :per-page="rowsPerPage" v-model="currentPage" class="mb-0"/>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts">

import Facets from '../components/Facets';
import { overlayRouteEntry, getRouteFacetPage } from '../util/routes';
import { Dictionary } from '../util/dict';
import { Highlight, RowSelection } from '../store/data/index';
import { getters as dataGetters } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { Group } from '../util/facets';
import { updateHighlightRoot, getHighlights, clearHighlightRoot } from '../util/highlights';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';
import Vue from 'vue';

export default Vue.extend({
	name: 'variable-facets',

	components: {
		Facets
	},

	props: {
		enableSearch: Boolean,
		enableTitle: Boolean,
		enableTypeChange: Boolean,
		enableHighlighting: Boolean,
		groups: Array,
		dataset: String,
		subtitle: String,
		html: [ String, Object, Function ],
		instanceName: { type: String, default: 'variable-facets' },
	},

	data() {
		return {
			filter: '',
			numRows: 1,
			rowsPerPage: 10,
			sortMethod: 'importanceDesc'
		};
	},

	computed: {
		currentPage: {
			set(page: number) {
				const entry = overlayRouteEntry(this.$route, {
					[this.routePageKey()]: page
				});
				this.$router.push(entry);
			},
			get(): number {
				return getRouteFacetPage(this.routePageKey(), this.$route);
			}
		},

		sortedGroups(): Group[] {
			// filter by search
			const searchFiltered = this.groups.filter(group => {
				return this.filter === '' || group.key.toLowerCase().includes(this.filter.toLowerCase());
			});

			// sort by current function - sort looks for key to hold sort key
			const sorted = searchFiltered.map(g => ({ key: g.key, group: g }))
				.sort((a, b) => this[this.sortMethod](a, b))
				.map(g => g.group);

			// if necessary, refilter applying pagination rules
			this.numRows = searchFiltered.length;
			let filtered = sorted;
			if (this.numRows > this.rowsPerPage) {
				const firstIndex = this.rowsPerPage * (this.currentPage - 1);
				const lastIndex = Math.min(firstIndex + this.rowsPerPage, this.numRows);
				filtered = sorted.slice(firstIndex, lastIndex);
			}

			// highlight
			if (this.enableHighlighting && this.highlights.root) {
				filtered.forEach(group => {
					if (group) {
						if (group.key === this.highlights.root.key) {
							group.facets.forEach(facet => {
								facet.filterable = true;
							});
						}
					}
				});
			}

			return filtered;
		},

		highlights(): Highlight {
			return getHighlights(this.$store);
		},

		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},

		importance(): Dictionary<number> {
			const variables = dataGetters.getVariables(this.$store);
			const importance: Dictionary<number> = {};
			variables.forEach(variable => {
				importance[variable.name] = variable.importance;
			});
			return importance;
		},

		sort() {
			return (<any>this)[(<any>this).sortMethod];
		}
	},

	methods: {
		importanceDesc(a: { key: string }, b: { key: string }): number {
			const importance = this.importance;
			return importance[b.key] - importance[a.key];
		},

		availableVariables(): string[] {
			// filter by search
			const searchFiltered = this.groups.filter(group => {
				return this.filter === '' || group.key.toLowerCase().includes(this.filter.toLowerCase());
			});
			return searchFiltered.map(v => v.key);
		},

		// creates a facet key for the route from the instance-name component arg
		// or uses a default if unset
		routePageKey(): string {
			if (this.instanceName) {
				return `${this.instanceName}Page`;
			}
			return 'facetPage';
		},

		onRangeChange(context: string, key: string, value: { from: number, to: number }) {
			updateHighlightRoot(this, {
				context: context,
				key: key,
				value: value
			});
			this.$emit('range-change', key, value);
		},

		onFacetClick(context: string, key: string, value: string) {
			if (this.enableHighlighting) {
				if (key && value) {
					// extract the var name from the key
					updateHighlightRoot(this, {
						context: context,
						key: key,
						value: value
					});
				} else {
					clearHighlightRoot(this);
				}
			}
			this.$emit('facet-click', context, key, value);
		},

		onCategoricalClick(context: string, key: string) {
			this.$emit('categorical-click', key);
		},

		onNumericalClick(context: string, key: string, value: { from: number, to: number }) {
			if (this.enableHighlighting) {
				if (!this.highlights.root || this.highlights.root.key !== key) {
					updateHighlightRoot(this, {
						context: this.instanceName,
						key: key,
						value: value
					});
				}
			}
			this.$emit('numerical-click', key);
		}
	}
});

</script>

<style>
button {
	cursor: pointer;
}
.page-link {
	color: #868e96;
}
.page-item.active .page-link {
	z-index: 2;
	color: #fff;
	background-color: #868e96;
	border-color: #868e96;
}
.variable-facets-container {
	overflow-x: hidden;
	overflow-y: auto;
}
.variable-facets-container .facets-root-container{
	margin: 2px;
}
.variable-facets-container .facets-root-container .facets-group-container{
	background-color: inherit;
}
.variable-facets-container .facets-root-container .facets-group-container .facets-group {
	background: white;
	margin: 2px 2px 4px 2px;
	font-size: 0.867rem;
	color: rgba(0,0,0,0.87);
	box-shadow: 0 1px 2px 0 rgba(0,0,0,0.10);
	transition: box-shadow 0.3s ease-in-out;
}
.variable-facets-container .facets-root-container .facets-group-container .facets-group .group-header {
	padding: 4px 8px 6px 8px;
}
.variable-facets-container .facets-root-container .facets-group-container .facets-group .group-header .enable-type-change-menu {
	float: right;
	margin-top: -4px;
	margin-right: -8px;
}
.facet-filters {
	margin: 0 -10px 4px -10px;
}
.facet-filters span {
	font-size: 0.9rem;
}
.variable-page-nav {
	padding-top: 10px;
}

</style>
