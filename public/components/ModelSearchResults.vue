<template>
  <div class="model-search-results">
    <div class="row justify-content-center" v-if="isPending">
      <div v-html="spinnerHTML"></div>
    </div>
    <div class="model-search-results-container" ref="modelResults">
      <div class="mb-3" :key="model.fittedSolutionId" v-for="model in models">
        <model-preview :model="model"> </model-preview>
      </div>
      <div
        class="row justify-content-center pt-3"
        v-if="!isPending && (!models || models.length === 0)"
      >
        <h5>No models found for search</h5>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import $ from "jquery";
import ModelPreview from "../components/ModelPreview";
import Vue from "vue";
import { spinnerHTML } from "../util/spinner";
import { getters as modelGetters } from "../store/model/module";
import { Model } from "../store/model/index";

export default Vue.extend({
  name: "model-search-results",

  components: {
    ModelPreview
  },

  props: {
    isPending: Boolean as () => boolean
  },

  computed: {
    models(): Model[] {
      return modelGetters.getModels(this.$store);
    },
    spinnerHTML(): string {
      return spinnerHTML();
    }
  },

  watch: {
    models() {
      // reset back to top on model change
      const $results = this.$refs.modelResults as Element;
      $results.scrollTop = 0;
    }
  }
});
</script>

<style>
.model-search-results {
  width: 100%;
  overflow-x: hidden;
  overflow-y: auto;
}
.model-search-results-container {
  width: 100%;
}
</style>
