<template>
  <facet-terms
    :data.prop="facetData"
    action-buttons="0"
    :selection.prop="selection"
    :subselection.prop="subSelection"
    @facet-element-updated="updateSelection"
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
      >
      </type-change-menu>
    </div>
    <facet-template target="facet-terms-value">
      <div slot="header" class="facet-image-preview-display">
        ${metadata.getImagePreview(metadata.imageContext)}
      </div>
    </facet-template>
    <div slot="footer" class="facet-footer-container">
      <div v-if="facetDisplayMore" class="facet-footer-more">
        <div class="facet-footer-more-section">
          <div class="facet-footer-more-count">
            <span v-if="facetMoreCount > 0">{{ facetMoreCount }} more</span>
          </div>
          <div class="facet-footer-more-controls">
            <span v-if="hasMore" @click="viewMore"> show more</span>
            <span v-if="hasLess" @click="viewLess"> show less</span>
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

import "@uncharted/facets-core";
import { FacetTermsData } from "@uncharted/facets-core/dist/types/facet-terms/FacetTerms";

import TypeChangeMenu from "../TypeChangeMenu";
import ImagePreview from "../ImagePreview";
import { Highlight, RowSelection, VariableSummary } from "../../store/dataset";
import {
  getCategoricalChunkSize,
  getGroupIcon,
  getSubSelectionValues,
  hasBaseline
} from "../../util/facets";
import _ from "lodash";
import { IMAGE_TYPE } from "../../util/types";

export default Vue.extend({
  name: "facet-image",

  components: {
    TypeChangeMenu,
    ImagePreview
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
    enabledTypeChanges: Array as () => string[],
    html: [
      String as () => string,
      Object as () => any,
      Function as () => Function
    ],
    expandCollapse: Function as () => Function,
    highlight: Object as () => Highlight,
    enableHighlighting: Boolean as () => boolean,
    instanceName: String as () => string,
    rowSelection: Object as () => RowSelection
  },

  data() {
    return {
      baseNumToDisplay: getCategoricalChunkSize(this.summary.type),
      moreNumToDisplay: 0
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
        values
      };
    },
    facetEnableTypeChanges(): boolean {
      const key = `${this.summary.dataset}:${this.summary.key}`;
      return Boolean(this.enabledTypeChanges.find(e => e === key));
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
      return this.hasExamplars
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
    }
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
        facetData.push({
          ratio: buckets[i].count / this.max,
          label: buckets[i].key,
          value: buckets[i].count,
          metadata: {
            imageContext: {
              store: this.$store,
              router: this.$router,
              imageUrl: this.hasExamplars
                ? summary.baseline.exemplars[i]
                : buckets[i].key,
              type: summary.varType
            },
            getImagePreview: this.getImagePreview
          }
        });
      }
      return facetData;
    },
    getImagePreview(imageContext: { store; router; imageUrl; type }) {
      const ip = new ImagePreview({
        store: imageContext.store,
        router: imageContext.router,
        propsData: {
          // NOTE: there seems to be an issue with the visibility plugin used
          // when injecting this way. Cancel the visibility flagging for facets.
          preventHiding: true,
          imageUrl: imageContext.imageUrl,
          type: imageContext.type
        }
      });
      ip.$mount();
      return ip.$el;
    },
    viewMore() {
      this.moreNumToDisplay =
        this.facetMoreCount > this.baseNumToDisplay
          ? this.moreNumToDisplay + this.baseNumToDisplay
          : this.moreNumToDisplay +
            (this.facetValueCount % this.baseNumToDisplay);
    },
    viewLess() {
      this.moreNumToDisplay =
        this.facetMoreCount === 0
          ? this.moreNumToDisplay -
            (this.facetValueCount % this.baseNumToDisplay)
          : this.moreNumToDisplay - this.baseNumToDisplay;
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
              iKey => oldKey.indexOf(iKey) < 0
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
              colName: this.summary.key
            })
          : this.html;
      }
      return null;
    },
    getGroupIcon
  }
});
</script>

<style scoped>
.facet-image-preview-display {
  padding-left: 10px;
}
.facet-header-container {
  display: flex;
  align-items: center;
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
  overflow-y: scroll !important;
}

.facet-header-container-no-scroll {
  overflow: auto;
}

.facet-header-container .dropdown-menu {
  max-height: 200px;
  overflow-y: auto;
}
</style>
