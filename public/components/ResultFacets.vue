<template>
	<div class='result-facets'>
		<facets class="result-facets-container"
			:groups="groups"
			:highlights="highlights"
			:html="html"
			v-on:expand="onExpand"
			v-on:collapse="onCollapse"
			v-on:range-change="onRangeChange"></facets>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import Facets from '../components/Facets';
import { decodeFilters, updateFilter, getFilterType, isDisabled,
	CATEGORICAL_FILTER, NUMERICAL_FILTER, NumericalFilter, CategoricalFilter } from '../util/filters';
import { createRouteEntryFromRoute } from '../util/routes';
import { PipelineInfo, PipelineState } from '../store/pipelines/index';
import { Dictionary } from '../store/data/index';
import { getters as dataGetters} from '../store/data/module';
import { getters as routeGetters} from '../store/route/module';
import { createGroups, Group, NumericalFacet, CategoricalFacet } from '../util/facets';
import { getPipelineResults, getPipelineResultsOkay } from '../util/pipelines';
import 'font-awesome/css/font-awesome.css';
import '../styles/spinner.css';
import Vue from 'vue';

export default Vue.extend({
	name: 'result-facets',

	components: {
		Facets
	},

	props: {
		'variables': Array,
		'dataset': String,
		'html': String
	},

	computed: {
		groups(): Group[] {
			// create the groups
			let groups = createGroups(this.variables, true, false, '');

			// sort alphabetically
			groups.sort((a, b) => {
				const textA = a.key.toLowerCase();
				const textB = b.key.toLowerCase();
				return (textA < textB) ? -1 : (textA > textB) ? 1 : 0;
			});

			// find pipeline result with the uri specified in the route and
			// flag it as the currently active result
			const requestId = routeGetters.getRouteCreateRequestId(this.$store) as string;
			const pipelineResults = getPipelineResultsOkay(<PipelineState>this.$store.state.pipelineModule, requestId) as PipelineInfo[];
			const activeResult = _.find(pipelineResults, p => {
				return btoa(p.pipeline.resultId) === routeGetters.getRouteResultId(this.$store);
			});

			const filters = routeGetters.getRouteResultFilters(this.$store);

			// if filters are empty this is the first group call - initialize
			// filter and group state
			if (_.isEmpty(filters)) {
				// set the selected value to the route value
				groups.forEach((group) => {
					if (group.key !== activeResult.name) {
						this.updateFilterRoute(group.key, { enabled: false }, null);
					} else {
						this.updateFilterRoute(group.key, { enabled: true}, activeResult.pipeline.resultId);
						this.$emit('activePipelineChange', {
							name: activeResult.name,
							id: activeResult.pipelineId
						});
					}
				});
			}
			// update collapsed state
			groups = this.updateGroupCollapses(groups);
			// update selections
			return this.updateGroupSelections(groups);
		},

		highlights(): Dictionary<any> {
			return dataGetters.getHighlightedFeatureValues(this.$store);
		}
	},

	methods: {
		updateFilterRoute(key: string, values: Dictionary<any>, resultUri: string) {

			// merge the updated filters back into the route query params if set
			const filters = routeGetters.getRouteResultFilters(this.$store);
			let updatedFilters = filters;
			if (key && values) {
				updatedFilters = updateFilter(filters, key, values);
			}

			const entry = createRouteEntryFromRoute(routeGetters.getRoute(this.$store), {
				resultId: resultUri ? btoa(resultUri) : routeGetters.getRouteResultId(this.$store),
				results: updatedFilters
			});

			this.$router.push(entry);
		},

		onExpand(key: string) {

			const createReqId = routeGetters.getRouteCreateRequestId(this.$store);
			const pipelineRequests = getPipelineResults(<PipelineState>this.$store.state.pipelineModule, createReqId);
			const completedReq = _.find(pipelineRequests, p => p.name === key);

			// disable all filters except this one
			this.groups.forEach(group => {
				if (group.key !== key) {
					this.updateFilterRoute(group.key, { enabled: false  }, null);
				}
			});

			// enable filter
			this.updateFilterRoute(key, { enabled: true }, completedReq.pipeline.resultId);
			// let listening components know the acitive pipeline changed
			this.$emit('activePipelineChange', { name: completedReq.name, id: completedReq.pipelineId });
		},

		onCollapse() {
			// TODO: prevent disabling?
			// no-op
		},

		onRangeChange(key: string, value: { from: { label: string[] }, to: { label: string[] } }) {
			// set range filter
			this.updateFilterRoute(key, {
					enabled: true,
					min: parseFloat(value.from.label[0]),
					max: parseFloat(value.to.label[0])
				}, null);
		},

		updateGroupCollapses(groups: Group[]): Group[] {
			const filters = routeGetters.getRouteResultFilters(this.$store);
			const decoded = decodeFilters(filters);
			return groups.map(group => {
				// return if disabled
				group.collapsed = isDisabled(decoded[group.key]);
				return group;
			});
		},
		updateGroupSelections(groups: Group[]): Group[] {
			const filters = routeGetters.getRouteResultFilters(this.$store);
			const decoded = decodeFilters(filters);
			return groups.map(group => {
				// get filter
				const filter = decoded[group.key];
				switch (getFilterType(filter)) {
					case NUMERICAL_FILTER:
						// add selection to facets
						group.facets.forEach(facet => {
							if ((<NumericalFacet>facet).selection) {
								(<NumericalFacet>facet).selection = {
									// NOTE: the `from` / `to` values MUST be strings.
									range: {
										from: `${(<NumericalFilter>filter).min}`,
										to: `${(<NumericalFilter>filter).max}`,
									}
								};
							}
						});
						break;

					case CATEGORICAL_FILTER:
						// add selection to facets
						group.facets.forEach(facet => {
							if ((<CategoricalFacet>facet).value) {
								const categoricalFacet = <CategoricalFacet>facet;
								if ((<CategoricalFilter>filter).categories.indexOf(categoricalFacet.value) !== -1) {
									// select
									categoricalFacet.selected = {
										count: categoricalFacet.count
									};
								} else {
									delete categoricalFacet.selected;
								}
							}
						});
						break;
				}
				return group;
			});
		}
	}
});
</script>

<style>
button {
	cursor: pointer;
}

.variable-facets {
	display: flex;
	flex-direction: column;
	padding: 8px;
}

.result-facets-container {
	overflow-x: hidden;
	overflow-y: hidden;
}
</style>
