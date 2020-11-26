<template>
  <i :class="getLabel" aria-hidden="true"></i>
</template>

<script lang="ts">
import Vue from "vue";
import { LowShotLabels, LOW_SHOT_LABEL_COLUMN_NAME } from "../../util/data";
export default Vue.extend({
  name: "label-annotation",
  props: {
    item: Object as () => any,
  },
  computed: {
    getLabel(): string {
      switch (this.item[LOW_SHOT_LABEL_COLUMN_NAME].value) {
        case LowShotLabels.positive:
          return "fa fa-check text-success p-1";
          break;
        case LowShotLabels.negative:
          return "fa fa-times red p-1";
          break;
        default:
          return "d-none";
          break;
      }
    },
  },
});
</script>

<style scoped>
.red {
  color: var(--red);
}
</style>
