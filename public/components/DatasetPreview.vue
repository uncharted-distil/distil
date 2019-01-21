<template>
	<div class='card card-result'>
		<div class='dataset-header hover card-header'  variant="dark" @click.stop='setActiveDataset()' v-bind:class='{collapsed: !expanded}'>
			<a class='nav-link'><b>Name:</b> {{name}}</a>
			<a class='nav-link'><b>Features:</b> {{variables.length}}</a>
			<a class='nav-link'><b>Rows:</b> {{numRows}}</a>
			<a class='nav-link'><b>Size:</b> {{formatBytes(numBytes)}}</a>
			<a v-if="allowImport && provenance==='datamart'"><b-button variant="danger" @click.stop='importDataset()'>Import</b-button></a>
			<a v-if="allowJoin"><b-button variant="primary" @click.stop='joinDataset()'>Join</b-button></a>
		</div>
		<div class='card-body'>
			<div class='row'>
				<div class='col-4'>
					<span><b>Top features:</b></span>
					<ul>
						<li :key="variable.name" v-for='variable in topVariables'>
							{{variable.colDisplayName}}
						</li>
					</ul>
				</div>
				<div class='col-8'>
					<div v-if="summaryML.length > 0">
						<span><b>Topics:</b></span>
						<p class='small-text'>
							{{summaryML}}
						</p>
					</div>
					<span><b>Summary:</b></span>
					<p class='small-text'>
						{{summary}}
					</p>
				</div>
			</div>

			<div v-if='!expanded' class='card-expanded'>
				<b-button class='full-width hover' variant='outline-secondary' v-on:click='toggleExpansion()'>
					More Details...
				</b-button>
			</div>

			<div v-if='expanded' class='card-expanded'>
				<span><b>Full Description:</b></span>
				<p v-html='highlightedDescription()'></p>
				<b-button class='full-width hover' variant='outline-secondary' v-on:click='toggleExpansion()'>
					Less Details...
				</b-button>
			</div>

		</div>
	</div>

</template>

<script lang="ts">

import _ from 'lodash';
import Vue from 'vue';
import { createRouteEntry } from '../util/routes';
import { formatBytes } from '../util/bytes';
import { sortVariablesByImportance } from '../util/data';
import { getters as routeGetters } from '../store/route/module';
import { Variable } from '../store/dataset/index';
import { actions as datasetActions } from '../store/dataset/module';
import { SELECT_TARGET_ROUTE } from '../store/route/index';
import localStorage from 'store';

const NUM_TOP_FEATURES = 5;

export default Vue.extend({
	name: 'dataset-preview',

	props: {
		id: String as () => string,
		name: String as () => string,
		description: String as () => string,
		summary: String as () => string,
		summaryML: String as () => string,
		variables: Array as () => Variable[],
		numRows: Number as () => number,
		numBytes: Number as () => number,
		provenance: String as () => string,
		allowImport: Boolean as () => boolean,
		allowJoin: Boolean as () => boolean
	},

	computed: {
		topVariables(): Variable[] {
			return sortVariablesByImportance(this.variables.slice(0)).slice(0, NUM_TOP_FEATURES);
		}
	},

	data() {
		return {
			expanded: false
		};
	},

	methods: {
		formatBytes(n: number): string {
			return formatBytes(n);
		},
		setActiveDataset() {
			const entry = createRouteEntry(SELECT_TARGET_ROUTE, {
				terms: routeGetters.getRouteTerms(this.$store),
				dataset: this.name
			});
			this.$router.push(entry);
			this.addRecentDataset(this.name);
		},
		toggleExpansion() {
			this.expanded = !this.expanded;
		},
		highlightedDescription(): string {
			const terms = routeGetters.getRouteTerms(this.$store);
			if (_.isEmpty(terms)) {
				return this.description;
			}
			const split = terms.split(/[ ,]+/); // split on whitespace
			const joined = split.join('|'); // join
			const regex = new RegExp(`(${joined})(?![^<]*>)`, 'gm');
			return this.description.replace(regex, '<span class="highlight">$1</span>');
		},
		addRecentDataset(dataset: string) {
			const datasets = localStorage.get('recent-datasets') || [];
			if (datasets.indexOf(dataset) === -1) {
				datasets.unshift(dataset);
				localStorage.set('recent-datasets', datasets);
			}
		},
		importDataset() {
			datasetActions.importDataset(this.$store, {
				datasetID: this.id
			});
		},
		joinDataset() {
			this.$emit('join-dataset', this.name);
		}

	}
});
</script>

<style>
.highlight {
	background-color: #87CEFA;
}
.dataset-header {
	display: flex;
	padding: 4px 8px;
	color: white;
	justify-content: space-between;
	border: none;
	border-bottom: 1px solid rgba(0, 0, 0, 0.125);
}
.card-result .card-header {
	background-color: #424242;
}
.card-result .card-header:hover {
	color: #fff;
	background-color: #535353;
}
.dataset-header:hover {
	text-decoration: underline;
}
.full-width {
	width: 100%;
}
.card-expanded {
	padding-top: 15px;
}
</style>
