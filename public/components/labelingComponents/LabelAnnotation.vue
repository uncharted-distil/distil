<template>
  <i :class="getLabel" aria-hidden="true"></i>
</template>

<script lang="ts">
import Vue from "vue";
import { LowShotLabels } from "../../util/data";
export default Vue.extend({
  name: "label-annotation",
  props: {
    item: Object as () => any,
  },
  computed: {
    getLabel(): string {
      switch (this.item.LowShotLabels) {
        case LowShotLabels.positive:
          return "fa fa-check text-success";
          break;
        case LowShotLabels.negative:
          return "fa fa-times red";
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
