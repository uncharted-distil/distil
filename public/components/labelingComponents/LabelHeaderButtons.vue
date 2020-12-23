<template>
  <div class="pt-2">
    <b-button @click="onButtonClick(positive)">
      <i class="fa fa-check text-success" aria-hidden="true"></i>
      Positive
    </b-button>
    <b-button @click="onButtonClick(negative)">
      <i class="fa fa-times red" aria-hidden="true"></i>
      Negative</b-button
    >
    <b-button @click="onButtonClick(unlabeled)">Unlabeled</b-button>
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
</style>
