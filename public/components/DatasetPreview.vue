<template>
	<div class='card'>
		<div class='dataset-header card-header btn btn-outline-success hover' v-on:click='setActiveDataset()' v-bind:class='{collapsed: !expanded}'>
			<a class='nav-link'><b>Name:</b> {{name}}</a>
			<a class='nav-link'><b>Columns:</b> {{variables.length}}</a>
			<a class='nav-link'><b>Rows:</b> {{numRows}}</a>
			<a class='nav-link'><b>Size:</b> {{formatBytes(numBytes)}}</a>
		</div>
		<div class='card-body'>
			<div class='row'>
				<div class='col-4'>
					<span><b>Top Features:</b></span>
					<ul id='example-1'>
						<li class="small-text" v-for='variable in topVariables'>
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

			<div v-if='!expanded'>
				<b-button class='full-width hover' variant='outline-secondary' v-on:click='toggleExpansion()'>
					More Details...
				</b-button>
			</div>

			<div v-if='expanded'>
				<span><b>Full Description:</b></span>
				<p class='p-2' v-html='highlightedDescription()'></p>
				<b-button class='full-width hover'variant='outline-secondary' v-on:click='toggleExpansion()'>
					Less Details...
				</b-button>
			</div>

		</div>

	</div>
</template>

<script>

import _ from 'lodash';
import {createRouteEntry} from '../util/routes';

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

export default {
	name: 'dataset-preview',

	props: [
		'name',
		'description',
		'summary',
		'variables',
		'numRows',
		'numBytes'
	],

	computed: {
		topVariables() {
			return this.variables.slice(0).sort((a, b) => {
				return a.importance - b.importance;
			}).slice(0, NUM_TOP_FEATURES);
		}
	},

	data() {
		return {
			expanded: false
		};
	},

	methods: {
		formatBytes(n) {
			function formatRecursive(size, powerOfThousand) {
				if (size > 1024) {
					return formatRecursive(size/1024, powerOfThousand+1);
				}
				return `${size.toFixed(2)}${SUFFIXES[powerOfThousand]}`;
			}
			return formatRecursive(n, 0);
		},
		datasetSummary() {
			return 'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.';
		},
		setActiveDataset() {
			const entry = createRouteEntry('/explore', {
				terms: this.$store.getters.getRouteTerms(),
				dataset: this.name
			});
			this.$router.push(entry);
		},
		toggleExpansion() {
			this.expanded = !this.expanded;
		},
		highlightedDescription() {
			const terms = this.$store.getters.getRouteTerms();
			if (_.isEmpty(terms)) {
				return this.description;
			}
			const split = terms.split(/[ ,]+/); // split on whitespace
			const joined = split.join('|'); // join
			const regex = new RegExp(`(${joined})(?![^<]*>)`, 'gm');
			return this.description.replace(regex, '<span class="highlight">$1</span>');
		}
	}
};
</script>

<style>
.highlight {
	background-color: #87CEFA;
}
.dataset-header {
	display: flex;
	padding: 4px 8px;
	color: #28a745;
	justify-content: space-between;
	border: none;
	border-bottom: 1px solid rgba(0, 0, 0, 0.125);
	text-decoration: underline;
}
.dataset-header:hover {
	text-decoration: underline;
}
.full-width {
	width: 100%;
}
.small-text {
	font-size: 14px;
}
</style>
