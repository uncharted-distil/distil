<template>
  <div>
    <component v-if="facetType" :is="facetType" :data.prop="facetData">
      <div slot="header-label" class="facet-header-container">
        <i
          v-if="facetType === 'facet-terms'"
          :class="getGroupIcon(summary) + ' facet-header-icon'"
        ></i>
        <span>{{ summary.label.toUpperCase() }}</span>
        <type-change-menu
          class="facet-header-dropdown"
          :dataset="summary.dataset"
          :field="summary.key"
          :expandCollapse="expandCollapse"
        >
        </type-change-menu>
      </div>

      <div slot="footer" class="facet-footer-container">
        <div v-if="facetDisplayMore" class="facet-footer-more">
          <span>{{ facetValueCount - getNumToDisplay(summary) }} more</span>
        </div>
        <div
          v-if="this.html"
          v-child="computeCustomHTML()"
          class="facet-footer-custom-html"
        ></div>
      </div>
    </component>
    <div
      class="facets-root"
      v-bind:class="{ 'highlighting-enabled': enableHighlighting }"
    >
      <div class="facet-tooltip" style="display:none;"></div>
    </div>
  </div>
</template>

<script lang="ts">
import "@uncharted/facets-core";
import { FacetBarsData } from "@uncharted/facets-core/dist/types/facet-bars/FacetBars";
import { FacetTermsData } from "@uncharted/facets-core/dist/types/facet-terms/FacetTerms";

import _ from "lodash";
import $ from "jquery";
import Vue from "vue";
import moment from "moment";

import IconFork from "../icons/IconFork.vue";
import IconBookmark from "../icons/IconBookmark.vue";
import { createIcon } from "../../util/icon";
import {
  CATEGORICAL_FILTER,
  NUMERICAL_FILTER,
  DATETIME_FILTER,
  BIVARIATE_FILTER,
  TIMESERIES_FILTER,
  INCLUDE_FILTER
} from "../../util/filters";
import {
  createGroup,
  Group,
  CategoricalFacet,
  isCategoricalFacet,
  getCategoricalChunkSize,
  isNumericalFacet,
  getGroupIcon,
  CATEGORICAL_CHUNK_SIZE
} from "../../util/facets";
import {
  VariableSummary,
  Highlight,
  RowSelection,
  Row,
  Variable,
  CATEGORICAL_SUMMARY,
  NUMERICAL_SUMMARY
} from "../../store/dataset";
import { Dictionary } from "../../util/dict";
import { getSelectedRows } from "../../util/row";
import Facets from "@uncharted.software/stories-facets";
import ImagePreview from "../ImagePreview.vue";
import TypeChangeMenu from "../TypeChangeMenu.vue";
import {
  getVarType,
  isClusterType,
  addClusterPrefix,
  hasComputedVarPrefix,
  GEOCOORDINATE_TYPE,
  DATETIME_UNIX_ADJUSTMENT,
  TIMESERIES_TYPE
} from "../../util/types";
import { IMPORTANT_VARIABLE_RANKING_THRESHOLD } from "../../util/data";
import { getters as datasetGetters } from "../../store/dataset/module";

import "@uncharted.software/stories-facets/dist/facets.css";

