<template>
	<div class='card card-result'>
		<div class='dataset-header btn btn-success hover card-header' v-on:click='setActiveDataset()' v-bind:class='{collapsed: !expanded}'>
			<a class='nav-link'><b>Name:</b> {{name}}</a>
			<a class='nav-link'><b>Columns:</b> {{variables.length}}</a>
			<a class='nav-link'><b>Rows:</b> {{numRows}}</a>
			<a class='nav-link'><b>Size:</b> {{formatBytes(numBytes)}}</a>
		</div>
		<div class='card-body'>
			<div class='row'>
				<div class='col-4'>
					<span><b>Top features:</b></span>
					<ul id='example-1'>
						<li :key="variable.name" v-for='variable in topVariables'>
							{{variable.name}}
						</li>
					</ul>
				</div>
				<div class='col-8'>
					<span><b>Summary:</b></span>
					<p class='small-text'>
						{{summary}}
					</p>
				</div>
			</div>

			<div v-if='!expanded' class='card-expanded'>
				<b-button class='full-width hover' variant='secondary' v-on:click='toggleExpansion()'>
					More Details...
				</b-button>
			</div>

			<div v-if='expanded' class='card-expanded'>
				<span><h3>Full Description:</h3></span>
				<p v-html='highlightedDescription()'></p>
				<b-button class='full-width hover'variant='secondary' v-on:click='toggleExpansion()'>
					Less Details...
				</b-button>
			</div>

		</div>

	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import { createRouteEntry } from '../util/routes';
import { addRecentDataset } from '../util/data';
import { getters } from '../store/route/module';
import { Variable } from '../store/data/index';
import { SELECT_ROUTE } from '../store/route/index';

import Vue from 'vue';

const NUM_TOP_FEATURES = 5;
const SUFFIXES = {
	0: 'B',
	1: 'KB',
	2: 'MB',
	3: 'GB',
	4: 'TB',
	5: 'PB',
	6: 'EB'
};

export default Vue.extend({
	name: 'dataset-preview',

	props: {
		'name': String,
		'description': String,
		'summary': String,
		'variables': Array,
		'numRows': Number,
		'numBytes': Number
	},

	computed: {
		topVariables(): Variable[] {
			return (<Variable[]>this.variables).slice(0).sort((a, b) => {
				return b.importance - a.importance;
			}).slice(0, NUM_TOP_FEATURES);
		}
	},

	data() {
		return {
			expanded: false
		};
	},

	methods: {
		formatBytes(n: number): string {
			function formatRecursive(size: number, powerOfThousand: number): string {
				if (size > 1024) {
					return formatRecursive(size/1024, powerOfThousand+1);
				}
				return `${size.toFixed(2)}${SUFFIXES[powerOfThousand]}`;
			}
			return formatRecursive(n, 0);
		},
		setActiveDataset() {
			const entry = createRouteEntry(SELECT_ROUTE, {
				terms: getters.getRouteTerms(this.$store),
				dataset: this.name
			});
			this.$router.push(entry);
			addRecentDataset(this.name);
		},
		toggleExpansion() {
			this.expanded = !this.expanded;
		},
		highlightedDescription(): string {
			const terms = getters.getRouteTerms(this.$store);
			if (_.isEmpty(terms)) {
				return this.description;
			}
			const split = terms.split(/[ ,]+/); // split on whitespace
			const joined = split.join('|'); // join
			const regex = new RegExp(`(${joined})(?![^<]*>)`, 'gm');
			return this.description.replace(regex, '<span class="highlight">$1</span>');
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
	text-decoration: underline;
}
.card-result .card-header{
	background-color: #28a745;
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
.card-result .card-header:hover {
	color: #fff;
	background-color: #218838;
	border-color: #1e7e34;
}

</style>

