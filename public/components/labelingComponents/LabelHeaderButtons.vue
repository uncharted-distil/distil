<template>
  <div class="pt-2">
    <b-button title="Select all items on page" @click="onSelectAll"
      >Select All</b-button
    >
    <b-button
      title="Annotate selected items to positive"
      @click="onButtonClick(positive)"
    >
      <span class="stacked-icons">
        <i class="fa fa-circle fa-stack-1x" />
        <i class="fa fa-plus-circle text-success fa-stack-1x" />
      </span>
      Positive
    </b-button>
    <b-button
      title="Annotate selected items to negative"
      @click="onButtonClick(negative)"
    >
      <span class="stacked-icons">
        <i class="fa fa-circle fa-stack-1x" />
        <i class="fa fa-minus-circle red fa-stack-1x" />
      </span>
      Negative</b-button
    >
    <b-button
      title="Annotate select items to negative"
      @click="onButtonClick(unlabeled)"
      >Unlabeled</b-button
    >
    <layer-selection v-if="isRemoteSensing" />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { LowShotLabels } from "../../util/data";
import LayerSelection from "../LayerSelection.vue";
import { getters as routeGetters } from "../../store/route/module";
export default Vue.extend({
  name: "label-header-buttons",
  components: {
    LayerSelection,
  },
  computed: {
    negative(): string {
      return LowShotLabels.negative;
    },
    positive(): string {
      return LowShotLabels.positive;
    },
    unlabeled(): string {
      return LowShotLabels.unlabeled;
    },
  },
  methods: {
    onButtonClick(event: string) {
      this.$emit("button-event", event);
    },
    onSelectAll() {
      this.$emit("select-all");
    },
    isRemoteSensing(): boolean {
      return routeGetters.isMultiBandImage(this.$store);
    },
  },
});
</script>

<style scoped>
.red {
  color: var(--red);
}
.stacked-icons {
  position: relative;
  display: inline-block;
  width: 2em;
  height: 1em;
  line-height: 1em;
  vertical-align: middle;
}
</style>
