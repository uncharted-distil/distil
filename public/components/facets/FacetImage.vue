<template>
  <facet-terms
    :data.prop="facetData"
    action-buttons="0"
    :selection.prop="selection"
    :subselection.prop="subSelection"
    :disabled.prop="!enableHighlighting"
    @facet-element-updated="updateSelection"
    class="facet-image"
  >
    <div slot="header-label" :class="headerClass">
      <i :class="getGroupIcon(summary) + ' facet-header-icon'"></i>
      <span>{{ summary.label.toUpperCase() }}</span>
      <type-change-menu
        v-if="facetEnableTypeChanges"
        class="facet-header-dropdown"
        :dataset="summary.dataset"
        :field="summary.key"
        :expandCollapse="expandCollapse"
      />
    </div>
    <facet-template target="facet-terms-value" class="facet-content-container">
      <div slot="header" class="facet-image-preview-display">${metadata}</div>
      <div slot="label" class="facet-image-label" title="${value} ${label}">
        ${value} ${label}
      </div>
      <div slot="annotation" class="collapse-unused" />
      <div slot="value" class="collapse-unused" />
    </facet-template>
    <div slot="footer" class="facet-footer-container">
      <div v-if="facetDisplayMore" class="facet-footer-more">
        <div class="facet-footer-more-section">
          <div class="facet-footer-more-count">
            <span v-if="facetMoreCount > 0">{{ facetMoreCount }} more</span>
          </div>
          <div class="facet-footer-more-controls">
            <span v-if="hasLess" @click="viewLess"> show less</span>
            <span v-if="hasMore" @click="viewMore"> show more</span>
          </div>
        </div>
      </div>
      <div
        v-if="this.html"
        v-child="computeCustomHTML()"
        class="facet-footer-custom-html"
      ></div>
    </div>
  </facet-terms>
</template>

<script lang="ts">
import Vue from "vue";

import "@uncharted.software/facets-core";
import { FacetTermsData } from "@uncharted.software/facets-core/dist/types/facet-terms/FacetTerms";

import TypeChangeMenu from "../TypeChangeMenu.vue";
import ImagePreview from "../ImagePreview.vue";
import { Highlight, RowSelection, VariableSummary } from "../../store/dataset";
import {
  getCategoricalChunkSize,
  getGroupIcon,
  getSubSelectionValues,
  hasBaseline,
  viewMoreData,
  viewLessData,
  facetTypeChangeState,
} from "../../util/facets";
import _ from "lodash";
import { IMAGE_TYPE } from "../../util/types";

