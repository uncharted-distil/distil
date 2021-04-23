<!--

    Copyright © 2021 Uncharted Software Inc.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
-->

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
}
</style>
