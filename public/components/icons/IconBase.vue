<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    class="svg-icon"
    :width="width"
    :height="height"
    viewBox="0 0 18 18"
    role="presentation"
  >
    <title v-if="title" lang="en">{{ title }}</title>
    <g :fill="iconColor">
      <slot />
    </g>
  </svg>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "icon-base",
  props: {
    title: {
      type: String,
      default: "",
    },
    width: {
      type: [Number, String],
      default: 18,
    },
    height: {
      type: [Number, String],
      default: 18,
    },
    iconColor: {
      type: String,
      default: "currentColor",
    },
  },
});
</script>

<style scoped>
svg {
  display: inline-block;
  vertical-align: baseline;
  margin-bottom: -2px; /* yes, I'm that particular about formatting */
}
</style>
