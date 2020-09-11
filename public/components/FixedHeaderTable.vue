<template>
  <div class="fixed-header-table">
    <!-- slot for html table -->
    <slot></slot>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";

export default Vue.extend({
  name: "fixed-header-table",

  methods: {
    checkScrollBar() {
      // Check if there is fixed vertical scrollbar (when using mouse) in the table body
      // and give some space to the table header so that they align.
      const scrollbarWidth = this.tbody.offsetWidth - this.tbody.clientWidth;
      if (scrollbarWidth) {
        this.thead.style.width = `calc(100% - ${scrollbarWidth + 1}px)`;
        this.thead.style["margin-right"] = `${scrollbarWidth + 1}px`;
      }
    },

    resizeTableCells() {
      this.thead = this.$el.querySelector("thead");
      const header = this.thead && this.thead.querySelector("tr");
      const theadCells = header && header.querySelectorAll("th");
      const firstRow = this.tbody && this.tbody.querySelector("tr");
      const tbodyCells = firstRow && firstRow.querySelectorAll("td");

      if (_.isEmpty(theadCells) || _.isEmpty(tbodyCells)) {
        return;
      }

      const headTargetCells = [];
      const bodyTargetCells = [];
      const minCellWidth = 200;
      const evenCellWidth = Math.ceil(
        this.tbody.clientWidth / theadCells.length
      );
      const maxCellWidth =
        minCellWidth > evenCellWidth ? minCellWidth : evenCellWidth;
      const setCellWidth = (cell) => {
        cell.elem.style["max-width"] = cell.width + "px";
        cell.elem.style["min-width"] = cell.width + "px";
      };
      // reset element style so that table renders with initial layout set by css
      for (let i = 0; i < theadCells.length; i++) {
        tbodyCells[i].removeAttribute("style");
        theadCells[i].removeAttribute("style");
      }
      // set all cells to even number
      for (let i = 0; i < theadCells.length; i++) {
        const theadWidth = Math.ceil(
          theadCells[i].getBoundingClientRect().width
        );
        const tbodyWidth = Math.ceil(
          theadCells[i].getBoundingClientRect().width
        );
        theadCells[i].style["max-width"] = theadWidth + "px";
        theadCells[i].style["min-width"] = theadWidth + "px";
        tbodyCells[i].style["max-width"] = tbodyWidth + "px";
        tbodyCells[i].style["min-width"] = tbodyWidth + "px";
      }
      // get new adjusted header cell width based on the corresponding data cell width
      for (let i = 0; i < theadCells.length; i++) {
        const headCellWidth = theadCells[i].offsetWidth;
        const bodyCellWidth = tbodyCells[i].offsetWidth;
        if (headCellWidth < bodyCellWidth) {
          headTargetCells.push({ elem: theadCells[i], width: bodyCellWidth });
        }
      }
      headTargetCells.forEach(setCellWidth);
      // adjust body to fit head
      for (let i = 0; i < theadCells.length; i++) {
        const headCellWidth = theadCells[i].offsetWidth;
        const bodyCellWidth = tbodyCells[i].offsetWidth;
        if (bodyCellWidth < headCellWidth) {
          bodyTargetCells.push({ elem: tbodyCells[i], width: headCellWidth });
        }
      }
      bodyTargetCells.forEach(setCellWidth);
      let sum = 0;
      // check for remainder width and redistribute it evenly among the cells
      for (let i = 0; i < theadCells.length; ++i) {
        sum += theadCells[i].offsetWidth;
      }
      if (sum < this.tbody.clientWidth) {
        const residualSpace =
          (this.tbody.clientWidth - sum) / theadCells.length;
        for (let i = 0; i < theadCells.length; ++i) {
          theadCells[i].style["max-width"] =
            (theadCells[i].offsetWidth + residualSpace).toString() + "px";
          theadCells[i].style["min-width"] =
            (theadCells[i].offsetWidth + residualSpace).toString() + "px";
          tbodyCells[i].style["max-width"] =
            (tbodyCells[i].offsetWidth + residualSpace).toString() + "px";
          tbodyCells[i].style["min-width"] =
            (tbodyCells[i].offsetWidth + residualSpace).toString() + "px";
        }
      }
    },

    onScroll() {
      const scrollLeft = this.tbody && this.tbody.scrollLeft;
      const tableHeaderRow = this.thead && this.thead.querySelector("tr");
      if (!_.isNil(scrollLeft) && tableHeaderRow) {
        tableHeaderRow.style["margin-left"] = scrollLeft
          ? `-${this.tbody.scrollLeft}px`
          : 0;
      }
    },

    setScrollLeft(scrollLeft: number) {
      this.tbody.scrollLeft = scrollLeft;
    },

    onMouseOverTableCell(event) {
      const target = event.target;
      if (!target) {
        return;
      }

      const isCell = target.tagName === "TD";
      const isLeafNode = target.childElementCount === 0;
      const text = target.innerText;
      const title = target.getAttribute("title");

      if (isCell && isLeafNode && text && !title) {
        // set title for displaying full text on hover
        target.setAttribute("title", text);
      }
    },
  },

  data() {
    return {
      table: {} as HTMLTableElement,
      tbody: {} as HTMLTableSectionElement,
      thead: {} as HTMLTableSectionElement,
      tableHeaderRow: {} as HTMLTableRowElement,
    };
  },

  mounted: function () {
    this.table = this.$el.querySelector("table");
    this.tbody = this.$el.querySelector("tbody");
    this.thead = this.$el.querySelector("thead");

    this.tbody.addEventListener("scroll", this.onScroll);

    window.addEventListener("resize", this.resizeTableCells);

    this.checkScrollBar();
    this.resizeTableCells();
  },

  beforeDestroy: function () {
    this.tbody.removeEventListener("scroll", this.onScroll);
    window.removeEventListener("resize", this.resizeTableCells);
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
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex-shrink: 0;
  flex-grow: 1;
}
.fixed-header-table tbody {
  width: 100%;
  overflow-x: scroll;
  overflow-y: auto;
  flex: 1;
}
.fixed-header-table tbody td {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
