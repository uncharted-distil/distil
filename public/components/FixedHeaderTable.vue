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

		checkScrollBar() {
			// Check if there is fixed vertical scrollbar (when using mouse) in the table body
			// and give some space to the table header so that they align.
			const scrollbarWidth = this.tbody.offsetWidth - this.tbody.clientWidth;
			if (scrollbarWidth) {
				this.thead.style.width = `calc(100% - ${scrollbarWidth + 1}px)`;
				this.thead.style['margin-right'] = `${scrollbarWidth + 1}px`;
			}
		},

		resizeTableCells() {
			const theadCells = this.thead && this.thead.querySelectorAll('th');
			const firstRow = this.tbody && this.tbody.querySelector('tr');
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

			// get body and header cell width again from computed table header cells
			const bodyCells = [];
			const headCells = [];
			for (let i = 0; i < theadCells.length; i++) {
				const headCellWidth = theadCells[i].offsetWidth;
				headCells.push({ elem: theadCells[i], width: headCellWidth });
				bodyCells.push({ elem: tbodyCells[i], width: headCellWidth });
			}
			// set new cell width
			headCells.forEach(setCellWidth);
			bodyCells.forEach(setCellWidth);
		},

		onScroll() {
			const scrollLeft = this.tbody && this.tbody.scrollLeft;
			const tableHeaderRow = this.thead && this.thead.querySelector('tr');
			if (!_.isNil(scrollLeft) && tableHeaderRow) {
				tableHeaderRow.style['margin-left'] = scrollLeft ? `-${this.tbody.scrollLeft}px` : 0;
			}
		}
	},

	data: {
		tbody: {} as HTMLTableSectionElement,
		thead: {} as HTMLTableSectionElement,
		tableHeaderRow: {} as HTMLTableRowElement,
	},

	mounted: function () {
		this.thead = this.$el.querySelector('thead');
		this.tbody = this.$el.querySelector('tbody');

		this.tbody.addEventListener('scroll', this.onScroll);

		window.addEventListener('resize', this.resizeTableCells);
		Vue.nextTick(() => {
			this.checkScrollBar();
			this.resizeTableCells();
		});
	},

	beforeDestroy: function () {
		this.tbody.removeEventListener('scroll', this.onScroll);
		window.removeEventListener('resize', this.resizeTableCells);
	},
});

</script>


<style>
.fixed-header-table {
	height: inherit;
	width: 100%;
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
	overflow-x: hidden;

	/*
	  Subtract 1px from table header width.
	  This resolves the issue that table row overflows table width
	  by < 1px and creates horizontal scrollbar when it's not needed.
	*/
	width: calc(100% - 1px);
    margin-right: 1px;
}
.fixed-header-table thead tr {
	display: flex;
	width: 100%;
}
.fixed-header-table thead th {
	flex-shrink: 0;
	flex-grow: 1;
}
.fixed-header-table tbody {
	width: 100%;
	overflow: auto;
	flex: 1;
}
.fixed-header-table tbody td {
	overflow-wrap: break-word;
}
</style>
