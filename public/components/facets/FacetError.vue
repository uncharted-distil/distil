<template>
  <!-- facet-container is undocumented! shhhhh.... -->
  <facet-container>
    <div slot="header" class="facet-header">
      <i class="fa fa-exclamation-circle"></i>
      <span>{{ summary.label.toUpperCase() }}</span>
      <type-change-menu
        v-if="facetEnableTypeChanges"
        class="facet-header-dropdown"
        :dataset="summary.dataset"
        :field="summary.key"
      >
      </type-change-menu>
    </div>
    <div slot="content" class="facet-content">
      <span>{{ summary.err }}</span>
    </div>
  </facet-container>
</template>

<script lang="ts">
import Vue from "vue";

import "@uncharted.software/facets-core";
import { VariableSummary } from "../../store/dataset";

import TypeChangeMenu from "../TypeChangeMenu";

export default Vue.extend({
  name: "facet-error",

  components: {
    TypeChangeMenu
  },

  props: {
    summary: Object as () => VariableSummary,
    enabledTypeChanges: Array as () => string[]
  },

  computed: {
    facetEnableTypeChanges(): boolean {
      const key = `${this.summary.dataset}:${this.summary.key}`;
      return Boolean(this.enabledTypeChanges.find(e => e === key));
    }
  }
});
</script>

<style scoped>
.facet-header {
  height: 20px;
  color: #1a1b1c;
  font-family: "IBM Plex Sans", sans-serif;
  font-size: 14px;
  font-style: normal;
  font-weight: 600;
  line-height: 20px;
  padding: 6px 12px 5px;
  box-sizing: content-box;

  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.facet-header > i {
  color: red;
  padding-right: 6px;
}

.facet-header-dropdown {
  position: absolute;
  right: 12px;
  top: 5px;
}

.facet-content {
  padding: 4px 12px 25px;
}
</style>