export default Vue.extend({
  name: "facet-entry",

  components: {
    TypeChangeMenu
  },

  directives: {
    child(el, binding): void {
      el.innerHTML = "";
      if (binding.value) {
        el.appendChild(binding.value);
      }
    }
  },

  props: {
    summary: Object as () => VariableSummary,
    highlight: Object as () => Highlight,
    rowSelection: Object as () => RowSelection,
    deemphasis: Object as () => any,
    enabledTypeChanges: Array as () => string[],
    enableHighlighting: Boolean as () => boolean,
    showOrigin: Boolean as () => boolean,
    ignoreHighlights: Boolean as () => boolean,
    instanceName: String as () => string,
    ranking: Number as () => number,
    html: [
      String as () => string,
      Object as () => any,
      Function as () => Function
    ],
    expandCollapse: Function as () => Function
  },

  data() {
    const component = this as any;
    return {
      menus: {} as any,
      facets: {} as any,
      numToDisplay: {} as Dictionary<number>,
      numAddedToDisplay: {} as Dictionary<number>
    };
  },

  mounted() {
    const component = this;
    // Instantiate the external facets widget. The facets maintain their own copies
    // of group objects which are replaced wholesale on changes.  Elsewhere in the code
    // we modify local copies of the group objects, then replace those in the Facet component
    // with copies.
    this.facets = new Facets(this.$el, [this.groupSpec]);

    // Call customization hook

    this.injectHTML(
      this.groupSpec,
      this.facets.getGroup(this.groupSpec.key)._element
    );
    this.augmentGroup(this.groupSpec, this.facets.getGroup(this.groupSpec.key));
    this.updateImportantBadge(this.groupSpec);

    // proxy events
    this.facets.on("facet-group:expand", (event: Event, key: string) => {
      component.$emit("expand", this.groupSpec.colName);
    });

    this.facets.on("facet-group:collapse", (event: Event, key: string) => {
      component.$emit("collapse", this.groupSpec.colName);
    });

    this.facets.on(
      "facet-histogram:rangechangeduser",
      (event: Event, key: string, value: any, facet: any) => {
        const range = this.buildNumericalRange(
          value.from.label[0],
          value.to.label[0]
        );
        component.$emit(
          "range-change",
          this.instanceName,
          this.groupSpec.colName,
          range,
          facet.dataset
        );
      }
    );

    // hover over events

    this.facets.on(
      "facet-histogram:mouseenter",
      (event: Event, key: string, value: any) => {
        component.$emit("histogram-mouse-enter", this.groupSpec.colName, value);
      }
    );

    this.facets.on(
      "facet-histogram:mouseleave",
      (event: Event, key: string) => {
        component.$emit("histogram-mouse-leave", this.groupSpec.colName);
      }
    );

    this.facets.on(
      "facet:mouseenter",
      (event: Event, key: string, value: number) => {
        component.$emit("facet-mouse-enter", this.groupSpec.colName, value);
      }
    );

    this.facets.on("facet:mouseleave", (event: Event, key: string) => {
      component.$emit("facet-mouse-leave", this.groupSpec.colName);
    });

    // more events

    this.facets.on("facet-group:more", (event: Event, key: string) => {
      const chunkSize = getCategoricalChunkSize(this.groupSpec.type);
      if (!component.numToDisplay[key]) {
        Vue.set(component.numToDisplay, key, chunkSize);
        Vue.set(component.numAddedToDisplay, key, 0);
      }
      Vue.set(
        component.numToDisplay,
        key,
        component.numToDisplay[key] + chunkSize
      );
      Vue.set(
        component.numAddedToDisplay,
        key,
        component.numAddedToDisplay[key] + chunkSize
      );
      component.$emit("facet-more", key);
    });

    this.facets.on("facet-group:less", (event: Event, key: string) => {
      const chunkSize = getCategoricalChunkSize(this.groupSpec.type);
      if (!component.numToDisplay[key]) {
        Vue.set(component.numToDisplay, key, chunkSize);
        Vue.set(component.numAddedToDisplay, key, 0);
      }
      Vue.set(
        component.numToDisplay,
        key,
        component.numToDisplay[key] - chunkSize
      );
      Vue.set(
        component.numAddedToDisplay,
        key,
        component.numAddedToDisplay[key] - chunkSize
      );
      component.$emit("facet-less", key);
    });

    // click events

    this.facets.on(
      "facet-histogram:click",
      (event: Event, key: string, value: any, facet: any) => {
        // if this is a click on value previously used as highlight root, clear
        const range = {
          from: _.toNumber(value.label),
          to: _.toNumber(value.toLabel)
        };
        if (
          this.isHighlightedValue(this.highlight, this.groupSpec.colName, range)
        ) {
          // clear current selection
          component.$emit(
            "histogram-click",
            this.instanceName,
            null,
            null,
            facet.dataset
          );
        } else {
          // set selection
          component.$emit(
            "histogram-click",
            this.instanceName,
            this.groupSpec.colName,
            range,
            facet.dataset
          );
        }
      }
    );

    this.facets.on(
      "facet:click",
      (event: Event, key: string, value: string, count: number, facet: any) => {
        // User clicked on the value that is currently the highlight root
        if (
          this.isHighlightedValue(this.highlight, this.groupSpec.colName, value)
        ) {
          // clear current selection
          component.$emit(
            "facet-click",
            this.instanceName,
            null,
            null,
            facet.dataset
          );
        } else {
          // set selection
          component.$emit(
            "facet-click",
            this.instanceName,
            this.groupSpec.colName,
            value,
            facet.dataset
          );
        }
      }
    );

    this.facets.on(
      "facet-histogram:mouseenter",
      (event: Event, key: string, value: any) => {
        const $target = $(event.target);
        const $parent = $target.parent();
        const $root = $(this.$el);
        const $tooltip = $(this.$el).find(".facet-tooltip");
        const TOOLTIP_BUFFER = 8;

        // set the text now so that the dimensions are correct
        $tooltip.html(`<b>${value.label} - ${value.toLabel}</b>`);
        const posX = $parent.offset().left - $root.offset().left;
        const posY =
          $parent.offset().top - $root.offset().top + $root.scrollTop();
        const offsetX =
          posX + $target.outerWidth() / 2 - $tooltip.outerWidth() / 2;
        const offsetY = posY - $tooltip.height() - TOOLTIP_BUFFER;

        // ensure it doesnt go outside of root
        const x = Math.min($root.width(), Math.max(0, offsetX));
        const y = Math.max(0, offsetY);

        $tooltip.css("left", `${x}px`);
        $tooltip.css("top", `${y}px`);
        $tooltip.show();
      }
    );

    this.facets.on(
      "facet-histogram:mouseleave",
      (event: Event, key: string, value: string) => {
        $(this.$el)
          .find(".facet-tooltip")
          .hide();
      }
    );

    this.injectHighlights(this.highlight, this.rowSelection, this.deemphasis);
  },

  computed: {
    groupSpec(): Group {
      const group = createGroup(this.summary);

      const numToDisplay = this.numToDisplay[group.key];
      if (numToDisplay) {
        const show = group.all.slice(0, numToDisplay);
        const hide = group.all.slice(numToDisplay);
        group.facets = show;

        let remainingTotal = 0;
        hide.forEach(facet => {
          if (isCategoricalFacet(facet)) {
            remainingTotal += facet.count;
          }
        });
        group.more = group.all.length - numToDisplay;
        group.moreTotal = remainingTotal;

        // track how many are already added
        if (
          this.numAddedToDisplay[group.key] &&
          this.numAddedToDisplay[group.key] > 0
        ) {
          group.less = this.numAddedToDisplay[group.key];
        }
      }

      // TODO: move this reference to highlights, as it forces a refresh

      if (
        this.enableHighlighting &&
        this.highlight &&
        this.highlight.context === this.instanceName
      ) {
        if (group.colName === this.highlight.key) {
          group.facets.forEach(facet => {
            facet.filterable = true;
          });
        }
      }

      if (this.showOrigin) {
        group.facets.forEach((facet: any) => {
          if (facet.histogram) {
            facet.histogram.showOrigin = true;
          }
        });
      }

      return group;
    },

    facetType(): string {
      const summary = this.summary;
      if (!summary.err && !summary.pending) {
        switch (summary.type) {
          case CATEGORICAL_SUMMARY:
            if (summary.varType === TIMESERIES_TYPE) {
              // not implemented yet
              break;
            } else {
              return "facet-terms";
            }
          case NUMERICAL_SUMMARY:
            return "facet-bars";
          default:
            break;
        }
      }
      return null;
    },

    facetData(): FacetBarsData | FacetTermsData | null {
      if (this.facetType === "facet-terms") {
        return this.computeTermsData(this.summary);
      } else if (this.facetType === "facet-bars") {
        return this.computeBarsData(this.summary);
      }
      return null;
    },

    facetValueCount(): number {
      return this.summary.baseline.buckets.length;
    },

    facetDisplayMore(): boolean {
      if (this.facetType === "facet-terms") {
        const chunkSize = getCategoricalChunkSize(this.summary.type);
        return this.facetValueCount > chunkSize;
      }
      return false;
    }
  },

  watch: {
    // handle changes to the facet group list
    groupSpec: {
      handler(currGroup: Group, prevGroup: Group) {
        // update and groups
        this.updateGroups(currGroup, prevGroup);
        this.injectHighlights(
          this.highlight,
          this.rowSelection,
          this.deemphasis
        );
      },
      deep: true
    },

    ranking(currRanking: number, prevRanking: number) {
      if (currRanking !== prevRanking) {
        this.updateImportantBadge(this.groupSpec);
      }
    },

    // handle external highlight changes by updating internal facet select states
    highlight(currHighlights: Highlight) {
      if (this.enableHighlighting) {
        this.addHighlightArrow(currHighlights);
      }
    },

    // handle external highlight changes by updating internal facet select states
    rowSelection(currSelection: RowSelection) {
      this.injectHighlights(this.highlight, currSelection, this.deemphasis);
    },

    deemphasis(currDemphasis: any) {
      this.injectHighlights(this.highlight, this.rowSelection, currDemphasis);
    },
    html(customHtml: () => HTMLElement) {
      const $elem = this.facets.getGroup(this.groupSpec.key)._element;
      this.updateCustomHtml(this.groupSpec, $elem, customHtml);
    }
  },

  methods: {
    getNumToDisplay(summary: VariableSummary): number {
      return (
        this.numToDisplay[summary.key] || getCategoricalChunkSize(summary.type)
      );
    },

    computeTermsData(summary: VariableSummary): FacetTermsData {
      const numToDisplay = this.getNumToDisplay(summary);
      const values = [];
      if (summary.baseline.buckets.length) {
        const buckets = summary.baseline.buckets;
        const maxCount = summary.baseline.extrema.max;
        for (
          let i = 0, n = Math.min(buckets.length, numToDisplay);
          i < n;
          ++i
        ) {
          values.push({
            ratio: buckets[i].count / maxCount,
            label: buckets[i].key,
            value: buckets[i].count
          });
        }
      }
      return {
        label: summary.label.toUpperCase(),
        values
      };
    },

    computeBarsData(summary: VariableSummary): FacetBarsData {
      const values = [];
      if (summary.baseline.buckets.length) {
        const buckets = summary.baseline.buckets;
        // seems to be incorrect compute based on the current buckets
        // const maxCount = summary.baseline.extrema.max;
        const maxCount = buckets.reduce(
          (max, bucket) => Math.max(max, bucket.count),
          0
        );
        for (let i = 0, n = buckets.length; i < n; ++i) {
          values.push({
            ratio: buckets[i].count / maxCount,
            label: buckets[i].key
          });
        }
      }
      return {
        label: summary.label.toUpperCase(),
        values
      };
    },

    computeCustomHTML(): HTMLElement | null {
      // hack to get the custom html buttons showing up
      // changing this would mean to change how the instantiation of the facets works
      // right now they are wrapped by other components like
      // available-target-variables, available-training-variables, etc
      // those components inject HTML into the facets through their `html` function
      // we might want to change that in the future though
      if (this.html) {
        return _.isFunction(this.html)
          ? this.html({
              colName: this.summary.key
            })
          : this.html;
      }
      return null;
    },

    getGroupIcon,

    injectHTML(group: Group, $elem: JQuery) {
      const $groupFooter = $(
        '<div class="group-footer"><div class="html-slot"></div></div>'
      ).appendTo($elem.find(".facets-group"));
      $elem.click(event => {
        if (group.facets.length >= 1) {
          const facet = group.facets[0];
          if (isNumericalFacet(facet)) {
            const slices = facet.histogram.slices;
            const first = slices[0];
            const last = slices[slices.length - 1];
            const range = this.buildNumericalRange(first.label, last.toLabel);
            this.$emit(
              "numerical-click",
              this.instanceName,
              group.colName,
              range,
              group.dataset
            );
          } else if (isCategoricalFacet(facet)) {
            this.$emit(
              "categorical-click",
              this.instanceName,
              group.colName,
              null,
              group.dataset
            );
          }
        }
      });

      $elem.find(".facet-histogram g").click(event => {
        if (group.facets.length >= 1) {
          const facet = group.facets[0];

          if (isNumericalFacet(facet)) {
            const slices = facet.histogram.slices;
            const first = slices[0];
            const last = slices[slices.length - 1];
            const range = this.buildNumericalRange(first.label, last.toLabel);
            this.$emit(
              "numerical-click",
              this.instanceName,
              group.colName,
              range,
              group.dataset
            );
          }
        }
      });
      // inject type icon in group header
      this.wrapGroupHeaderText(group, $elem);

      // inject type icon in group header
      this.injectTypeIcon(group, $elem);

      // inject type change header menus
      this.injectTypeChangeHeaders(group, $elem);

      // inject html
      this.updateCustomHtml(group, $elem, this.html);

      // inject image preview if image type
      this.injectImagePreview(group, $elem);

      this.injectImportantBadge(group, $elem);
    },

    updateCustomHtml(group: Group, $elem: JQuery, html: any) {
      if (html) {
        const $htmlSlot = $elem.find(".html-slot");
        const customHtml = _.isFunction(this.html)
          ? this.html(group)
          : this.html;
        $htmlSlot.empty();
        $htmlSlot.append(customHtml);
        this.$emit("html-appended", customHtml);
      }
    },

    addHighlightArrow(highlight: Highlight) {
      const $elem = $(this.$el);
      // remove previous
      $elem.find(".highlight-arrow").remove();

      // NOTE: first group is a query group, ignore it
      const QUERY_OFFSET = 1;

      const $groups = $elem.find(".facets-group");
      // add highlight arrow
      if (this.isHighlightedGroup(highlight, this.groupSpec.colName)) {
        const $group = $($groups.get(QUERY_OFFSET));
        $group.append(
          '<div class="highlight-arrow"><i class="fa fa-arrow-circle-right fa-2x"></i></div>'
        );
      }
    },

    isHighlightedInstance(highlight: Highlight): boolean {
      return highlight && highlight.context === this.instanceName;
    },

    isHighlightedGroup(highlight: Highlight, colName: string): boolean {
      return this.isHighlightedInstance(highlight) && highlight.key === colName;
    },

    isHighlightedFacet(highlight: Highlight, facet: any): boolean {
      const highlightValue = this.getHighlightValue(highlight);
      return (
        highlightValue &&
        facet &&
        facet.value &&
        typeof highlightValue === "string" &&
        typeof facet.value === "string" &&
        highlightValue.toLowerCase() === facet.value.toLowerCase()
      );
    },

    isHighlightedValue(
      highlight: Highlight,
      colName: string,
      value: any
    ): boolean {
      // if not instance, return false
      if (!this.isHighlightedGroup(highlight, colName)) {
        return false;
      }
      if (_.isArray(highlight.value)) {
        return false;
      }
      // if string, check for match
      if (_.isString(highlight.value)) {
        return highlight.value === value;
      }

      // if datetime, convert value
      let fromValue = value.from;
      let toValue = value.to;
      if (highlight.value.type === DATETIME_FILTER) {
        fromValue = Date.parse(value.from) / DATETIME_UNIX_ADJUSTMENT;
        toValue = Date.parse(value.to) / DATETIME_UNIX_ADJUSTMENT;
      }

      // otherwise, check range
      return (
        highlight.value.from === fromValue && highlight.value.to === toValue
      );
    },

    getHighlightValue(highlight: Highlight): any {
      if (highlight && highlight.value) {
        return highlight.value;
      }
      return null;
    },

    selectCategoricalFacet(facet: any, count?: number) {
      if (
        count === undefined &&
        facet._spec.segments &&
        facet._spec.segments.length > 0
      ) {
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

    selectTimeseriesFacet(facet: any, count?: number) {
      facet._sparklineContainer.parent().addClass("facet-sparkline-selected");
    },

    deselectTimeseriesFacet(facet: any) {
      facet._sparklineContainer
        .parent()
        .removeClass("facet-sparkline-selected");
    },

    injectDeemphasis(group: any, deemphasis: any) {
      if (!deemphasis) {
        return;
      }

      // clear deemphasis
      for (const facet of group.facets) {
        // only emphasize histograms
        if (facet._histogram) {
          facet._histogram.bars.forEach(bar => {
            if (!bar._element.hasClass("row-selected")) {
              // NOTE: don't trample row selections
              bar._element.css("fill", "");
            }
          });
        }
      }

      // de-emphasis within range
      for (const facet of group.facets) {
        // only emphasize histograms
        if (facet._histogram) {
          facet._histogram.bars.forEach(bar => {
            const entry: any = _.last(bar.metadata);
            if (
              _.toNumber(entry.label) >= deemphasis.min &&
              _.toNumber(entry.toLabel) < deemphasis.max
            ) {
              if (!bar._element.hasClass("row-selected")) {
                // NOTE: don't trample row selections
                bar._element.css("fill", "#ddd");
              }
            }
          });
        }
      }
    },

    findClusterCol(key: string, row: Row) {
      const clustered = addClusterPrefix(key);
      return _.find(row.cols, c => {
        return c.key === clustered;
      });
    },

    addRowSelectionToFacet(facet: any, col: any) {
      if (facet._histogram) {
        facet._histogram.bars.forEach(bar => {
          const entry: any = _.last(bar.metadata);
          if (!_.isNaN(_.toNumber(entry.label))) {
            if (
              col.value.value >= _.toNumber(entry.label) &&
              col.value.value < _.toNumber(entry.toLabel)
            ) {
              bar._element.css("fill", "#ff0067");
              bar._element.addClass("row-selected");
            }
          } else {
            // datetime labels
            const dateString = moment(col.value).format("YYYY/MM/DD");
            if (dateString >= entry.label && dateString < entry.toLabel) {
              bar._element.css("fill", "#ff0067");
              bar._element.addClass("row-selected");
            }
          }
        });
      } else if (facet._sparkline) {
        // TODO: sparkline
      } else {
        facet._sparklineContainer
          .parent()
          .css("box-shadow", "inset 0 0 0 1000px rgba(255,0,103,.2)");
        facet._barForeground.css("box-shadow", "inset 0 0 0 1000px #ff0067");
      }
    },

    removeRowSelectionFromFacet(facet: any) {
      if (facet._histogram) {
        facet._histogram.bars.forEach(bar => {
          bar._element.css("fill", "");
          bar._element.removeClass("row-selected");
        });
      } else if (facet._sparkline) {
        // TODO: sparkline
      } else {
        facet._barForeground.css("box-shadow", "");
        facet._sparklineContainer.parent().css("box-shadow", "");
      }
    },

    injectSelectedRowIntoGroup(group: any, selection: RowSelection) {
      // clear existing selections
      for (const facet of group.facets) {
        // ignore placeholder facets
        if (facet._type === "placeholder") {
          continue;
        }

        this.removeRowSelectionFromFacet(facet);
      }

      // if no selection, exit early
      if (!selection || selection.d3mIndices.length === 0) {
        return;
      }

      const rows = getSelectedRows(selection);
      rows.forEach(row => {
        // get col
        const col = _.find(row.cols, c => {
          return c.key === this.groupSpec.colName;
        });

        // no matching col and not a cluster with effectively 2+ types, exit early
        if (!col) {
          return;
        }

        for (const facet of group.facets) {
          // ignore placeholder facets
          if (facet._type === "placeholder") {
            continue;
          }

          if (facet._histogram) {
            this.addRowSelectionToFacet(facet, col);
          } else if (facet._sparkline) {
            // TODO: sparkline
          } else {
            if (isClusterType(this.groupSpec.type)) {
              const clusterCol = this.findClusterCol(
                this.groupSpec.colName,
                row
              );
              if (
                facet.value === clusterCol.value ||
                (clusterCol.value.value &&
                  facet.value === clusterCol.value.value)
              ) {
                this.addRowSelectionToFacet(facet, col);
              }
              continue;
            }

            if (facet.value === col.value.value) {
              this.addRowSelectionToFacet(facet, col);
            }
          }
        }
      });
    },

    injectHighlightDatasetDeemphasis(group: any, highlight: Highlight) {
      // if the dataset of the highlight does not match the dataset of this
      // facet, deemphasis the group

      if (!highlight || !group || highlight.dataset === group.dataset) {
        group._element.removeClass("deemphasis");
        return;
      }
      if (highlight.dataset !== group.dataset) {
        group._element.addClass("deemphasis");
      }
    },

    injectHighlightsIntoGroup(group: any, highlight: Highlight) {
      if (this.ignoreHighlights) {
        return;
      }

      // loop through groups ensure that selection is clear on each
      group.facets.forEach(facet => {
        if (facet._type === "placeholder") {
          return;
        }
        if (facet._histogram) {
          facet.deselect();
        } else if (facet._sparkline) {
          if (this.isHighlightedFacet(highlight, facet)) {
            this.selectTimeseriesFacet(facet);
          } else {
            this.deselectTimeseriesFacet(facet);
          }
        } else {
          this.selectCategoricalFacet(facet);
        }
      });

      const highlightRootValue = this.getHighlightValue(highlight);
      const highlightSummary = this.groupSpec.summary
        ? this.groupSpec.summary.filtered
        : null;
      const isHighlightedGroup = this.isHighlightedGroup(
        highlight,
        this.groupSpec.colName
      );

      for (const facet of group.facets) {
        // ignore placeholder facets
        if (facet._type === "placeholder") {
          continue;
        }

        if (facet._histogram) {
          const selection = {} as any;

          // if this is the highlighted group, create filter selection
          if (isHighlightedGroup) {
            // NOTE: the `from` / `to` values MUST be strings.
            // if datetime, need to get date label back.
            selection.range = {
              from:
                highlightRootValue &&
                highlightRootValue.type === DATETIME_FILTER
                  ? moment
                      .unix(highlightRootValue.from)
                      .utc()
                      .format("YYYY/MM/DD")
                  : `${highlightRootValue.from}`,
              to:
                highlightRootValue &&
                highlightRootValue.type === DATETIME_FILTER
                  ? moment
                      .unix(highlightRootValue.to)
                      .utc()
                      .format("YYYY/MM/DD")
                  : `${highlightRootValue.to}`
            };
          } else {
            const bars = facet._histogram.bars;

            if (
              highlightSummary &&
              highlightSummary.buckets.length === bars.length
            ) {
              const slices = {};

              highlightSummary.buckets.forEach((bucket, index) => {
                const entry: any = _.last(bars[index].metadata);
                slices[entry.label] = bucket.count;
              });

              selection.slices = slices;
            }
          }

          facet.select({
            selection: selection
          });
        } else if (facet._sparkline) {
          if (this.isHighlightedFacet(highlight, facet)) {
            this.selectCategoricalFacet(facet);
            this.selectTimeseriesFacet(facet);
          } else {
            this.deselectCategoricalFacet(facet);
            this.deselectTimeseriesFacet(facet);
          }
        } else {
          if (isHighlightedGroup) {
            if (this.isHighlightedFacet(highlight, facet)) {
              this.selectCategoricalFacet(facet);
              this.selectTimeseriesFacet(facet);
            } else {
              this.deselectCategoricalFacet(facet);
              this.deselectTimeseriesFacet(facet);
            }
          } else {
            if (highlightSummary) {
              const bucket = _.find(highlightSummary.buckets, b => {
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

    injectHighlights(
      highlight: Highlight,
      selection: RowSelection,
      deemphasis: any
    ) {
      // Clear highlight state incase it was set via a click on on another
      // component
      $(this.$el)
        .find(".select-highlight")
        .removeClass("select-highlight");
      $(this.$el)
        .find(".facet-sparkline-selected")
        .removeClass("facet-sparkline-selected");
      // Update highlight
      const group = this.facets.getGroup(this.groupSpec.key);
      if (!group) {
        return;
      }
      this.injectHighlightsIntoGroup(group, highlight);
      this.injectHighlightDatasetDeemphasis(group, highlight);
      this.injectSelectedRowIntoGroup(group, selection);
    },

    augmentGroup(distilGroup: Group, facetsGroup: any) {
      // inject any custom properties required for the distil app
      facetsGroup.dataset = distilGroup.dataset;
      facetsGroup.facets.forEach(f => {
        f.dataset = distilGroup.dataset;
      });
    },

    buildNumericalRange(
      fromValue: string,
      toValue: string
    ): { from: number; to: number; type: string } {
      const isNumber = !_.isNaN(_.toNumber(fromValue));
      const range = {
        from: isNumber
          ? _.toNumber(fromValue)
          : Date.parse(fromValue) / DATETIME_UNIX_ADJUSTMENT,
        to: isNumber
          ? _.toNumber(toValue)
          : Date.parse(toValue) / DATETIME_UNIX_ADJUSTMENT,
        type: isNumber ? NUMERICAL_FILTER : DATETIME_FILTER
      };
      return range;
    },

    groupsEqual(a: Group, b: Group): boolean {
      const OMITTED_FIELDS = ["selection", "selected"];
      // NOTE: we dont need to check key, we assume its already equal
      if (a.label !== b.label) {
        return false;
      }
      if (a.facets.length !== b.facets.length) {
        return false;
      }
      for (let i = 0; i < a.facets.length; i++) {
        if (
          !_.isEqual(
            _.omit(a.facets[i], OMITTED_FIELDS),
            _.omit(b.facets[i], OMITTED_FIELDS)
          )
        ) {
          return false;
        }
      }
      return true;
    },

    updateGroups(currGroup: Group, prevGroup: Group) {
      const groupSpec = _.cloneDeep(currGroup);

      // check if it already exists
      if (prevGroup) {
        // check if equal, if so, no need to change
        if (this.groupsEqual(currGroup, prevGroup)) {
          return;
        }
        // replace group if it is existing
        this.facets.replaceGroup(groupSpec);
        this.augmentGroup(currGroup, this.facets.getGroup(currGroup.key));
        this.injectHTML(
          currGroup,
          this.facets.getGroup(currGroup.key)._element
        );
        this.injectHighlightsIntoGroup(
          this.facets.getGroup(currGroup.key),
          this.highlight
        );
        this.injectHighlightDatasetDeemphasis(
          this.facets.getGroup(currGroup.key),
          this.highlight
        );
        this.injectSelectedRowIntoGroup(
          this.facets.getGroup(currGroup.key),
          this.rowSelection
        );
        this.injectDeemphasis(
          this.facets.getGroup(currGroup.key),
          this.deemphasis
        );
      } else {
        // add to appends
        this.facets.append(groupSpec);
        this.augmentGroup(currGroup, this.facets.getGroup(currGroup.key));
        this.injectHTML(
          currGroup,
          this.facets.getGroup(currGroup.key)._element
        );
        this.injectHighlightsIntoGroup(
          this.facets.getGroup(currGroup.key),
          this.highlight
        );
        this.injectHighlightDatasetDeemphasis(
          this.facets.getGroup(currGroup.key),
          this.highlight
        );
        this.injectSelectedRowIntoGroup(
          this.facets.getGroup(currGroup.key),
          this.rowSelection
        );
        this.injectDeemphasis(
          this.facets.getGroup(currGroup.key),
          this.deemphasis
        );
      }
      this.updateImportantBadge(currGroup);
    },

    updateImportantBadge(group: Group) {
      // update 'important' class
      const $group = this.facets.getGroup(group.key)._element;
      const isImportant = this.ranking > IMPORTANT_VARIABLE_RANKING_THRESHOLD;
      $group.toggleClass("important", Boolean(isImportant));
    },

    // wrap group header text
    wrapGroupHeaderText(group: Group, $elem: JQuery) {
      const $headerElement = $elem.find(".group-header");
      const headerText = $headerElement.text();
      const tooltipText = (this.summary.description
        ? headerText.concat(": ", this.summary.description)
        : headerText
      ).replace(/(\r\n|\n|\r|\t)/gm, "");
      $headerElement.empty();
      const $headerTextWrapped = $(
        `<div class="header-text" v-b-tooltip.hover title="${tooltipText}">${headerText}</div>`
      );
      $headerElement.append($headerTextWrapped);
    },

    // inject type icon
    injectTypeIcon(group: Group, $elem: JQuery) {
      if (isCategoricalFacet(group.facets.length > 0 && group.facets[0])) {
        const facetSpecs = <CategoricalFacet[]>group.facets;
        const typeicon = facetSpecs[0].icon.class;
        const $icon = $(`<i class="${typeicon}"></i>`);
        $elem.find(".group-header").append($icon);
      }
      if (hasComputedVarPrefix(group.colName)) {
        const $forkIcon = createIcon(IconFork);
        $elem.find(".group-header").append($forkIcon);
      }
    },

    // inject type headers
    injectTypeChangeHeaders(group: Group, $elem: JQuery) {
      const facetId = `${group.dataset}:${group.colName}`;
      if (this.enabledTypeChanges.find(e => e === facetId)) {
        // if we have a menu for this already, destroy it to replace it
        if (this.menus[facetId]) {
          this.menus[facetId].$destroy();
        }
        const $slot = $("<span/>");
        $elem.find(".group-header").append($slot);
        const menu = new TypeChangeMenu({
          store: this.$store,
          router: this.$router,
          propsData: {
            dataset: group.dataset,
            field: group.colName,
            expandCollapse: this.expandCollapse
          }
        });
        menu.$mount($slot[0]);
        this.menus[facetId] = menu;
      }
    },

    injectImagePreview(group: Group, $elem: JQuery) {
      if (group.type === "image") {
        const $facets = $elem.find(".facet-block");
        group.facets.forEach((facet: any, index) => {
          const $facet = $($facets.get(index));
          const $slot = $("<span/>");
          $facet.append($slot);
          const preview = new ImagePreview({
            store: this.$store,
            router: this.$router,
            propsData: {
              // NOTE: there seems to be an issue with the visibility plugin used
              // when injecting this way. Cancel the visibility flagging for facets.
              preventHiding: true,
              imageUrl: facet.file || facet.value
            }
          });
          preview.$mount($slot[0]);
        });
      }
    },
    injectImportantBadge(group: Group, $elem: JQuery) {
      const $groupFooter = $elem.find(".group-footer").find(".html-slot");
      const importantBadge = document.createElement("div");
      importantBadge.className += "important-badge";
      const $bookMarkIcon = createIcon(IconBookmark);
      importantBadge.append($bookMarkIcon);
      $groupFooter.append(importantBadge);
    }
  },

  destroyed: function() {
    this.facets.destroy();
    this.facets = null;

    // specifically destroy menus so because we injected them
    // and so we have to take manual action to destroy them
    _.forIn(this.menus, menu => menu.$destroy());
    this.menus = null;
  }
});
</script>

<style>
.facet-header-container {
  display: flex;
  align-items: center;
}

.facet-header-icon {
  margin-right: 4px;
}

.facet-header-dropdown {
  position: absolute;
  right: 12px;
}

.facet-footer-container {
  min-height: 12px;
  padding: 6px 12px 5px;
  font-family: "IBM Plex Sans", sans-serif;
  font-size: 12px;
  font-weight: 600;
  line-height: 16px;
}

.facet-footer-more {
  margin-bottom: 4px;
}

.facet-footer-custom-html {
  margin-top: 6px;
}

.group-facet-container {
  position: relative;
}

.facet-highlight-spinner {
  position: absolute;
  right: 0;
  bottom: 0;
  margin-right: 8px;
  margin-bottom: 6px;
}

.facets-root {
  position: relative;
}

.facets-group-container.deemphasis {
  opacity: 0.5;
}
.facets-root.highlighting-enabled {
  padding-left: 32px;
}
.highlighting-enabled .facets-group {
  cursor: pointer !important;
  border: 1px solid rgba(0, 0, 0, 0);
}
.highlighting-enabled .facets-group:hover {
  border: 1px solid #00c6e1;
}

.highlighting-enabled .group-header,
.highlighting-enabled .facet-range-controls,
.highlighting-enabled .facets-facet-horizontal {
  cursor: pointer !important;
}

.facets-facet-horizontal .facet-histogram-bar-transform {
  transition: none;
}

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
  color: rgba(0, 0, 0, 0.54);
}
.facets-group .group-header {
  font-family: inherit;
  font-size: 0.867rem;
  font-weight: bold;
  text-transform: uppercase;
  word-wrap: break-word;
  word-break: break-all;
  color: rgba(0, 0, 0, 0.54);
}
.facets-group .group-header i {
  margin-left: 5px;
}
.facets-group .group-header .svg-icon {
  height: 14px;
  margin-left: 2px;
}
/* .facets-group .group-footer {
	display: flex;
} */
.facets-group .group-footer .important-badge {
  display: none;
}
.facets-group .group-facet-container {
  width: 100%;
  max-height: 240px;
  overflow-y: auto;
  overflow-x: hidden;
}
.facets-group-container.important .group-footer .html-slot .important-badge {
  display: block;
  position: absolute;
  bottom: 5px;
  right: 5px;
}

.facets-facet-horizontal .select-highlight,
.facets-facet-horizontal .facet-histogram-bar-highlighted.select-highlight {
  fill: #007bff;
}

.excluded .facet-range-filter {
  box-shadow: inset 0 0 0 1000px rgba(220, 53, 0, 0.15) !important;
}
.excluded .facets-facet-horizontal .facet-histogram-bar-highlighted {
  fill: #333;
}
.excluded .facet-bar-base.facet-bar-selected {
  box-shadow: inset 0 0 0 1000px #333;
}
.facet-histogram {
  cursor: pointer !important;
}

.facets-facet-vertical.select-highlight .facet-bar-selected {
  box-shadow: inset 0 0 0 1000px #007bff;
}

.highlight-arrow {
  position: absolute;
  left: -28px;
  top: calc(50% - 14px);
  color: #00c6e1;
}
.facet-tooltip {
  position: absolute;
  padding: 4px 8px;
  border-radius: 4px;
  background-color: #333;
  color: #fff;
  z-index: 2;
  pointer-events: none;
}
.facet-tooltip::after {
  content: " ";
  position: absolute;
  top: 100%; /* At the bottom of the tooltip */
  left: 50%;
  margin-left: -5px;
  border-width: 5px;
  border-style: solid;
  border-color: #333 transparent transparent transparent;
}
.facet-sparkline,
.facet-sparkline-container {
  overflow: visible !important;
}
.facet-block.facet-sparkline-selected {
  background-color: rgba(0, 198, 225, 0.2);
}
</style>
