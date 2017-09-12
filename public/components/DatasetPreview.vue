<template>
	<div class="card">

		<div class="dataset-header card-header" v-bind:class="{collapsed: !expanded}">
			<a class="nav-link hover" v-on:click="setActiveDataset()">
				{{name}}
			</a>
			<a class="nav-link hover" v-on:click="toggleExpansion()">
				<i v-if="expanded" class="fa fa-minus"></i>
				<i v-if="!expanded" class="fa fa-plus"></i>
			</a>
		</div>
		<div class="card-body" v-if="expanded">
			<p class="p-2" v-html="highlightedDescription()"></p>
		</div>

	</div>
</template>

<script>

import _ from 'lodash';
import {createRouteEntry} from '../util/routes';

export default {
	name: 'dataset-preview',

	props: [
		'name',
		'description'
	],

	data() {
		return {
			expanded: false
		};
	},

	methods: {
		setActiveDataset() {
			const entry = createRouteEntry('/explore', {
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
	justify-content: space-between
}

.dataset-header.collapsed {
	border-bottom: none;
}
</style>