export default Vue.extend({
  name: "facet-image",

  components: {
    TypeChangeMenu,
    ImagePreview,
  },

  directives: {
    child(el, binding): void {
      el.innerHTML = "";
      if (binding.value) {
        el.appendChild(binding.value);
      }
    },
  },

  props: {
    summary: Object as () => VariableSummary,
    enabledTypeChanges: Array as () => string[],
    html: [
      String as () => string,
      Object as () => any,
      Function as () => Function,
    ],
    expandCollapse: Function as () => Function,
    highlight: Object as () => Highlight,
    enableHighlighting: Boolean as () => boolean,
    instanceName: String as () => string,
    rowSelection: Object as () => RowSelection,
  },

  data() {
    return {
      baseNumToDisplay: getCategoricalChunkSize(this.summary.type),
      moreNumToDisplay: 0,
    };
  },

  computed: {
    facetData(): FacetTermsData {
      let values = [];
      if (hasBaseline(this.summary)) {
        values = this.getFacetValues();
      }
      return {
        label: this.summary.label.toUpperCase(),
        values,
      };
    },
    facetEnableTypeChanges(): boolean {
      return facetTypeChangeState(
        this.summary.dataset,
        this.summary.key,
        this.enabledTypeChanges
      );
    },
    headerClass(): string {
      return this.facetEnableTypeChanges
        ? "facet-header-container"
        : "facet-header-container-no-scroll";
    },
    subSelection(): number[][] {
      return getSubSelectionValues(this.summary, this.rowSelection, this.max);
    },
    selection(): {} {
      if (!this.isHighlightedGroup(this.highlight, this.summary.key)) {
        return null;
      }
      const highlightValue = this.getHighlightValue(this.highlight);
      if (!highlightValue) {
        return null;
      }
      const highlightAsSelection = this.summary.baseline.buckets.reduce(
        (acc, val, ind) => {
          if (val.key === highlightValue) acc[ind] = true;
          return acc;
        },
        {}
      );
      return highlightAsSelection;
    },
    facetValueCount(): number {
      return this.hasBaseline ? this.summary.baseline.buckets.length : 0;
    },
    facetDisplayMore(): boolean {
      const chunkSize = getCategoricalChunkSize(this.summary.type);
      return this.facetValueCount > chunkSize;
    },
    facetMoreCount(): number {
      return this.facetValueCount - this.numToDisplay;
    },
    numToDisplay(): number {
      return this.hasExamplars &&
        this.summary.baseline.exemplars.length < this.baseNumToDisplay
        ? this.summary.baseline.exemplars.length
        : this.hasBaseline && this.facetValueCount < this.baseNumToDisplay
        ? this.facetValueCount
        : this.baseNumToDisplay + this.moreNumToDisplay;
    },
    max(): number {
      return this.hasBaseline ? this.summary.baseline.extrema.max : 0;
    },
    hasExamplars(): boolean {
      return !!this.summary.baseline.exemplars;
    },
    hasBaseline(): boolean {
      return hasBaseline(this.summary);
    },
    hasMore(): boolean {
      return this.numToDisplay < this.facetValueCount;
    },
    hasLess(): boolean {
      return this.moreNumToDisplay > 0;
    },
  },

  methods: {
    getFacetValues(): {
      ratio: number;
      label: string;
      value: number;
      metadata: {};
    }[] {
      const summary = this.summary;
      const buckets = summary.baseline.buckets;
      const facetData = [];
      for (let i = 0; i < this.numToDisplay; ++i) {
        const imageUrl = this.hasExamplars
          ? summary.baseline.exemplars[i]
          : buckets[i].key;
        facetData.push({
          ratio: buckets[i].count / this.max,
          label: buckets[i].key,
          value: buckets[i].count,
          metadata: this.getImagePreview(imageUrl),
        });
      }
      return facetData;
    },
    getImagePreview(imageUrl: string) {
      const ip = new ImagePreview({
        store: this.$store,
        router: this.$router,
        propsData: {
          // NOTE: there seems to be an issue with the visibility plugin used
          // when injecting this way. Cancel the visibility flagging for facets.
          preventHiding: true,
          imageUrl,
          type: this.summary.varType,
        },
      });
      ip.$mount();
      return ip.$el;
    },
    viewMore() {
      this.moreNumToDisplay = viewMoreData(
        this.moreNumToDisplay,
        this.facetMoreCount,
        this.baseNumToDisplay,
        this.facetValueCount
      );
    },
    viewLess() {
      this.moreNumToDisplay = viewLessData(
        this.moreNumToDisplay,
        this.facetMoreCount,
        this.baseNumToDisplay,
        this.facetValueCount
      );
    },
    getHighlightValue(highlight: Highlight): any {
      if (highlight && highlight.value) {
        return highlight.value;
      }
      return null;
    },
    isHighlightedInstance(highlight: Highlight): boolean {
      return highlight && highlight.context === this.instanceName;
    },
    isHighlightedGroup(highlight: Highlight, colName: string): boolean {
      return this.isHighlightedInstance(highlight) && highlight.key === colName;
    },
    updateSelection(event) {
      if (!this.enableHighlighting) return;
      const facet = event.currentTarget;
      if (
        event.detail.changedProperties.get("selection") !== undefined &&
        !_.isEqual(facet.selection, this.selection)
      ) {
        let value = null;
        if (facet.selection) {
          if (this.selection) {
            const oldKey = Object.keys(this.selection)[0];
            const incomingKeys = Object.keys(facet.selection);
            const newKey = incomingKeys.filter(
              (iKey) => oldKey.indexOf(iKey) < 0
            )[0];
            value = this.facetData.values[newKey].label;
          } else {
            const newKey = Object.keys(facet.selection)[0];
            value = this.facetData.values[newKey].label;
          }
        }
        this.$emit(
          "facet-click",
          this.instanceName,
          this.summary.key,
          value,
          this.summary.dataset
        );
      }
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
              colName: this.summary.key,
            })
          : this.html;
      }
      return null;
    },
    getGroupIcon,
  },
});
</script>

<style>
.facet-image .facet-terms-container {
  max-height: 200px !important;
  overflow-y: auto;
  display: flex;
  flex-wrap: wrap;
}
</style>
<style scoped>
.collapse-unused {
  display: none;
}
.facet-image-label {
  max-width: 75px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.facet-content-container {
  display: inline-block;
  width: 85px;
  overflow: hidden;
}
.facet-image-preview-display {
  padding-left: 10px;
}

.facet-header-icon {
  margin-right: 6px;
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

.facet-footer-more-section {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  justify-content: flex-start;
  align-content: stretch;
  align-items: flex-start;
}

.facet-footer-more-count {
  order: 0;
  flex: 1 1 auto;
  align-self: auto;
}

.facet-footer-more-controls {
  order: 0;
  flex: 0 1 auto;
  align-self: auto;
}

.facet-footer-more-controls > span {
  cursor: pointer;
}

.facet-footer-custom-html {
  margin-top: 6px;
}

.facet-header-container {
  color: rgba(0, 0, 0, 0.54);
  display: flex;
  align-items: center;
  overflow-y: scroll !important;
}

.facet-header-container-no-scroll {
  color: rgba(0, 0, 0, 0.54);
  overflow: auto;
}

.facet-header-container .dropdown-menu {
  max-height: 200px;
  overflow-y: auto;
}
</style>
