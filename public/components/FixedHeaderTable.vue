<template>
	<div class="fixed-header-table">
		<!-- slot for html table -->
        <slot></slot>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import _ from 'lodash';

export default Vue.extend({
	name: 'fixed-header-table',
	methods: {
		resizeTableCells() {
			const thead = this.$el.querySelector('thead');
			const theadCells = thead && thead.querySelectorAll('th');
			const firstRow = this.$el.querySelector('tbody tr');
			const tbodyCells = firstRow && firstRow.querySelectorAll('td');

			if (_.isEmpty(theadCells) || _.isEmpty(tbodyCells)) {
				return;
			}

			const headTargetCells = [];

			// reset element style so that table renders with initial layout set by css
			for (let i = 0; i < theadCells.length; i++) {
				tbodyCells[i].removeAttribute('style');
				theadCells[i].removeAttribute('style');
			}
			// get new adjusted header cell width based on the corresponding data cell width
			for (let i = 0; i < theadCells.length; i++) {
				const headCellWidth = theadCells[i].offsetWidth;
				const bodyCellWidth = tbodyCells[i].offsetWidth;
				if (headCellWidth < bodyCellWidth) {
					headTargetCells.push({ elem: theadCells[i], width: bodyCellWidth });
				}
			}
			const setCellWidth = cell => {
				cell.elem.style['max-width'] = cell.width + 'px';
				cell.elem.style['min-width'] = cell.width + 'px';
			};
			headTargetCells.forEach(setCellWidth);

			// get body cell width from computed table header cell width
			const bodyCells = [];
			for (let i = 0; i < theadCells.length; i++) {
				const headCellWidth = theadCells[i].offsetWidth;
				bodyCells.push({ elem: tbodyCells[i], width: headCellWidth });
			}
			// set new body cell width
			bodyCells.forEach(setCellWidth);
		}
	},
	mounted: function () {
		window.addEventListener('resize', this.resizeTableCells);
		this.resizeTableCells();
	},
	beforeDestroy: function () {
		window.removeEventListener('resize', this.resizeTableCells);
	},
});

</script>


<style>
.fixed-header-table {
	overflow-x: auto;
	height: inherit;
	width: inherit;
	position: relative;
	border: 1px solid rgb(245, 245, 245);
}
.fixed-header-table table {
	table-layout: fixed;
	height: 100%;
	margin: 0;
	border: none;
	display: flex;
	flex-direction: column;
	align-items: flex-start;
}
.fixed-header-table thead {
	width: 100%
}
.fixed-header-table thead tr {
	display: flex;
}
.fixed-header-table thead th {
	flex-shrink: 0;
	flex-grow: 1;
}
.fixed-header-table tbody {
	overflow-y: auto;
	flex: 1;
}
.fixed-header-table tbody td {
	overflow-wrap: break-word;
}
</style>